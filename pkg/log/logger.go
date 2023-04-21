package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InfoLogger represents the ability to log non-error messages, at a particular verbosity.
type InfoLogger interface {
	Info(msg string, fields ...Field)
	Infof(format string, v ...interface{})
	Infow(msg string, keysAndValues ...interface{})

	// Enabled tests whether this InfoLogger is enabled.  For example,
	// commandline flags might be used to set the logging verbosity and disable
	// some info logs.
	Enabled() bool
}

type Logger interface {
}

type infoLogger struct {
	level zapcore.Level
	log   *zap.Logger
}

// zapLogger is a logr.Logger that uses Zap to log.
type zapLogger struct {
	// NB: this looks very similar to zap.SugaredLogger, but
	// deals with our desire to have multiple verbosity levels.
	zapLogger *zap.Logger
	infoLogger
}

// noopInfoLogger is a logr.InfoLogger that's always disabled, and does nothing.
type noopInfoLogger struct {
}

var _ Logger = &zapLogger{}

// Info method output info level log.
func Info(msg string, fields ...Field) {
	std.zapLogger.Info(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.zapLogger.Info(msg, fields...)
}
