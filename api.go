package chainsaw

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	defaultLogBufferSize  = 50
	defaultChanBufferSize = 10
)

// MakeLogger creates a new logger instance.
// Params:
// Mandatory is name, can be empty.
// In addition you can supply two ints:
// logBufferSize - the size of the circular buffer
// chanBufferSize - how big the channels buffers should be
func MakeLogger(name string, options ...int) *CircularLogger {
	logBufferSize := defaultLogBufferSize
	chanBufferSize := defaultChanBufferSize
	if len(options) > 1 {
		logBufferSize = options[0]
	}
	if len(options) > 2 {
		chanBufferSize = options[1]
	}
	c := CircularLogger{
		name:           name,
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
		_ = l.sendCtrlAndWait(cMsg)
	} else {
		fmt.Printf("Error! Stop called on a passive logger")
	}
}

// Reset the circular buffer of the logger. Flush the logs.
func (l *CircularLogger) Reset() {
	cMsg := controlMessage{
		cType: crtlRst,
	}
	_ = l.sendCtrlAndWait(cMsg)
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
	/*
		There is a race condition here. What happens is:
		ctx is cancelled.
		We send the control message then we get blocked --> deadlock.

	*/
	// Make the channel we're gonna return.
	retCh := make(chan LogMessage, l.chanBufferSize)
	cMessage := controlMessage{cType: ctrlAddOutputChan, outputCh: retCh}
	_ = l.sendCtrlAndWait(cMessage)
	go func(outputCh chan LogMessage) {
		<-ctx.Done()
		cMessage := controlMessage{cType: ctrlRemoveOutputChan, outputCh: outputCh}
		err := l.sendCtrlAndWait(cMessage) // waits for the response.
		if err != nil {
			fmt.Println("chainsaw internal error in GetStream():", err.Error())
		}
	}(retCh)
	return retCh
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
	_ = l.sendCtrlAndWait(cMsg)
}

// AddWriter takes a io.Writer and will copy log messages here going forward
// If it already exists an error is returned.
func (l *CircularLogger) AddWriter(o io.Writer) error {
	cMsg := controlMessage{
		cType:     ctrlAddWriter,
		newWriter: o,
	}
	return l.sendCtrlAndWait(cMsg)
}

// AddWriter takes a io.Writer and will copy log messages here going forward
// If it already exists an error is returned.
func AddWriter(o io.Writer) error {
	l := defaultLogger
	return l.AddWriter(o)
}

// RemoveWriter removes the io.Writer from the logger.
// if the Writer isn't there an error is returned
func (l *CircularLogger) RemoveWriter(o io.Writer) error {
	cMsg := controlMessage{
		cType:     ctrlRemoveWriter,
		newWriter: o,
	}
	return l.sendCtrlAndWait(cMsg)
}

// RemoveWriter removes the io.Writer from the logger.
// if the Writer isn't there an error is returned
func RemoveWriter(o io.Writer) error {
	l := defaultLogger
	return l.RemoveWriter(o)
}

// GetStatus returns true if the logger goroutine is running.
func (l *CircularLogger) GetStatus() bool {
	return l.running.Get()
}

// Flush sends a no-op message and waits for the reply. This flushes the
// buffers.
func (l *CircularLogger) Flush() error {
	cMsg := controlMessage{cType: ctrlFlush}
	return l.sendCtrlAndWait(cMsg)
}

func (m LogMessage) String() string {
	tStr := m.TimeStamp.Format("2006-01-02T15:04:05-0700")
	return fmt.Sprintf("%s: [%s] %s\n", tStr, m.LogLevel.String(), m.Content)
}
