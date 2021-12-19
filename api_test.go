package chainsaw

import "testing"

func TestParseLogLevel(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want LogLevel
	}{
		{"trace", args{"trace"}, TraceLevel},
		{"debug", args{"debug"}, DebugLevel},
		{"info", args{"info"}, InfoLevel},
		{"warn", args{"warn"}, WarnLevel},
		{"error", args{"error"}, ErrorLevel},
		{"fatal", args{"fatal"}, FatalLevel},
		{"traceUpper", args{"TRACE"}, TraceLevel},
		{"debugUpper", args{"DEBUG"}, DebugLevel},
		{"infoUpper", args{"INFO"}, InfoLevel},
		{"warnUpper", args{"WARN"}, WarnLevel},
		{"errorUpper", args{"ERROR"}, ErrorLevel},
		{"fatalUpper", args{"FATAL"}, FatalLevel},
		{"empty", args{""}, InfoLevel},
		{"invalid", args{"blalbla"}, InfoLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseLogLevel(tt.args.s); got != tt.want {
				t.Errorf("ParseLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
