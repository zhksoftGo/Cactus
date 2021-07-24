package Common

import "github.com/gookit/slog"

type Log4gnet struct {
}

// Debugf logs messages at DEBUG level.
func (l Log4gnet) Debugf(format string, args ...interface{}) {
	slog.Debugf(format, args...)
}

// Infof logs messages at INFO level.
func (l Log4gnet) Infof(format string, args ...interface{}) {
	slog.Infof(format, args...)
}

// Warnf logs messages at WARN level.
func (l Log4gnet) Warnf(format string, args ...interface{}) {
	slog.Warnf(format, args...)
}

// Errorf logs messages at ERROR level.
func (l Log4gnet) Errorf(format string, args ...interface{}) {
	slog.Errorf(format, args...)
}

// Fatalf logs messages at FATAL level.
func (l Log4gnet) Fatalf(format string, args ...interface{}) {
	slog.Fatalf(format, args...)
}
