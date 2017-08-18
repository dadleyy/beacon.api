package logging

// LeveledLogger is a logger that allows Errorf, etc.. logging.
type LeveledLogger interface {
	Errorf(string, ...interface{})
	Warnf(string, ...interface{})
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
}
