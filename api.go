package chainsaw

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

const (
	defaultLogBufferSize  = 50
	defaultChanBufferSize = 10
)

// MakeLogger creates a new logger instance.
// Params:
// Mandatory is name, can be empty.
// In addition, you can supply two ints:
// logBufferSize - the size of the circular buffer
// chanBufferSize - how big the channels buffers should be
// Note that the log level is fetched from the default logger.
func MakeLogger(name string, options ...int) *CircularLogger {
	logBufferSize := defaultLogBufferSize
	chanBufferSize := defaultChanBufferSize
	if len(options) > 0 {
		logBufferSize = options[0]
	}
	if len(options) > 1 {
		chanBufferSize = options[1]
	}

	var printLevel LogLevel
	if defaultLogger != nil {
		printLevel = defaultLogger.printLevel
	} else {
		printLevel = InfoLevel
	}

	c := CircularLogger{
		name:           name,
		printLevel:     printLevel,
		messages:       make([]LogMessage, logBufferSize),
		logCh:          make(logChan, chanBufferSize),
		outputChs:      make([]logChan, 0),
		controlCh:      make(controlChannel, 10), // Control buffer must be buffered.
		current:        0,
		logBufferSize:  logBufferSize,
		chanBufferSize: chanBufferSize,
		outputWriters:  []io.Writer{os.Stderr},
		TimeFmt:        "2006-01-02T15:04:05-0700",
		backTraceLevel: TraceLevel,
	}
	go c.channelHandler()
	runtime.Gosched()
	return &c
}

// Stop the goroutine which handles the log channel.
// Things might deadlock if you log while it is down.
func (l *CircularLogger) Stop() {
	cMsg := controlMessage{cType: ctrlQuit}
	_ = l.sendCtrlAndWait(cMsg)
}

// Reset the circular buffer of the logger. Flush the logs.
func (l *CircularLogger) Reset() {
	cMsg := controlMessage{cType: ctrlRst}
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
	// Make the channel we're gonna return.
	retCh := make(chan LogMessage, l.chanBufferSize)
	cMessage := controlMessage{cType: ctrlAddOutputChan, outputCh: retCh}
	_ = l.sendCtrlAndWait(cMessage)
	// spin of a goroutine that will wait for the context.
	go func(outputCh chan LogMessage) {
		<-ctx.Done() // wait for the context to cancel
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

func SetBackTraceLevel(level LogLevel) {
	l := defaultLogger
	l.SetBackTraceLevel(level)
}

func (l *CircularLogger) SetBackTraceLevel(level LogLevel) {
	l.backTraceLevel = level
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

// Flush sends a no-op message and waits for the reply. This flushes the
// buffers.
func (l *CircularLogger) Flush() error {
	cMsg := controlMessage{cType: ctrlFlush}
	return l.sendCtrlAndWait(cMsg)
}

func GetLevels() []LogLevel {
	levels := make([]LogLevel, 0)
	for l := TraceLevel; l <= FatalLevel; l++ {
		levels = append(levels, l)
	}
	return levels
}

func ParseLogLevel(s string) (LogLevel, error) {
	s = strings.ToLower(s)
	s = strings.TrimSuffix(s, "level")
	for l := TraceLevel; l <= FatalLevel; l++ {
		s2 := strings.ToLower(l.String())
		s2 = strings.TrimSuffix(s2, "level")
		if s2 == s {
			return l, nil
		}
	}
	return 0, fmt.Errorf("invalid log level: %s", s)
}
