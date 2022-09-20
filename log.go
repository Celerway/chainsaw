// Code generated by go generate; DO NOT EDIT.
// This file was generated by gen/main.go at 2022-09-20 16:00:00.129406 +0200 CEST m=+0.001519376
package chainsaw

import (
	"fmt"
	"strings"
)

func join(vals []interface{}) string {
	ss := make([]string, len(vals))
	for i, value := range vals {
		ss[i] = fmt.Sprint(value)
	}
	return strings.Join(ss, " ")
}

// Trace takes a number of arguments, makes them into strings and logs
// them with level TraceLevel
func (l *CircularLogger) Trace(v ...interface{}) {
	s := join(v)
	l.log(TraceLevel, s, "")
}

// Tracef takes a format and arguments, formats a string a logs it with TraceLevel
func (l *CircularLogger) Tracef(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(TraceLevel, s, "")
}

// Tracew takes a message and a number of pairs, type chainsaw.P. A fully structured
// log message will be generated with the pairs formatted into key/value pairs
// Note that keys must be strings, values can be strings, numbers, time.Time and a few other types.
func (l *CircularLogger) Tracew(msg string, pairs ...P) {
	fields := l.formatFields(pairs)
	l.log(TraceLevel, msg, fields)
}

func Trace(v ...interface{}) {
	l := defaultLogger
	l.Trace(v...)
}

func Tracef(f string, v ...interface{}) {
	l := defaultLogger
	l.Tracef(f, v...)
}

// Debug takes a number of arguments, makes them into strings and logs
// them with level DebugLevel
func (l *CircularLogger) Debug(v ...interface{}) {
	s := join(v)
	l.log(DebugLevel, s, "")
}

// Debugf takes a format and arguments, formats a string a logs it with DebugLevel
func (l *CircularLogger) Debugf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(DebugLevel, s, "")
}

// Debugw takes a message and a number of pairs, type chainsaw.P. A fully structured
// log message will be generated with the pairs formatted into key/value pairs
// Note that keys must be strings, values can be strings, numbers, time.Time and a few other types.
func (l *CircularLogger) Debugw(msg string, pairs ...P) {
	fields := l.formatFields(pairs)
	l.log(DebugLevel, msg, fields)
}

func Debug(v ...interface{}) {
	l := defaultLogger
	l.Debug(v...)
}

func Debugf(f string, v ...interface{}) {
	l := defaultLogger
	l.Debugf(f, v...)
}

// Info takes a number of arguments, makes them into strings and logs
// them with level InfoLevel
func (l *CircularLogger) Info(v ...interface{}) {
	s := join(v)
	l.log(InfoLevel, s, "")
}

// Infof takes a format and arguments, formats a string a logs it with InfoLevel
func (l *CircularLogger) Infof(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(InfoLevel, s, "")
}

// Infow takes a message and a number of pairs, type chainsaw.P. A fully structured
// log message will be generated with the pairs formatted into key/value pairs
// Note that keys must be strings, values can be strings, numbers, time.Time and a few other types.
func (l *CircularLogger) Infow(msg string, pairs ...P) {
	fields := l.formatFields(pairs)
	l.log(InfoLevel, msg, fields)
}

func Info(v ...interface{}) {
	l := defaultLogger
	l.Info(v...)
}

func Infof(f string, v ...interface{}) {
	l := defaultLogger
	l.Infof(f, v...)
}

// Warn takes a number of arguments, makes them into strings and logs
// them with level WarnLevel
func (l *CircularLogger) Warn(v ...interface{}) {
	s := join(v)
	l.log(WarnLevel, s, "")
}

// Warnf takes a format and arguments, formats a string a logs it with WarnLevel
func (l *CircularLogger) Warnf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(WarnLevel, s, "")
}

// Warnw takes a message and a number of pairs, type chainsaw.P. A fully structured
// log message will be generated with the pairs formatted into key/value pairs
// Note that keys must be strings, values can be strings, numbers, time.Time and a few other types.
func (l *CircularLogger) Warnw(msg string, pairs ...P) {
	fields := l.formatFields(pairs)
	l.log(WarnLevel, msg, fields)
}

func Warn(v ...interface{}) {
	l := defaultLogger
	l.Warn(v...)
}

func Warnf(f string, v ...interface{}) {
	l := defaultLogger
	l.Warnf(f, v...)
}

// Error takes a number of arguments, makes them into strings and logs
// them with level ErrorLevel
func (l *CircularLogger) Error(v ...interface{}) {
	s := join(v)
	l.log(ErrorLevel, s, "")
}

// Errorf takes a format and arguments, formats a string a logs it with ErrorLevel
func (l *CircularLogger) Errorf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(ErrorLevel, s, "")
}

// Errorw takes a message and a number of pairs, type chainsaw.P. A fully structured
// log message will be generated with the pairs formatted into key/value pairs
// Note that keys must be strings, values can be strings, numbers, time.Time and a few other types.
func (l *CircularLogger) Errorw(msg string, pairs ...P) {
	fields := l.formatFields(pairs)
	l.log(ErrorLevel, msg, fields)
}

func Error(v ...interface{}) {
	l := defaultLogger
	l.Error(v...)
}

func Errorf(f string, v ...interface{}) {
	l := defaultLogger
	l.Errorf(f, v...)
}

// Fatal takes a number of arguments, makes them into strings and logs
// them with level FatalLevel
func (l *CircularLogger) Fatal(v ...interface{}) {
	s := join(v)
	l.log(FatalLevel, s, "")
}

// Fatalf takes a format and arguments, formats a string a logs it with FatalLevel
func (l *CircularLogger) Fatalf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log(FatalLevel, s, "")
}

// Fatalw takes a message and a number of pairs, type chainsaw.P. A fully structured
// log message will be generated with the pairs formatted into key/value pairs
// Note that keys must be strings, values can be strings, numbers, time.Time and a few other types.
func (l *CircularLogger) Fatalw(msg string, pairs ...P) {
	fields := l.formatFields(pairs)
	l.log(FatalLevel, msg, fields)
}

func Fatal(v ...interface{}) {
	l := defaultLogger
	l.Fatal(v...)
}

func Fatalf(f string, v ...interface{}) {
	l := defaultLogger
	l.Fatalf(f, v...)
}
