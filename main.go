package chainsaw

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

var defaultLogger *CircularLogger

type LogLevel int

var bufferSize = 150

const chanBufferSize = 10

const (
	Any LogLevel = iota
	TraceLevel
	DebugLevel
	InfoLevel
	WarnLevel
	ErrLevel
	FatalLevel
	Never
	ErrorLevel   = ErrLevel
	WarningLevel = WarnLevel
)

var Levels = []string{"Any", "trace", "debug", "info", "warn", "error", "fatal", "never"}

func (l LogLevel) String() string {
	return Levels[l]
	//return [...]string{"Any", "trace", "debug", "info", "warn", "error", "fatal", "never"}[l]
}

type controlType int

// controlMessage struct describes the control message sent to the worker. This is a multi-purpose struct that
// allows you to do a number of operations.
// - Add or remove new output channels
// - Dump messages
// - Stop the worker goroutine.
type controlMessage struct {
	cType      controlType
	returnChan chan []LogMessage // channel for dumps of messages
	level      LogLevel
	output     chan LogMessage
}

type controlChannel chan controlMessage

const (
	controlDump controlType = iota + 1
	controlAddOutput
	controlRemoveOutput
	controlReset
	controlQuit
)

type logChan chan LogMessage

type LogMessage struct {
	content   string
	loglevel  LogLevel
	timeStamp time.Time
}

type CircularLogger struct {
	// at what log level should messages be printed to stdout
	printLevel LogLevel
	// messages is the internal log buffer keeping the last messages in a circular buffer
	messages []LogMessage
	// logChan Messages are send over this channel to the log worker.
	logCh logChan
	// outputs is a list of channels that messages are copied onto when they arrive
	// if a stream has been requested it is added here.
	outputs []logChan
	// controlCh is the internal channel that is used for control messages.
	controlCh    controlChannel
	current      int
	len          int
	outputWriter io.Writer
	running      bool
}

func (l *CircularLogger) log(level LogLevel, m string) {
	t := time.Now()
	tStr := t.Format("2006-01-02T15:04:05-0700")
	if level >= l.printLevel {
		str := fmt.Sprintf("%s: [%s] %s\n", tStr, level.String(), m)
		_, err := io.WriteString(l.outputWriter, str)
		if err != nil {
			fmt.Printf("Internal error in chainsaw: Can't write to outputWriter: %s", err)
		}
	}
	logm := LogMessage{
		content:   m,
		loglevel:  level,
		timeStamp: t,
	}
	l.logCh <- logm
}

// channelHandler run whenever the logger has been used.
// this is the main goroutine where everything happens
func (l *CircularLogger) channelHandler() {
	l.running = true
	for {
		// fmt.Println("Waiting for log or control message.")
		select {
		case cMessage := <-l.controlCh:
			switch cMessage.cType {
			case controlAddOutput:
				l.addOutputChan(cMessage.output)
			case controlRemoveOutput:
				l.removeOutputChan(cMessage.output)
			case controlDump:
				buf := l.getMessageOverCh(cMessage.level)
				cMessage.returnChan <- buf
			case controlReset:
				l.current = 0
				l.messages = make([]LogMessage, bufferSize)
			case controlQuit:
				l.running = false
				return // ends the goroutine.
			}
		case msg := <-l.logCh:
			l.messages[l.current] = msg
			l.current = (l.current + 1) % l.len
			for _, ch := range l.outputs {
				ch <- msg
			}
		}
	}
}
func (l *CircularLogger) addOutputChan(ch logChan) {
	l.outputs = append(l.outputs, ch)
}

func (l *CircularLogger) removeOutputChan(ch logChan) {
	newoutput := make([]logChan, 0)
	for _, outputCh := range l.outputs {
		if outputCh != ch {
			newoutput = append(newoutput, outputCh)
		}
	}
	l.outputs = newoutput
}

func mkCircularLogger() *CircularLogger {
	c := &CircularLogger{
		printLevel:   TraceLevel,
		messages:     make([]LogMessage, bufferSize),
		logCh:        make(logChan, chanBufferSize),
		outputs:      make([]logChan, 0),
		controlCh:    make(controlChannel, chanBufferSize),
		current:      0,
		len:          bufferSize,
		outputWriter: os.Stdout,
	}
	go c.channelHandler()
	return c

}

// Stop stops the internal goroutine which handles the log channel
// Things will deadlock if you log while it is down.
func (l *CircularLogger) Stop() {
	cMsg := controlMessage{
		cType: controlQuit,
	}
	l.controlCh <- cMsg // We don't expect a reply.
}

func (l *CircularLogger) Reset() {
	cMsg := controlMessage{
		cType: controlReset,
	}
	l.controlCh <- cMsg // We don't expect a reply.
}

func MakeLogger() *CircularLogger {
	return mkCircularLogger()
}

// GetStream returns a channel of log messages.
// Log messages will be streamed on this channel. The channel MUST
// be serviced or the logger will lock up.
func (l *CircularLogger) GetStream(ctx context.Context) chan LogMessage {
	ch := make(chan LogMessage, chanBufferSize)
	cMessage := controlMessage{
		cType:      controlAddOutput,
		returnChan: nil,
		output:     ch,
	}
	l.controlCh <- cMessage
	go func(outputCh chan LogMessage) {
		<-ctx.Done()
		cMessage := controlMessage{
			cType:      controlRemoveOutput,
			returnChan: nil,
			output:     outputCh,
		}
		l.controlCh <- cMessage
	}(ch)
	return ch
}

func (l *CircularLogger) getMessageOverCh(level LogLevel) []LogMessage {
	buf := make([]LogMessage, 0)
	for i := l.current; i < l.current+l.len; i++ {
		msg := l.messages[i%l.len]
		if msg.loglevel < level {
			continue
		}
		buf = append(buf, msg)
	}
	return buf
}

func (l *CircularLogger) GetMessages(level LogLevel) []LogMessage {
	retCh := make(chan []LogMessage, chanBufferSize)
	cMsg := controlMessage{
		cType:      controlDump,
		returnChan: retCh,
		level:      level,
		output:     nil,
	}
	fmt.Println("Requesting messages over control channel")
	l.controlCh <- cMsg
	fmt.Println("Waiting for reply")
	ret := <-retCh
	return ret
}

func SetLevel(level LogLevel) {
	defaultLogger.printLevel = level
}

func (l *CircularLogger) SetOutput(o io.Writer) {
	l.outputWriter = o
}

// GetStatus is somewhat unsafe. There is no locking here, and we peek into the internals.
func (l *CircularLogger) GetStatus() bool {
	return l.running
}
