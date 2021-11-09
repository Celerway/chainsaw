package chainsaw

import (
	"fmt"
	"io"
	"time"
)

//go:generate go run gen/main.go

var defaultLogger *CircularLogger

type LogLevel int

var bufferSize = 150

const chanBufferSize = 30

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

func init() {
	defaultLogger = MakeLogger()
}

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
	newWriter  io.Writer
}

type controlChannel chan controlMessage

const (
	controlDump controlType = iota + 1
	controlAddOutput
	controlRemoveOutput
	controlReset
	controlQuit
	controlAddWriter
	controlRemoveWriter
)

type logChan chan LogMessage

type LogMessage struct {
	Content   string
	LogLevel  LogLevel
	TimeStamp time.Time
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
	controlCh     controlChannel
	current       int
	len           int
	outputWriters []io.Writer
	running       bool
}

func (l *CircularLogger) log(level LogLevel, m string) {
	t := time.Now()
	tStr := t.Format("2006-01-02T15:04:05-0700")
	if level >= l.printLevel {
		str := fmt.Sprintf("%s: [%s] %s\n", tStr, level.String(), m)
		for _, output := range l.outputWriters {
			_, err := io.WriteString(output, str)
			if err != nil {
				fmt.Printf("Internal error in chainsaw: Can't write to outputWriter: %s", err)
			}
		}
	}
	logm := LogMessage{
		Content:   m,
		LogLevel:  level,
		TimeStamp: t,
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
			case controlAddWriter:
				l.addWriter(cMessage.newWriter)
			case controlRemoveWriter:
				l.removeWriter(cMessage.newWriter)
			default:
				panic("unknown control message")
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

func (l *CircularLogger) addWriter(o io.Writer) {
	l.outputWriters = append(l.outputWriters, o)
}

func (l *CircularLogger) removeWriter(o io.Writer) {
	for i, wr := range l.outputWriters {
		if o == wr {
			l.outputWriters[i] = l.outputWriters[len(l.outputWriters)-1]
			l.outputWriters = l.outputWriters[:len(l.outputWriters)-1]
		}
	}
}

func (l *CircularLogger) getMessageOverCh(level LogLevel) []LogMessage {
	buf := make([]LogMessage, 0)
	for i := l.current; i < l.current+l.len; i++ {
		msg := l.messages[i%l.len]
		if msg.LogLevel < level {
			continue
		}
		buf = append(buf, msg)
	}
	return buf
}
