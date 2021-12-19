package chainsaw

import (
	"testing"
)

func TestCircularLogger_Infow(t *testing.T) {
	type args struct {
		msg   string
		pairs []P
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "simple",
			args: args{
				msg: "Message:",
				pairs: []P{
					{"username", "perbu"},
				},
			},
		}, {
			name: "contains spaces",
			args: args{
				msg: "Message:",
				pairs: []P{
					{"sentence", "five fat foxes"},
					{"foxes", "5"},
				},
			},
		},
	}
	l := MakeLogger("test")
	defer l.Stop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l.Infow(tt.args.msg, tt.args.pairs...)
		})
	}
}

func TestField_usage(t *testing.T) {
	l := MakeLogger("test")
	defer l.Stop()
	l.SetFields(P{"routerid", 42}, P{"username", "perbu"})
	l.AddFields(P{"country", "no"})
	l.Infof("Could not open file: %s", "no such file")
	l.Infow("Could not open file.", P{"err", "no such file"})
	l.Flush()
}
