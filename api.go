package chainsaw

import (
	"context"
	"fmt"
	"io"
	"os"
)

func MakeLogger() *CircularLogger {
	c := CircularLogger{
		printLevel:    TraceLevel,
		messages:      make([]LogMessage, bufferSize),
		logCh:         make(logChan, chanBufferSize),
		outputs:       make([]logChan, 0),
		controlCh:     make(controlChannel, chanBufferSize),
		current:       0,
		len:           bufferSize,
		outputWriters: make([]io.Writer, 0),
	}
	c.AddOutput(os.Stdout)
	go c.channelHandler()
	return &c
}

// Stop
// goroutine which handles the log channel
// Things will deadlock if you log while it is down.
func (l *CircularLogger) Stop() {
	if l.running {
		cMsg := controlMessage{
			cType: controlQuit,
		}
		l.controlCh <- cMsg // We don't expect a reply.
	} else {
		fmt.Printf("Error! Stop called on a passive logger")
	}
}

func Stop() {
	l := defaultLogger
	l.Stop()
}

func (l *CircularLogger) Reset() {
	cMsg := controlMessage{
		cType: controlReset,
	}
	l.controlCh <- cMsg // We don't expect a reply.
}

func Reset() {
	l := defaultLogger
	l.Reset()
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
func GetStream(ctx context.Context) chan LogMessage {
	l := defaultLogger
	return l.GetStream(ctx)
}

func (l *CircularLogger) GetMessages(level LogLevel) []LogMessage {
	retCh := make(chan []LogMessage, chanBufferSize)
	cMsg := controlMessage{
		cType:      controlDump,
		returnChan: retCh,
		level:      level,
		output:     nil,
	}
	l.controlCh <- cMsg // Requesting messages over control channel
	ret := <-retCh
	return ret
}

func GetMessages(level LogLevel) []LogMessage {
	l := defaultLogger
	return l.GetMessages(level)
}

func SetLevel(level LogLevel) {
	l := defaultLogger
	l.SetLevel(level)
}

func (l *CircularLogger) SetLevel(level LogLevel) {
	l.printLevel = level
}

func (l *CircularLogger) AddOutput(o io.Writer) {
	cMsg := controlMessage{
		cType:     controlAddWriter,
		newWriter: o,
	}
	l.controlCh <- cMsg // Requesting messages over control channel
}

func AddOutput(o io.Writer) {
	l := defaultLogger
	l.AddOutput(o)
}

func (l *CircularLogger) RemoveWriter(o io.Writer) {
	cMsg := controlMessage{
		cType:     controlRemoveWriter,
		newWriter: o,
	}
	l.controlCh <- cMsg
}

func RemoveWriter(o io.Writer) {
	l := defaultLogger
	l.RemoveWriter(o)
}

// GetStatus is somewhat unsafe. There is no locking here, and we peek into the internals.
func (l *CircularLogger) GetStatus() bool {
	return l.running
}
