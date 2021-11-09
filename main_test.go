package chainsaw

import (
	"bytes"
	"context"
	"fmt"
	"github.com/matryer/is"
	"os"
	"testing"
	"time"
)

const (
	defaultLogBufferSize  = 50
	defaultChanBufferSize = 0                    // In tests we run with unbuffered channels to detect deadlocks.
	defaultSleepTime      = 1 * time.Millisecond // The default sleep time to let the logger finish its async work.
)

func TestLogging(t *testing.T) {
	log := MakeLogger(defaultLogBufferSize, defaultChanBufferSize)
	defer log.Stop()
	time.Sleep(10 * time.Millisecond)
	is := is.New(t)
	buffer := bytes.NewBuffer(nil)
	log.AddOutput(buffer)
	time.Sleep(defaultSleepTime)
	log.Trace("Trace message")
	log.Tracef("Tracef message: %d", 1)
	log.Debug("Debug message")
	log.Debugf("Debugf message: %d", 1)
	log.Info("Info message")
	log.Infof("Infof message: %d", 1)
	log.Warn("Warn message")
	log.Warnf("Warnf message: %d", 1)
	log.Error("Error message")
	log.Errorf("Errorf message: %d", 1)
	time.Sleep(defaultSleepTime)
	b := buffer.Bytes()
	is.True(!bytes.Contains(b, []byte("Trace message")))
	is.True(!bytes.Contains(b, []byte("Tracef message: 1")))
	is.True(!bytes.Contains(b, []byte("Debug message")))
	is.True(!bytes.Contains(b, []byte("Debugf message: 1")))
	is.True(bytes.Contains(b, []byte("Info message")))
	is.True(bytes.Contains(b, []byte("Infof message: 1")))
	is.True(bytes.Contains(b, []byte("Warn message")))
	is.True(bytes.Contains(b, []byte("Warnf message: 1")))
	is.True(bytes.Contains(b, []byte("Error message")))
	is.True(bytes.Contains(b, []byte("Errorf message: 1")))
}

// TestRemoveWriter uses the default logger instance.
func TestRemoveWriter(t *testing.T) {
	defer Stop()
	is := is.New(t)
	buffer := bytes.NewBuffer(nil)
	AddOutput(buffer)
	time.Sleep(defaultSleepTime)
	Trace("Trace message")
	Tracef("Tracef message: %d", 1)
	Debug("Debug message")
	Debugf("Debugf message: %d", 1)
	Info("Info message")
	Infof("Infof message: %d", 1)
	Warn("Warn message")
	Warnf("Warnf message: %d", 1)
	Error("Error message")
	Errorf("Errorf message: %d", 1)
	RemoveWriter(buffer)
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
	const logBufferSize = 50
	log := MakeLogger(logBufferSize, defaultChanBufferSize)
	log.RemoveWriter(os.Stdout)
	defer log.Stop()
	const noOfMessages = 5
	var counted = 0
	for i := 0; i < noOfMessages; i++ {
		log.Tracef("Trace message %d", i)
		log.Debugf("Debug message %d", i)
		log.Infof("Info message %d", i)
		log.Warnf("Warn message %d", i)
		log.Errorf("Error message %d", i)
		counted += 5
	}
	time.Sleep(defaultSleepTime) // Sleep a few ms while the logs get to the right place.
	fmt.Printf("Generated %d messages\n", counted)
	msgTrace := log.GetMessages(TraceLevel)
	msgDebug := log.GetMessages(DebugLevel)
	msgInfo := log.GetMessages(InfoLevel)
	msgWarn := log.GetMessages(WarnLevel)
	msgError := log.GetMessages(ErrorLevel)
	is := is.New(t)
	is.Equal(len(msgTrace), noOfMessages*5) // Trace
	is.Equal(len(msgDebug), noOfMessages*4) // Debug
	is.Equal(len(msgInfo), noOfMessages*3)  // Info
	is.Equal(len(msgWarn), noOfMessages*2)  // Warn
	is.Equal(len(msgError), noOfMessages*1) // Error
	is.Equal(counted, noOfMessages*5)
	fmt.Println("All messages accounted for")

}

// TestDumpLimited tests overrunning the log buffer so we can make sure it is actually circular
func TestDumpLimited(t *testing.T) {
	const logBufferSize = 10
	const logBufferOverrun = logBufferSize * 2
	log := MakeLogger(logBufferSize, defaultChanBufferSize)
	defer log.Stop()
	Infof("Hello there, will fire off %d trace messages.", logBufferSize)
	for i := 0; i < logBufferSize; i++ {
		log.Tracef("Trace message %d/%d", i, defaultLogBufferSize)
	}
	log.Infof("Will fire off %d info messages, overrunning the buffer", logBufferOverrun)
	for i := 0; i < logBufferOverrun; i++ {
		log.Infof("Info message %d", i)
	}
	msgs := log.GetMessages(InfoLevel)
	fmt.Printf("Got %d messages from the log system\n", len(msgs))
	is := is.New(t)
	is.Equal(len(msgs), logBufferSize)
	for i, m := range msgs {
		is.Equal(fmt.Sprintf("Info message %d", i+logBufferSize), m.Content)
	}

}

func TestStream(t *testing.T) {
	const noOfMessages = 20
	testLogger := MakeLogger(defaultLogBufferSize, defaultChanBufferSize)
	testLogger.SetLevel(TraceLevel)
	defer testLogger.Stop()

	log := MakeLogger(defaultLogBufferSize, defaultChanBufferSize)
	defer log.Stop()

	streamCtx, streamCancel := context.WithCancel(context.Background())

	stream := testLogger.GetStream(streamCtx)
	streamedMessages := 0
	go func(streamCh chan LogMessage) {
		for msg := range streamCh {
			fmt.Printf("STREAM: %s|%s|%s\n", msg.TimeStamp.Format(time.RFC3339), msg.LogLevel, msg.Content)
			streamedMessages++
		}
	}(stream)
	log.Infof("Hello there, will fire off %d trace messages.", noOfMessages)
	for i := 0; i < noOfMessages; i++ {
		testLogger.Tracef("Trace message %d/%d", i, noOfMessages)
	}
	log.Info("Done")
	// Wait a bit for noOfMessages to get right.
	for i := 0; i < 1000; i++ {
		time.Sleep(1 * time.Millisecond)
		if streamedMessages == noOfMessages {
			break
		}
	}
	is := is.New(t)
	is.Equal(streamedMessages, noOfMessages) // Compare the number of messages stream to what we sent.
	streamCancel()                           // Cancel the stream. This should force the above goroutine to exit.
	time.Sleep(defaultSleepTime)
	for i := 0; i < noOfMessages; i++ {
		testLogger.Tracef("Messages not hitting the stream %d/%d", i, noOfMessages)
	}
	is.Equal(streamedMessages, noOfMessages) // Compare the number of messages stream to what we sent.
}

func TestQuit(t *testing.T) {
	log := MakeLogger(defaultLogBufferSize, defaultChanBufferSize)
	time.Sleep(defaultSleepTime)
	is := is.New(t)
	log.Info("test message")
	is.True(log.GetStatus()) // Goroutine should be running here.

	log.Stop()
	time.Sleep(defaultSleepTime)
	is.True(!log.GetStatus()) // Goroutine should be stopped here.
}

func TestManyLoggers(t *testing.T) {
	const (
		many              = 10
		messagesPerLogger = 1000
		logBufferSize     = 10
	)

	is := is.New(t)
	loggers := make([]*CircularLogger, 10)

	for i := 0; i < many; i++ {
		loggers[i] = MakeLogger(logBufferSize, defaultChanBufferSize)
	}

	for i := 0; i < messagesPerLogger; i++ {
		for l, logger := range loggers {
			logger.Tracef("Message %d on logger %d", i, l)
		}
	}
	for i, logger := range loggers {
		msgs := logger.GetMessages(TraceLevel)
		m := msgs[0]
		is.Equal(fmt.Sprintf("Message %d on logger %d", messagesPerLogger-logBufferSize, i), m.Content)
	}

}
