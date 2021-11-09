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

func TestLogging(t *testing.T) {
	log := MakeLogger()
	defer log.Stop()
	time.Sleep(10 * time.Millisecond)
	is := is.New(t)
	buffer := bytes.NewBuffer(nil)
	log.AddOutput(buffer)
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
	log.AddOutput(os.Stdout)
	b := buffer.Bytes()
	is.True(bytes.Contains(b, []byte("Trace message")))
	is.True(bytes.Contains(b, []byte("Tracef message: 1")))
	is.True(bytes.Contains(b, []byte("Debug message")))
	is.True(bytes.Contains(b, []byte("Debugf message: 1")))
	is.True(bytes.Contains(b, []byte("Info message")))
	is.True(bytes.Contains(b, []byte("Infof message: 1")))
	is.True(bytes.Contains(b, []byte("Warn message")))
	is.True(bytes.Contains(b, []byte("Warnf message: 1")))
	is.True(bytes.Contains(b, []byte("Error message")))
	is.True(bytes.Contains(b, []byte("Errorf message: 1")))
}

// same as above, but with the default logger.
func TestLogging2(t *testing.T) {
	defer Stop()
	is := is.New(t)
	buffer := bytes.NewBuffer(nil)
	AddOutput(buffer)
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
	Error("XXX message")
	Errorf("XXXf message: %d", 1)
	b := buffer.Bytes()
	is.True(bytes.Contains(b, []byte("Trace message")))
	is.True(bytes.Contains(b, []byte("Tracef message: 1")))
	is.True(bytes.Contains(b, []byte("Debug message")))
	is.True(bytes.Contains(b, []byte("Debugf message: 1")))
	is.True(bytes.Contains(b, []byte("Info message")))
	is.True(bytes.Contains(b, []byte("Infof message: 1")))
	is.True(bytes.Contains(b, []byte("Warn message")))
	is.True(bytes.Contains(b, []byte("Warnf message: 1")))
	is.True(bytes.Contains(b, []byte("Error message")))
	is.True(bytes.Contains(b, []byte("Errorf message: 1")))
	is.True(!bytes.Contains(b, []byte("XXX message")))
	is.True(!bytes.Contains(b, []byte("XXXf message: 1")))

}

func TestDump(t *testing.T) {
	log := MakeLogger()
	defer log.Stop()
	is := is.New(t)
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
	time.Sleep(10 * time.Millisecond) // Sleep a few ms while the logs get to the right place.
	fmt.Printf("Generated %d messages. ", counted)
	msgTrace := log.GetMessages(TraceLevel)
	msgDebug := log.GetMessages(DebugLevel)
	msgInfo := log.GetMessages(InfoLevel)
	msgWarn := log.GetMessages(WarnLevel)
	msgError := log.GetMessages(ErrorLevel)
	is.Equal(len(msgTrace), noOfMessages*5) // Trace
	is.Equal(len(msgDebug), noOfMessages*4) // Debug
	is.Equal(len(msgInfo), noOfMessages*3)  // Info
	is.Equal(len(msgWarn), noOfMessages*2)  // Warn
	is.Equal(len(msgError), noOfMessages*1) // Error
}

func TestDumpLimited(t *testing.T) {
	log := MakeLogger()
	defer log.Stop()
	log.Info("Hello there, will fire off 20 trace messages.")
	for i := 0; i < 20; i++ {
		log.Tracef("Trace message %d", i)
	}
	log.Info("Done")
	log.Info("Hello there, will fire off 20 info messages.")
	for i := 0; i < 20; i++ {
		log.Infof("Info message %d", i)
	}
	log.Info("Done")
	msgs := log.GetMessages(InfoLevel)
	fmt.Printf("Got %d messages from the log system\n", len(msgs))
	for _, msg := range msgs {
		fmt.Printf("DUMP: %s\n", msg.Content)
	}
}

func TestStream(t *testing.T) {
	const noOfMessages = 20
	testLogger := MakeLogger()
	defer testLogger.Stop()
	log := MakeLogger()
	defer log.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	stream := testLogger.GetStream(ctx)
	streamedMessages := 0
	go func(c context.Context, s chan LogMessage) {
	loop:
		for {
			select {
			case <-c.Done():
				fmt.Println("Breaking out inner loop")
				break loop
			case msg := <-s:
				fmt.Printf("STREAM: %s|%s|%s\n", msg.TimeStamp.Format(time.RFC3339), msg.LogLevel, msg.Content)
				streamedMessages++
			}
		}

	}(ctx, stream)
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
	cancel()
	is := is.New(t)
	is.Equal(streamedMessages, noOfMessages) // Compare the number of messages stream to what we sent.
}

func TestQuit(t *testing.T) {
	log := MakeLogger()
	time.Sleep(10 * time.Millisecond)
	is := is.New(t)
	log.Info("test message")
	is.True(log.GetStatus()) // Goroutine should be running here.

	log.Stop()
	time.Sleep(10 * time.Millisecond)
	is.True(!log.GetStatus()) // Goroutine should be stopped here.
}

func TestManyLoggers(t *testing.T) {
	const many = 10
	is := is.New(t)
	loggers := make([]*CircularLogger, 10)

	for i := 0; i < many; i++ {
		loggers[i] = MakeLogger()
	}

	for i := 0; i < 1000; i++ {
		for l, logger := range loggers {
			logger.Tracef("Message %d on logger %d", i, l)
		}
	}
	for i, logger := range loggers {
		msgs := logger.GetMessages(TraceLevel)
		m := msgs[0]
		is.Equal(fmt.Sprintf("Message %d on logger %d", 850, i), m.Content)
	}

}
