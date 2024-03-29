package chainsaw

import (
	"bytes"
	"context"
	"fmt"
	is2 "github.com/matryer/is"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	testDefaultLogBufferSize  = 50
	testDefaultChanBufferSize = 0                    // In tests we run with unbuffered channels to detect deadlocks.
	defaultSleepTime          = 1 * time.Millisecond // The default sleep time to let the logger finish its async work.
)

func TestDemo(t *testing.T) {
	is := is2.New(t)
	logger := MakeLogger("test")
	logger.SetLevel(TraceLevel)
	logger.Trace("trace", 5, 1.0, false)
	logger.Debug("debug")
	logger.Info("info", "info")
	logger.Warn("warn")
	logger.Error("error")

	// 2021-11-11T08:19:42+0100/test: [info] This message is an info message
	err := logger.Flush()
	is.NoErr(err)
	msgs := logger.GetMessages(TraceLevel)
	for _, msg := range msgs {
		fmt.Println(logger.formatMessage(msg))
	}
	logger.Stop()
}

func TestNilLogger(t *testing.T) {
	var logger *CircularLogger
	logger.Trace("This should not panic")
	logger.Debugf("This should not panic either: %d", 1)
	logger.Infow("This should not panic either", P{"test", 1})
	logger.Warn("This should not panic either")
	logger.Errorf("This should not panic either: %d", 1)
	logger.Errorw("This should not panic either", P{"test", 1})
	logger.Println("This should not panic either")
	logger.Printf("This should not panic either: %d", 1)
}

func TestLoggingPerformance(t *testing.T) {
	is := is2.New(t)

	const runs = 20000
	const timeout = 3 * time.Millisecond
	logger := MakeLogger("", testDefaultLogBufferSize, testDefaultChanBufferSize)
	err := logger.RemoveWriter(os.Stderr) // Reduce noise.
	is.NoErr(err)
	start := time.Now()
	for i := 0; i < runs; i++ {
		logger.Debug("Dummy message")
	}
	dur := time.Since(start)
	avg := dur / runs
	fmt.Printf("Duration per logging invokation: %v\n", avg)
	is.True(avg < timeout) // Check if we are somewhat performant.
	_ = logger.Flush()
	start = time.Now()
	for i := 0; i < runs; i++ {
		_ = logger.Flush()
	}
	dur = time.Since(start)
	avg = dur / runs
	fmt.Printf("Duration per flush invokation: %v\n", avg)
	is.True(avg < timeout) // Check if we are somewhat performant.
}

func TestLogging(t *testing.T) {
	logger := MakeLogger("", testDefaultLogBufferSize, testDefaultChanBufferSize)
	defer logger.Stop()
	is := is2.New(t)
	stringOutput := &stringLogger{}
	err := logger.AddWriter(stringOutput)
	is.NoErr(err)
	err = logger.RemoveWriter(os.Stderr)
	is.NoErr(err)
	logger.Trace("Trace", "concatenated")
	logger.Tracef("Tracef message: %d", 1)
	logger.Tracew("Trace field", P{"test", 1})

	logger.Debug("Debug message", "concatenated")
	logger.Debugf("Debugf message: %d", 1)
	logger.Debugw("Debug field", P{"test", 2})

	logger.Info("Info", "concatenated")
	logger.Infof("Infof message: %d", 1)
	logger.Infow("Info field", P{"infotest", 3})

	logger.Warn("Warn", "concatenated")
	logger.Warnf("Warnf message: %d", 1)
	logger.Warnw("Warn field", P{"warntest", 4})

	logger.Error("Error", "concatenated")
	logger.Errorf("Errorf message: %d", 1)
	logger.Errorw("Error field", P{"errortest", 5})
	_ = logger.Flush()
	is.Equal(len(stringOutput.loglines), 9)
	is.True(!stringOutput.contains("Trace")) // no traces should have been logged.
	is.True(!stringOutput.contains("Debug")) // no debug should have been logged.
	// Check that the output arrived where we expected it to.
	is.True(strings.Contains(stringOutput.loglines[0], "Info concatenated"))
	is.True(strings.Contains(stringOutput.loglines[1], "Infof message: 1"))
	is.True(strings.Contains(stringOutput.loglines[2], "infotest=3"))
	is.True(strings.Contains(stringOutput.loglines[3], "Warn concatenated"))
	is.True(strings.Contains(stringOutput.loglines[4], "Warnf message: 1"))
	is.True(strings.Contains(stringOutput.loglines[5], "warntest=4"))
	is.True(strings.Contains(stringOutput.loglines[6], "Error concatenated"))
	is.True(strings.Contains(stringOutput.loglines[7], "Errorf message: 1"))
	is.True(strings.Contains(stringOutput.loglines[8], "errortest=5"))
}

func Test_Print(t *testing.T) {
	logger := MakeLogger("", testDefaultLogBufferSize, testDefaultChanBufferSize)
	defer logger.Stop()
	is := is2.New(t)
	stringOutput := &stringLogger{}
	err := logger.AddWriter(stringOutput)
	is.NoErr(err)
	err = logger.RemoveWriter(os.Stderr)
	is.NoErr(err)
	logger.Println("Println message")
	logger.Printf("Printf message: %d", 1)
	err = logger.Flush()
	is.NoErr(err)
	is.Equal(len(stringOutput.loglines), 2)
	is.True(stringOutput.contains("Println"))
	is.True(stringOutput.contains("Printf message: 1"))
}

// TestRemoveWriter uses the default logger instance.
func TestRemoveWriter(t *testing.T) {
	is := is2.New(t)
	tlog := MakeLogger("test", testDefaultLogBufferSize, testDefaultChanBufferSize)
	defer tlog.Stop()
	tlog.SetLevel(InfoLevel)
	buffer := &SafeBuffer{}
	err := tlog.AddWriter(buffer)
	is.NoErr(err)
	err = tlog.RemoveWriter(os.Stderr)
	is.NoErr(err)
	tlog.Trace("Trace message")
	tlog.Tracef("Tracef message: %d", 1)
	tlog.Debug("Debug message")
	tlog.Debugf("Debugf message: %d", 1)
	tlog.Info("Info message")
	tlog.Infof("Infof message: %d", 1)
	tlog.Warn("Warn message")
	tlog.Warnf("Warnf message: %d", 1)
	tlog.Error("Error message")
	tlog.Errorf("Errorf message: %d", 1)
	err = tlog.Flush()
	is.NoErr(err)
	time.Sleep(defaultSleepTime)
	is.Equal(len(tlog.GetMessages(InfoLevel)), 6)
	is.Equal(len(tlog.GetMessages(WarnLevel)), 4)
	is.Equal(len(tlog.GetMessages(ErrorLevel)), 2)
	err = tlog.RemoveWriter(buffer)
	is.NoErr(err)
	time.Sleep(defaultSleepTime)
	Error("XXX message")
	Errorf("XXXf message: %d", 1)
	time.Sleep(defaultSleepTime)
	bufferBytes := buffer.Bytes()
	is.True(!bytes.Contains(bufferBytes, []byte("Trace message")))
	is.True(!bytes.Contains(bufferBytes, []byte("Tracef message: 1")))
	is.True(!bytes.Contains(bufferBytes, []byte("Debug message")))
	is.True(!bytes.Contains(bufferBytes, []byte("Debugf message: 1")))
	is.True(bytes.Contains(bufferBytes, []byte("Info message")))
	is.True(bytes.Contains(bufferBytes, []byte("Infof message: 1")))
	is.True(bytes.Contains(bufferBytes, []byte("Warn message")))
	is.True(bytes.Contains(bufferBytes, []byte("Warnf message: 1")))
	is.True(bytes.Contains(bufferBytes, []byte("Error message")))
	is.True(bytes.Contains(bufferBytes, []byte("Errorf message: 1")))
	is.True(!bytes.Contains(bufferBytes, []byte("XXX message")))
	is.True(!bytes.Contains(bufferBytes, []byte("XXXf message: 1")))

}

// TestDumpMessages makes a few log messages. Fetches them back and sees that they are all there.
func TestDumpMessages(t *testing.T) {
	is := is2.New(t)
	const logBufferSize = 50
	logger := MakeLogger("", logBufferSize, testDefaultChanBufferSize)
	err := logger.RemoveWriter(os.Stderr)
	is.NoErr(err)
	defer logger.Stop()
	const noOfMessages = 5
	var counted = 0
	for i := 0; i < noOfMessages; i++ {
		logger.Tracef("Trace message %d", i)
		logger.Debugf("Debug message %d", i)
		logger.Infof("Info message %d", i)
		logger.Warnf("Warn message %d", i)
		logger.Errorf("Error message %d", i)
		counted += 5
	}
	err = logger.Flush()
	is.NoErr(err)
	time.Sleep(defaultSleepTime) // Sleep a few ms while the logs get to the right place.
	fmt.Printf("Generated %d messages\n", counted)
	msgTrace := logger.GetMessages(TraceLevel)
	verifyLogLevel(is, msgTrace, TraceLevel)
	msgDebug := logger.GetMessages(DebugLevel)
	verifyLogLevel(is, msgDebug, DebugLevel)
	msgInfo := logger.GetMessages(InfoLevel)
	verifyLogLevel(is, msgInfo, InfoLevel)
	msgWarn := logger.GetMessages(WarnLevel)
	verifyLogLevel(is, msgWarn, WarnLevel)
	msgError := logger.GetMessages(ErrorLevel)
	verifyLogLevel(is, msgWarn, WarnLevel)
	is.Equal(len(msgTrace), noOfMessages*5) // Trace
	is.Equal(len(msgDebug), noOfMessages*4) // Debug
	is.Equal(len(msgInfo), noOfMessages*3)  // Info
	is.Equal(len(msgWarn), noOfMessages*2)  // Warn
	is.Equal(len(msgError), noOfMessages*1) // Error
	is.Equal(counted, noOfMessages*5)
	fmt.Println("All messages accounted for")

}

func verifyLogLevel(is *is2.I, msgs []LogMessage, level LogLevel) {
	for _, m := range msgs {
		is.True(m.LogLevel >= level) // Verifies that the log level is what we expect or higher
	}
}

// TestDumpLimited tests overrunning the log buffer so we can make sure it is actually circular
func TestDumpLimited(t *testing.T) {
	is := is2.New(t)
	const logBufferSize = 10
	const logBufferOverrun = logBufferSize * 2
	logger := MakeLogger("", logBufferSize, testDefaultChanBufferSize)
	err := logger.RemoveWriter(os.Stderr)
	is.NoErr(err)
	defer logger.Stop()
	fmt.Printf("Generating %d trace messages...\n", logBufferSize)
	for i := 0; i < logBufferSize; i++ {
		logger.Tracef("Trace message %d/%d", i, testDefaultLogBufferSize)
	}
	fmt.Printf("overrunning the buffer with %d more messages\n", logBufferSize)
	for i := 0; i < logBufferOverrun; i++ {
		logger.Infof("Info message %d", i)
	}
	err = logger.Flush()
	is.NoErr(err)
	msgs := logger.GetMessages(InfoLevel)
	fmt.Printf("Got %d messages from the log system\n", len(msgs))
	is.Equal(len(msgs), logBufferSize)
	for i, m := range msgs {
		is.Equal(fmt.Sprintf("Info message %d", i+logBufferSize), m.Message)
	}
	fmt.Println("Validated log content")
}

// Create a stream and do various verifications that it works.
func TestStream(t *testing.T) {
	is := is2.New(t)
	const noOfMessages = 20
	testLogger := MakeLogger("", testDefaultLogBufferSize, 0) // Use zero buffer to provoke races.
	testLogger.SetLevel(TraceLevel)
	err := testLogger.RemoveWriter(os.Stderr) // reduce console noise.
	is.NoErr(err)
	defer testLogger.Stop()

	// The ctx we pass into getstream to stop it.
	streamCtx, streamCancel := context.WithCancel(context.Background())
	stream := testLogger.GetStream(streamCtx)
	streamedMessages := SafeInt{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(streamCh chan LogMessage) {
		for msg := range streamCh {
			fmt.Print(".")
			is.Equal(msg.LogLevel, TraceLevel) // Verify that these are debug messages.
			// And verify the content:
			is.Equal(fmt.Sprintf("Trace message %d/%d", streamedMessages.Get(), noOfMessages), msg.Message)
			streamedMessages.Inc()
		}
		wg.Done()
		fmt.Printf("streamer done. %d messages verified.\n", streamedMessages.Get())
	}(stream)
	fmt.Printf("Hello there, will fire off %d trace messages\n", noOfMessages)
	for i := 0; i < noOfMessages; i++ {
		testLogger.Tracef("Trace message %d/%d", i, noOfMessages)
	}
	err = testLogger.Flush()
	is.NoErr(err)
	streamCancel()                                 // Cancel the stream. This should force the above goroutine to exit.
	wg.Wait()                                      // Wait for the streamer to exit.
	is.Equal(streamedMessages.Get(), noOfMessages) // Compare the number of messages stream to what we sent.
	fmt.Println("All messages reached the stream")
	fmt.Println("Will issue more messages that will not hit the stream")
	for i := 0; i < noOfMessages; i++ {
		testLogger.Infof("Messages not hitting the stream %d/%d", i, noOfMessages)
	}
	err = testLogger.Flush()
	is.NoErr(err)
	is.Equal(streamedMessages.Get(), noOfMessages) // Compare the number of messages stream to what we sent.
	is.Equal(len(testLogger.GetMessages(InfoLevel)), noOfMessages)
	fmt.Println("Stream test passed")
}

// TestStreamBlocked will create a logger, open a stream, fail to service that stream and
// detect if time out.
// If we don't carefully write to channels this will cause a deadlock panic.
func TestStreamBlocked(t *testing.T) {
	is := is2.New(t)
	testLogger := MakeLogger("", 10, 0) // Unbuffered so we provoke races.
	defer testLogger.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	stream := testLogger.GetStream(ctx)
	wg := sync.WaitGroup{}
	wg.Add(1)
	start := time.Now()
	go func() {
		testLogger.Info("Silly message #1")
		_ = testLogger.Flush() // will block and trigger deadlock if it isn't handled.
		wg.Done()
	}()
	wg.Wait()
	cancel()
	fmt.Println(<-stream)
	timeTaken := time.Since(start)
	is.True(timeTaken < channelTimeout*2)
	fmt.Println("ok")
}

func handleStream(stream chan LogMessage, counter *SafeInt, wg *sync.WaitGroup) {
	for range stream {
		counter.Inc()
	}
	wg.Done()
}

// TestMultipleStreams
// Create a logger. Connect X stream to it. Log Y messages.
// See that the number of messages reaching the streams is X * Y
func TestMultipleStreams(t *testing.T) {
	is := is2.New(t)
	const noOfMessages = 20
	const noOfStreams = 10
	testLogger := MakeLogger("", testDefaultLogBufferSize, 0) // Use zero buffer to provoke races.
	testLogger.SetLevel(TraceLevel)
	err := testLogger.RemoveWriter(os.Stderr) // reduce console noise.
	is.NoErr(err)
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	counter := &SafeInt{}
	wg.Add(noOfStreams)
	for i := 0; i < noOfStreams; i++ {
		go handleStream(testLogger.GetStream(ctx), counter, &wg)
	}
	for i := 0; i < noOfMessages; i++ {
		testLogger.Trace("trace message")
	}
	err = testLogger.Flush()
	is.NoErr(err)
	time.Sleep(defaultSleepTime)
	cancel()
	wg.Wait() // Wait for the handleStream goroutines to finish.
	is.Equal(counter.Get(), noOfStreams*noOfMessages)
	println("Done")
}

func TestStreamRace(t *testing.T) {
	is := is2.New(t)
	const noOfLoggers = 20
	const noOfMessages = 100
	wg := sync.WaitGroup{}
	wg.Add(noOfLoggers)
	ctx, cancel := context.WithCancel(context.Background())
	messagesStreamed := SafeInt{}

	for i := 0; i < noOfLoggers; i++ {
		go func() {
			logger := MakeLogger("", 10, 0)
			stream := logger.GetStream(ctx)
			go func(streamCh chan LogMessage) {
				for range streamCh {
					fmt.Print(".")
					messagesStreamed.Inc()
				}
				wg.Done()
			}(stream)
			for j := 0; j < noOfMessages; j++ {
				logger.Tracef("trace message")
			}
			err := logger.Flush()
			is.NoErr(err)
		}()
	}
	// We've pushed a lot of messages into the logger. Let's cancel the context.
	cancel()
	wg.Wait()

	is.True(messagesStreamed.Get() > 10)
	fmt.Printf("\nNo of messages reaching the streams:%d\n", messagesStreamed.Get())
}

func TestManyLoggers(t *testing.T) {
	const (
		noOfLoggers       = 10
		messagesPerLogger = 1000
		logBufferSize     = 10
	)
	fmt.Printf("Running %d loggers with %d logbuffersize and %d messages per logger\n", noOfLoggers, logBufferSize, messagesPerLogger)
	is := is2.New(t)
	loggers := make([]*CircularLogger, 10)

	for i := 0; i < noOfLoggers; i++ {
		loggers[i] = MakeLogger("", logBufferSize, testDefaultChanBufferSize)
	}
	defer func() {
		for _, logger := range loggers {
			logger.Stop()
		}
	}()

	for i := 0; i < messagesPerLogger; i++ {
		for l, logger := range loggers {
			logger.Tracef("Message %d on logger %d", i, l)
		}
	}
	for i, logger := range loggers {
		err := logger.Flush()
		is.NoErr(err)
		msgs := logger.GetMessages(TraceLevel)
		m := msgs[0]
		is.Equal(fmt.Sprintf("Message %d on logger %d", messagesPerLogger-logBufferSize, i), m.Message)
	}
	fmt.Println("Done")
}

func TestOutput(t *testing.T) {
	testLogger := MakeLogger("", 10, 0)
	is := is2.New(t)
	err := testLogger.RemoveWriter(os.Stderr)
	is.NoErr(err)
	err = testLogger.RemoveWriter(os.Stderr)
	is.True(err != nil)
	fmt.Println("err as expected", err.Error())
}

func TestFatal(t *testing.T) {
	if os.Getenv("BE_FATAL") == "1" {
		log.Fatal("Fatal log message")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFatal")
	cmd.Env = append(os.Environ(), "BE_FATAL=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func useInterface(logger Logger) {
	logger.Trace("trace")
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("info")
	logger.Error("error")

	logger.Tracef("trace %d", 1)
	logger.Debugf("debug %d", 1)
	logger.Infof("info %d", 1)
	logger.Warnf("info %d", 1)
	logger.Errorf("error %d", 1)
}

func TestInterface(t *testing.T) {
	is := is2.New(t)
	logger := MakeLogger("", 20, 0)
	useInterface(logger)
	err := logger.Flush()
	is.NoErr(err)
	msgs := logger.GetMessages(TraceLevel)
	is.Equal(len(msgs), 10)
}

func TestFields0(t *testing.T) {
	is := is2.New(t)
	var err error
	logger := MakeLogger("test", 20, 0)
	logger.SetLevel(TraceLevel)
	outputBuffer := bytes.NewBuffer(make([]byte, 0))
	err = logger.RemoveWriter(os.Stderr)
	is.NoErr(err)
	err = logger.AddWriter(outputBuffer)
	is.NoErr(err)
	logger.SetFields(P{"field1", "value1"})
	logger.Tracew("trace level message", P{"field2", "value2"})
	logger.Debugw("debug level message", P{"field2", "value2"})
	logger.Infow("info level message", P{"field2", "value2"})
	logger.Warnw("warn level message", P{"field2", "value2"})
	logger.Errorw("error level message", P{"field2", "value2"})

	msgs := logger.GetMessages(TraceLevel) // implicit flush
	is.Equal(len(msgs), 5)
	for _, event := range msgs {
		is.True(strings.Contains(event.Fields, "field1=value1"))
		is.True(strings.Contains(event.Fields, "field2=value2"))
	}
	lines := 0
	for _, line := range strings.Split(outputBuffer.String(), "\n") {
		if len(line) == 0 {
			continue
		}
		is.Equal(strings.Count(line, "field1=value1"), 1)
		is.Equal(strings.Count(line, "field2=value2"), 1)
		lines++
	}
	is.Equal(lines, 5)
}

func TestBacktrace(t *testing.T) {
	explevels := []string{"trace", "debug", "info", "error"}
	is := is2.New(t)
	var err error
	logger := MakeLogger("test", 20, 0)
	logger.SetLevel(TraceLevel)
	logger.SetBackTraceLevel(WarnLevel)
	outputBuffer := bytes.NewBuffer(make([]byte, 0))
	err = logger.RemoveWriter(os.Stderr)
	is.NoErr(err)
	err = logger.AddWriter(outputBuffer)
	is.NoErr(err)
	logger.Trace("trace level message")
	logger.Debug("debug level message")
	logger.Info("info level message")
	logger.Error("error level message")

	logMsg := logger.GetMessages(TraceLevel)
	is.Equal(len(logMsg), 0) // should have no messages, they have been flushed by the backtrace.
	lines := strings.Split(outputBuffer.String(), "\n")
	for i, explevel := range explevels {
		fields := parseFields(lines[i])
		fmt.Printf("line %d %+v: \n", i, fields)
		is.Equal(fields["level"], explevel)
	}
	is.True(strings.Contains(lines[4], "begins"))
	is.True(strings.Contains(lines[5], "trace"))
	is.True(strings.Contains(lines[6], "debug"))
	is.True(strings.Contains(lines[7], "info"))
	is.True(strings.Contains(lines[8], "error"))
	is.True(strings.Contains(lines[9], "ends"))
	is.Equal(lines[10], "")
	is.Equal(len(lines), 11) // 10 lines of logs + 1 blank line
}

// takes a string and returns a map[string]string of key value pairs
// key and value are separated by '='
// key value pairs are separated by ' '
func parseFields(line string) map[string]string {
	fields := make(map[string]string)
	for _, field := range strings.Split(line, " ") {
		kv := strings.Split(field, "=")
		if len(kv) == 2 {
			fields[kv[0]] = kv[1]
		}
	}
	return fields
}
