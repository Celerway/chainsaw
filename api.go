package chainsaw

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
)

// MakeLogger creates a new logger instance. Params:
// logBufferSize - the size of the circular buffer
// chanBufferSize - how big the channels buffers should be
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
	wg := sync.WaitGroup{} // Waits for the goroutine to start.
	wg.Add(1)
	go c.channelHandler(&wg)
	wg.Wait()
	return &c
}

// Stop the goroutine which handles the log channel.
// Things might deadlock if you log while it is down.
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

// Reset the circular buffer of the logger. Flush the logs.
func (l *CircularLogger) Reset() {
	cMsg := controlMessage{
		cType: crtlRst,
	}
	l.controlCh <- cMsg // We don't expect a reply.
}

// Reset the default loggers buffer.
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

// GetStream creates a stream (channel) from the default logger. The channel MUST be serviced
// or the logger will lock up.
func GetStream(ctx context.Context) chan LogMessage {
	l := defaultLogger
	return l.GetStream(ctx)
}

// GetMessages fetches the messages currently in the circular buffer.
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

// GetMessages fetches the messages currently in the circular buffer.
func GetMessages(level LogLevel) []LogMessage {
	l := defaultLogger
	return l.GetMessages(level)
}

// SetLevel sets the log level. This affects if messages are printed to
// the outputs or not.
func SetLevel(level LogLevel) {
	l := defaultLogger
	l.SetLevel(level)
}

// SetLevel sets the log level. This affects if messages are printed to
// the outputs or not.
func (l *CircularLogger) SetLevel(level LogLevel) {
	cMsg := controlMessage{
		cType: ctrlSetLogLevel,
		level: level,
	}
	l.controlCh <- cMsg
}

// AddOutput takes a io.Writer and will copy log messages here going forward
// Note that it does not check if the Writer is already here. Copying os.Stdout
// will result in all messages being printed twice.
func (l *CircularLogger) AddOutput(o io.Writer) {
	cMsg := controlMessage{
		cType:     ctrlAddWriter,
		newWriter: o,
	}
	l.controlCh <- cMsg // Requesting messages over control channel
}

// AddOutput takes a io.Writer and will copy log messages here going forward
// Note that it does not check if the Writer is already here. Copying os.Stdout
// will result in all messages being printed twice.
func AddOutput(o io.Writer) {
	l := defaultLogger
	l.AddOutput(o)
}

// RemoveWriter removes the io.Writer from the logger.
// if the Writer isn't there nothing happens.
func (l *CircularLogger) RemoveWriter(o io.Writer) {
	cMsg := controlMessage{
		cType:     ctrlRemoveWriter,
		newWriter: o,
	}
	l.controlCh <- cMsg
}

// RemoveWriter removes the io.Writer from the logger.
// if the Writer isn't there nothing happens.
func RemoveWriter(o io.Writer) {
	l := defaultLogger
	l.RemoveWriter(o)
}

// GetStatus returns true if the logger goroutine is running.
func (l *CircularLogger) GetStatus() bool {
	return l.running.Get()
}
