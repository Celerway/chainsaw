package chainsaw

// Logger is a basic logging interface which should allow you to pass a chainsaw or a logrus instance around
type Logger interface {
	Trace(...interface{})
	Tracef(string, ...interface{})

	Debug(...interface{})
	Debugf(string, ...interface{})

	Info(...interface{})
	Infof(string, ...interface{})

	Warn(...interface{})
	Warnf(string, ...interface{})

	Error(...interface{})
	Errorf(string, ...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
}
