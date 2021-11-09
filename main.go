package chainsaw

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

//go:generate go run gen/main.go

var defaultLogger *CircularLogger

type LogLevel int

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
	defaultLogger = MakeLogger(30, 10)
}

var Levels = []string{"Any", "trace", "debug", "info", "warn", "error", "fatal", "never"}

func (l LogLevel) String() string {
	return Levels[l]
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
	outputCh   chan LogMessage
	newWriter  io.Writer
}

type controlChannel chan controlMessage

const (
	ctrlDump controlType = iota + 1
	crtlRst
	crtlQuit
	ctrlAddOutputChan
	ctrlRemoveOutputChan
	ctrlAddWriter
	ctrlRemoveWriter
	ctrlSetLogLevel
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
	// outputChs is a list of channels that messages are copied onto when they arrive
	// if a stream has been requested it is added here.
	outputChs []logChan
	// controlCh is the internal channel that is used for control messages.
	controlCh      controlChannel
	current        int
	logBufferSize  int
	outputWriters  []io.Writer
	running        atomicBool
	chanBufferSize int
}

func (l *CircularLogger) log(level LogLevel, m string) {
	t := time.Now()
	logM := LogMessage{
		Content:   m,
		LogLevel:  level,
		TimeStamp: t,
	}
	l.logCh <- logM
}

// channelHandler run whenever the logger has been used.
// this is the main goroutine where everything happens
func (l *CircularLogger) channelHandler() {
	l.running.Set(true)
	for {
		// fmt.Println("Waiting for log or control message.")
		select {
		case cMessage := <-l.controlCh:
			switch cMessage.cType {
			case ctrlAddOutputChan:
				l.addOutputChan(cMessage.outputCh)
			case ctrlRemoveOutputChan:
				l.removeOutputChan(cMessage.outputCh)
			case ctrlDump:
				buf := l.getMessageOverCh(cMessage.level)
				cMessage.returnChan <- buf
			case crtlRst:
				l.current = 0
				l.messages = make([]LogMessage, l.logBufferSize)
			case crtlQuit:
				l.running.Set(false)
				return // ends the goroutine.
			case ctrlAddWriter:
				fmt.Println("Adding writer")
				l.addWriter(cMessage.newWriter)
			case ctrlRemoveWriter:
				l.removeWriter(cMessage.newWriter)
			case ctrlSetLogLevel:
				l.SetLevel(cMessage.level)
			default:
				panic("unknown control message")
			}
		case msg := <-l.logCh:
			l.handleLogOutput(msg)
			l.messages[l.current] = msg
			l.current = (l.current + 1) % l.logBufferSize
			for _, ch := range l.outputChs {
				ch <- msg
			}
		}
	}
}

func (l *CircularLogger) handleLogOutput(m LogMessage) {
	tStr := m.TimeStamp.Format("2006-01-02T15:04:05-0700")
	if m.LogLevel >= l.printLevel {
		str := fmt.Sprintf("%s: [%s] %s\n", tStr, m.LogLevel.String(), m.Content)
		for _, output := range l.outputWriters {
			_, err := io.WriteString(output, str) // Should we check for a short write?
			if err != nil {
				fmt.Printf("Internal error in chainsaw: Can't write to outputWriter: %s", err)
			}
		}
	}

}

func (l *CircularLogger) addOutputChan(ch logChan) {
	l.outputChs = append(l.outputChs, ch)
}

func (l *CircularLogger) removeOutputChan(ch logChan) {
	for i, outputCh := range l.outputChs {
		if outputCh == ch {
			l.outputChs[i] = l.outputChs[len(l.outputChs)-1]
			l.outputChs = l.outputChs[:len(l.outputChs)-1]
			close(ch)
		}
	}
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
	for i := l.current; i < l.current+l.logBufferSize; i++ {
		msg := l.messages[i%l.logBufferSize]
		if msg.LogLevel < level {
			continue
		}
		buf = append(buf, msg)
	}
	return buf
}

func (l *CircularLogger) setLevel(level LogLevel) {
	l.printLevel = level
}

type atomicBool struct{ flag int32 }

func (b *atomicBool) Set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}
	atomic.StoreInt32(&(b.flag), int32(i))
}

func (b *atomicBool) Get() bool {
	if atomic.LoadInt32(&(b.flag)) != 0 {
		return true
	}
	return false
}
