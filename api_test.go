package chainsaw

import (
	"bytes"
	"fmt"
	is2 "github.com/matryer/is"
	"os"
	"testing"
	"time"
)

func TestParseLogLevel(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    LogLevel
		wantErr bool
	}{
		{"trace", args{"trace"}, TraceLevel, false},
		{"debug", args{"debug"}, DebugLevel, false},
		{"info", args{"info"}, InfoLevel, false},
		{"warn", args{"warn"}, WarnLevel, false},
		{"error", args{"error"}, ErrorLevel, false},
		{"fatal", args{"fatal"}, FatalLevel, false},
		{"tracelevel", args{"tracelevel"}, TraceLevel, false},
		{"debuglevel", args{"debuglevel"}, DebugLevel, false},
		{"infolevel", args{"infolevel"}, InfoLevel, false},
		{"warnlevel", args{"warnlevel"}, WarnLevel, false},
		{"errorlevel", args{"errorlevel"}, ErrorLevel, false},
		{"fatallevel", args{"fatallevel"}, FatalLevel, false},
		{"traceUpper", args{"TRACE"}, TraceLevel, false},
		{"debugUpper", args{"DEBUG"}, DebugLevel, false},
		{"infoUpper", args{"INFO"}, InfoLevel, false},
		{"warnUpper", args{"WARN"}, WarnLevel, false},
		{"errorUpper", args{"ERROR"}, ErrorLevel, false},
		{"fatalUpper", args{"FATAL"}, FatalLevel, false},
		{"empty", args{""}, InfoLevel, true},
		{"invalid", args{"blalbla"}, InfoLevel, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLogLevel(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLogLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircularLogger_BackTrace(t *testing.T) {
	is := is2.New(t)

	logger := MakeLogger("foo", 30, 10)
	err := logger.RemoveWriter(os.Stderr)
	is.NoErr(err)
	logger.SetBackTraceLevel(ErrorLevel)
	logger.SetLevel(ErrorLevel)
	logBuffer := bytes.NewBuffer(make([]byte, 0, 1024))
	err = logger.AddWriter(logBuffer)
	is.NoErr(err)

	logger.Trace("trace")
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
	err = logger.Flush() // trigger a flush, so we know things are flushed out to the buffer
	is.NoErr(err)
	time.Sleep(time.Millisecond)
	lines := bytes.Split(logBuffer.Bytes(), []byte{'\n'})
	for i, line := range lines {
		fmt.Printf("line %d: %s\n", i, line)
	}
	is.True(bytes.Contains(lines[0], []byte("error")))
	is.True(bytes.Contains(lines[1], []byte("backtrace begin")))
	is.True(bytes.Contains(lines[2], []byte("trace")))
	is.True(bytes.Contains(lines[3], []byte("debug")))
	is.True(bytes.Contains(lines[4], []byte("info")))
	is.True(bytes.Contains(lines[5], []byte("warn")))
	is.True(bytes.Contains(lines[6], []byte("error")))
	is.True(bytes.Contains(lines[7], []byte("backtrace end")))
}
