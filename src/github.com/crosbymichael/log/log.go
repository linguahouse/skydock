package log

import (
	"fmt"
)

const (
	// FATAL error log levels
	FATAL = iota // fatal errors

	// ERROR error log levels
	ERROR = iota // errors might happend

	// INFO log only info level
	INFO = iota // debug mode

	// DEBUG level
	DEBUG = iota // debug mode
)

// Logger is a struct ..
type Logger struct {
	level int
	name  string
}

var currentlog Logger

// Logf is a logging function
func Logf(level int, format string, v ...interface{}) {
	if currentlog.level <= level {
		if v != nil {
			fmt.Println(format, v)
		} else {
			fmt.Println(format)
		}

	}
}

// SetLogger defines logger pameters in the current conext
func SetLogger(logger Logger) error {
	// currentlog := logger
	return nil
}

// NewStandardLevelLogger creates a new statand logger
func NewStandardLevelLogger(n string) Logger {
	logger := Logger{level: INFO, name: n}
	return logger
}
