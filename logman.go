package chainsaw

import (
	"fmt"
	"os"
)

func (l *CircularLogger) Trace(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.log(TraceLevel, s)
}

func (l *CircularLogger) Tracef(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(TraceLevel, s)
}

func (l *CircularLogger) Debug(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.log(DebugLevel, s)
}
func (l *CircularLogger) Debugf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(DebugLevel, s)
}

func (l *CircularLogger) Info(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.log(InfoLevel, s)
}

func (l *CircularLogger) Infof(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(InfoLevel, s)
}

func (l *CircularLogger) Warn(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.log(WarnLevel, s)
}

func (l *CircularLogger) Warnf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(WarnLevel, s)
}

func (l *CircularLogger) Error(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.log(ErrorLevel, s)
}

func (l *CircularLogger) Errorf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(ErrorLevel, s)
}

func (l *CircularLogger) Fatal(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.log(FatalLevel, s)
	os.Exit(1)
}

func (l *CircularLogger) Fatalf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(FatalLevel, s)
	os.Exit(1)
}
