package chainsaw

import (
	"context"
	"fmt"
	"io"
	"os"
)

func MakeLogger(logBufferSize, chanBufferSize int) *CircularLogger {
	c := CircularLogger{
		printLevel:     InfoLevel, // this is the default printlevel.
		messages:       make([]LogMessage, logBufferSize),
		logCh:          make(logChan, chanBufferSize),
		outputChs:      make([]logChan, 0),
		controlCh:      make(controlChannel, 10), // Control buffer must be buffered.
		current:        0,
		logBufferSize:  logBufferSize,
		chanBufferSize: chanBufferSize,
		outputWriters:  []io.Writer{os.Stdout},
	}
	go c.channelHandler()
	return &c
}

// Stop
// goroutine which handles the log channel
// Things will deadlock if you log while it is down.
func (l *CircularLogger) Stop() {
	if l.running.Get() {
		cMsg := controlMessage{
			cType: crtlQuit,
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
		cType: crtlRst,
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
	ch := make(chan LogMessage, l.chanBufferSize)
	cMessage := controlMessage{
		cType:      ctrlAddOutputChan,
		returnChan: nil,
		outputCh:   ch,
	}
	l.controlCh <- cMessage
	go func(outputCh chan LogMessage) {
		<-ctx.Done()
		cMessage := controlMessage{
			cType:      ctrlRemoveOutputChan,
			returnChan: nil,
			outputCh:   outputCh,
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
	retCh := make(chan []LogMessage, l.chanBufferSize)
	cMsg := controlMessage{
		cType:      ctrlDump,
		returnChan: retCh,
		level:      level,
		outputCh:   nil,
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
	cMsg := controlMessage{
		cType: ctrlSetLogLevel,
		level: level,
	}
	l.controlCh <- cMsg
}

func (l *CircularLogger) AddOutput(o io.Writer) {
	cMsg := controlMessage{
		cType:     ctrlAddWriter,
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
		cType:     ctrlRemoveWriter,
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
	return l.running.Get()
}
