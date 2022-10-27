package chainsaw

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Code generation; the order matters, make sure to run the stringers first.
//
//go:generate stringer -type=controlType
//go:generate stringer -type=LogLevel
//go:generate go run gen/main.go
const channelTimeout = time.Second

var defaultLogger *CircularLogger

// LogLevel used to set the level at which output is printed to the Writer.
type LogLevel int

const (
	// TraceLevel is the lowest log level of logging for very detailed logging when debugging.
	TraceLevel LogLevel = iota + 1
	// DebugLevel is typically used when debugging
	DebugLevel
	// InfoLevel is used for informational messages that are important but expected
	InfoLevel
	// WarnLevel is used for warnings.
	WarnLevel
	// ErrorLevel is used for non-fatal errors
	ErrorLevel
	// FatalLevel is for fatal errors. Not really useful as a log level.
	FatalLevel
)

// init() runs when the library is imported. It will start a logger
// so the library can be used without initializing a custom logger.
func init() {
	defaultLogger = MakeLogger("", 30, 10)
}

// var Levels = []string{"any", "trace", "debug", "info", "warn", "error", "fatal", "never"}

type controlType int

// controlMessage struct describes the control message sent to the worker. This is a multipurpose struct that
// allows you to do a number of operations.
// - Add or remove new output channels
// - Dump messages
// - Stop the worker goroutine.
type controlMessage struct {
	cType      controlType
	returnChan chan []LogMessage // channel for dumps of messages
	errChan    chan error        // Used for acknowledgements
	level      LogLevel
	outputCh   chan LogMessage
	newWriter  io.Writer
}

type controlChannel chan controlMessage

const (
	// ctrlDump dump messages in the buffer
	ctrlDump controlType = iota + 1
	// ctrlRst resets the buffer, discarding any messages
	ctrlRst
	// ctrlQuit stops the worker goroutine
	ctrlQuit
	// ctrlAddOutputChan adds a new output channel
	ctrlAddOutputChan
	// ctrlRemoveOutputChan removes an output channel
	ctrlRemoveOutputChan
	// ctrlAddWriter adds a new io.Writer to the list of outputs
	ctrlAddWriter
	// ctrlRemoveWriter removes an io.Writer from the list of outputs
	ctrlRemoveWriter
	// ctrlSetLogLevel sets the log level
	ctrlSetLogLevel
	// ctrlFlush flushes the buffer to the outputs. also used to synchronize.
	ctrlFlush
)

type logChan chan LogMessage

// LogMessage contains a single log line. Timestamp is added by Chainsaw, the rest comes
// from the user.
type LogMessage struct {
	Message   string
	LogLevel  LogLevel
	TimeStamp time.Time
	Fields    string
}

// CircularLogger is the struct holding a chainsaw instance.
// All state within is private and should be access through methods.
type CircularLogger struct {
	name       string   // The name gets  into all the log-lines when printed.
	printLevel LogLevel // at what log level should messages be printed to stdout
	// messages is the internal log buffer keeping the last messages in a circular buffer
	messages []LogMessage
	logCh    logChan // logChan Messages are sent over this channel to the log worker.
	// outputChs is a list of channels that messages are copied onto when they arrive
	// if a stream has been requested it is added here.
	outputChs []logChan
	// controlCh is the internal channel that is used for control messages.
	controlCh      controlChannel
	current        int         // points to the current message in the circular buffer
	logBufferSize  int         // the size of the circular buffer we use.
	outputWriters  []io.Writer // List of io.Writers where the output gets copied.
	chanBufferSize int         // how big the channel buffer is. re-used when making streams.
	TimeFmt        string      // format string for date. Default is "2006-01-02T15:04:05-0700"
	fields         string      // pre-formatted set of fields. this gets included in every line.
	backTraceLevel LogLevel    // at what level should we trigger a backtrace? Default is TraceLevel, which disables this.
}

func (l *CircularLogger) log(level LogLevel, message string, fields string) {
	t := time.Now()
	msgFields := l.fields
	if len(fields) > 0 {
		msgFields += " " + fields
	}
	logM := LogMessage{
		Message:   message,
		Fields:    msgFields,
		LogLevel:  level,
		TimeStamp: t,
	}
	l.logCh <- logM
	// trigger backtrace if level is high enough to trigger it.
	// if backTraceLevel is set to TraceLevel, this is disabled.
	if l.backTraceLevel > TraceLevel && level >= l.backTraceLevel {
		err := l.Flush()
		if err != nil {
			fmt.Printf("Error flushing: %s", err)
		}
		err = l.backTrace()
		if err != nil {
			fmt.Println("Chainsaw internal error getting backtrace: ", err)
		}
	}
}

// channelHandler run whenever the logger has been used.
// this is the main goroutine where everything happens
func (l *CircularLogger) channelHandler() {
	for {
		select {
		case cMessage := <-l.controlCh:
			switch cMessage.cType {
			case ctrlAddOutputChan:
				cMessage.errChan <- l.addOutputChan(cMessage.outputCh)
			case ctrlRemoveOutputChan:
				cMessage.errChan <- l.removeOutputChan(cMessage.outputCh)
			case ctrlDump:
				cMessage.returnChan <- l.getMessageOverCh(cMessage.level)
			case ctrlRst:
				cMessage.errChan <- l.handleReset()
			case ctrlQuit:
				cMessage.errChan <- nil
				return // ends the goroutine.
			case ctrlAddWriter:
				cMessage.errChan <- l.addWriter(cMessage.newWriter)
			case ctrlRemoveWriter:
				cMessage.errChan <- l.removeWriter(cMessage.newWriter)
			case ctrlSetLogLevel:
				cMessage.errChan <- l.setLevel(cMessage.level)
			case ctrlFlush:
				cMessage.errChan <- nil // Flush is always a success
			default:
				panic("unknown control message")
			}
		case msg := <-l.logCh:
			l.handleLogOutput(msg)
			l.messages[l.current] = msg
			l.current = (l.current + 1) % l.logBufferSize
			for _, ch := range l.outputChs {
				select {
				case ch <- msg:
				case <-time.After(channelTimeout):
					fmt.Println("timed out during channel write, removing channel")
					err := l.removeOutputChan(ch) // This should be safe. The channel is buffered.
					if err != nil {
						fmt.Println("error while removing block channel: ", err)
					}
				}
			}
			if msg.LogLevel == FatalLevel {
				fmt.Println("[chainsaw causing exit]")
				os.Exit(1)
			}

		}
	}
}

func (l *CircularLogger) formatMessage(m LogMessage) string {
	output := make([]string, 0)
	// add time.
	output = append(output,
		l.formatPair("time", m.TimeStamp))
	// add logger name if set.
	if len(l.name) > 0 {
		output = append(output,
			l.formatPair("logger", l.name))
	}
	// add level

	output = append(output,
		l.formatPair("level", formatLevel(m.LogLevel.String())))

	// Add the fields that were passed in to this message:
	if len(m.Fields) > 0 {
		output = append(output, m.Fields)
	}
	output = append(output, l.formatPair("message", m.Message))
	outputStr := strings.Join(output, " ") + "\n"
	return outputStr
}

func formatLevel(s string) string {
	s = strings.TrimSuffix(s, "Level")
	s = strings.ToLower(s)
	return s
}

// handleLogOutput takes a message and spits it out on all available output
// Todo: Factor out the formatting code from this.
func (l *CircularLogger) handleLogOutput(m LogMessage) {
	if m.LogLevel < l.printLevel {
		return
	}
	outputStr := l.formatMessage(m)
	for _, outputW := range l.outputWriters {
		_, err := io.WriteString(outputW, outputStr)
		if err != nil {
			fmt.Printf("Internal error in chainsaw: Can't write to outputWriter: %s", err)
		}
	}
}

func (l *CircularLogger) handleReset() error {
	l.current = 0
	l.messages = make([]LogMessage, l.logBufferSize)
	return nil
}

func (l *CircularLogger) addOutputChan(ch logChan) error {
	found := false
	for _, outputCh := range l.outputChs {
		if outputCh == ch {
			found = true
		}
	}
	if found {
		return fmt.Errorf("output channel already added to the logger")
	}
	l.outputChs = append(l.outputChs, ch)
	return nil
}

// removeOutputChan removes a channel from the list of output channels.
// it will also close that channel.
func (l *CircularLogger) removeOutputChan(ch logChan) error {
	found := false
	for i, outputCh := range l.outputChs {
		if outputCh == ch {
			found = true
			l.outputChs[i] = l.outputChs[len(l.outputChs)-1] // Take the last element and put it at the place we've found
			l.outputChs = l.outputChs[:len(l.outputChs)-1]   // shrink the slice.
			close(ch)                                        // close the channel so the listener will be notified.
		}
	}
	if !found {
		return fmt.Errorf("output channel not found")
	}
	return nil
}

func (l *CircularLogger) addWriter(o io.Writer) error {
	found := false
	for _, writer := range l.outputWriters {
		if writer == o {
			found = true
		}
	}
	if found {
		return fmt.Errorf("writer already present in logger: %v", o)
	}
	l.outputWriters = append(l.outputWriters, o)
	return nil
}

func (l *CircularLogger) removeWriter(o io.Writer) error {
	found := false
	for i, wr := range l.outputWriters {
		if o == wr {
			found = true
			l.outputWriters[i] = l.outputWriters[len(l.outputWriters)-1]
			l.outputWriters = l.outputWriters[:len(l.outputWriters)-1]
		}
	}
	if !found {
		return fmt.Errorf("writer not found in logger: %v", o)
	}
	return nil
}

// getMessageOverCh is used by the internal goroutine to fetch log-lines
// Perhaps it would be more efficient to stream these over a channel instead.
// However, in terms of allocation this will do one big allocation and not many smaller ones.
func (l *CircularLogger) getMessageOverCh(level LogLevel) []LogMessage {
	buf := make([]LogMessage, 0)
	for i := l.current; i < l.current+l.logBufferSize; i++ {
		msg := l.messages[i%l.logBufferSize]
		// Add the fields from the logger:
		if msg.LogLevel < level {
			continue
		}
		buf = append(buf, msg)
	}
	return buf
}

func (l *CircularLogger) setLevel(level LogLevel) error {
	l.printLevel = level
	return nil
}

// sendCtrlAndWait is used to send a message on the control channel
// and return the error back. This keeps the API calls cleaner.
func (l *CircularLogger) sendCtrlAndWait(ctrlMsg controlMessage) error {
	errCh := make(chan error, 1)
	ctrlMsg.errChan = errCh // mutate the struct a bit.
	l.controlCh <- ctrlMsg
	return <-errCh // Return the error returned by the logger goroutine.
}

// backTrace logs messages from the current buffer to the log file when a log message has a high enough
// log level.
// This happens in one single write, so it'll be continuous in the logs.
func (l *CircularLogger) backTrace() error {
	msgs := l.GetMessages(TraceLevel)
	msgStrings := make([]string, len(msgs))
	for _, msg := range msgs {
		msgStrings = append(msgStrings, l.formatMessage(msg))
	}
	joined := strings.Join(msgStrings, "")
	for _, w := range l.outputWriters {
		_, _ = w.Write([]byte(l.formatMessage(l.traceMessage(true))))
		_, err := w.Write([]byte(joined))
		if err != nil {
			return fmt.Errorf("writing to output writer: %w", err)
		}
		_, _ = w.Write([]byte(l.formatMessage(l.traceMessage(false))))
	}
	l.Reset() // Clear the logs.
	return nil
}

func (l *CircularLogger) traceMessage(begin bool) LogMessage {
	var msg string
	if begin {
		msg = fmt.Sprintf("=== %s backtrace begins ===", l.name)
	} else {
		msg = fmt.Sprintf("=== %s backtrace ends ===", l.name)
	}
	return LogMessage{
		LogLevel:  l.backTraceLevel,
		Message:   msg,
		TimeStamp: time.Now(),
	}
}
