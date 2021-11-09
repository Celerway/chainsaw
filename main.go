package chainsaw

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

//go:generate go run gen/main.go

var defaultLogger *CircularLogger

// LogLevel used to set the level at which output is printed to the Writer.
type LogLevel int

const (
	// Any is not used outside of testing.
	Any LogLevel = iota
	// TraceLevel is the lowest log level of logging for very detailed logging when debugging.
	TraceLevel
	// DebugLevel is typically used when debugging
	DebugLevel
	// InfoLevel is used for informational messages that are important but expected
	InfoLevel
	// WarnLevel is used for warnings.
	WarnLevel
	// ErrLevel is used for non-fatal errors
	ErrLevel
	// FatalLevel is for fatal errors. Not really useful as a log level.
	FatalLevel
	// Never is not used ever.
	Never
	ErrorLevel   = ErrLevel
	WarningLevel = WarnLevel
)

func init() {
	defaultLogger = MakeLogger(30, 10)
}

var Levels = []string{"any", "trace", "debug", "info", "warn", "error", "fatal", "never"}

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

// LogMessage contains a single log line. Timestamp is added by Chainsaw, the rest comes
// from the user.
type LogMessage struct {
	Content   string
	LogLevel  LogLevel
	TimeStamp time.Time
}

// CircularLogger is the struct holding a chainsaw instance.
// All state within is private and should be access through methods.
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
	if l.GetStatus() == false {
		// If the logger is down we will deadlock quickly and nobody wants that.
		panic("chainsaw panic: Logging goroutine is stopped")
	}
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
func (l *CircularLogger) channelHandler(wg *sync.WaitGroup) {
	l.running.Set(true)
	wg.Done()
	for {
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
				l.addWriter(cMessage.newWriter)
			case ctrlRemoveWriter:
				l.removeWriter(cMessage.newWriter)
			case ctrlSetLogLevel:
				l.setLevel(cMessage.level)
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
	return atomic.LoadInt32(&(b.flag)) != 0
}
