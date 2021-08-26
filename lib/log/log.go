package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	formatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

type Log struct {
	logger *logrus.Entry
}

func (l *Log) Info(args ...interface{}){
	l.withFileField().Info(args...)
}

func (l *Log) Infof(format string, args ...interface{}){
	l.withFileField().Infof(format, args...)
}

func (l *Log) Debug(args ...interface{}){
	l.withFileField().Debug(args...)
}

func (l *Log) Debugf(format string, args ...interface{}){
	l.withFileField().Debugf(format, args...)
}

func (l *Log) Warn(args ...interface{}){
	l.withFileField().Warn(args...)
}

func (l *Log) Warnf(format string, args ...interface{}){
	l.withFileField().Infof(format, args...)
}

func (l *Log) Error(args ...interface{}){
	l.withFileField().Error(args...)
}

func (l *Log) Errorf(format string, args ...interface{}){
	l.withFileField().Errorf(format, args...)
}

func (l *Log) Fatal(args ...interface{}){
	l.withFileField().Fatal(args...)
}

func (l *Log) Fatalf(format string, args ...interface{}){
	l.withFileField().Fatalf(format, args...)
}

func (l *Log) Logf(lvl logrus.Level, format string, args ...interface{}){
	l.logger.Logf(lvl, format, args...)
}

func (l *Log) withFileField() *logrus.Entry {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return l.logger
	}
	file = file[strings.LastIndex(file, "/")+1:]
	return l.logger.WithField("file", fmt.Sprintf("%s:%d", file, line))
}

func (l *Log) WithField(name string, value interface{}) *Log {
	return &Log{
		logger: l.logger.WithField(name, value),
	}
}

var logger = newLog()

func newLog() *Log {
	var l = &logrus.Logger{
		Out:   os.Stderr,
		Formatter: &formatter.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			HideKeys:        true,
			ShowFullLevel:   true,
			TrimMessages:    true,
		},
		Level: logrus.InfoLevel,
	}
	lvl := os.Getenv("LOG_LEVEL")
	level, err := parseLevel(lvl)
	if err != nil {
		l.SetLevel(logrus.InfoLevel)
	} else {
		l.SetLevel(level)
	}
	return &Log{
		logger: logrus.NewEntry(l),
	}
}


func parseLevel(lvl string) (level logrus.Level, err error) {
	if len(lvl) == 0 {
		return logrus.InfoLevel, nil
	}
	switch strings.ToLower(lvl) {
	case "debug":
		level = logrus.DebugLevel
	case "info":
		level = logrus.InfoLevel
	case "warning":
		level = logrus.WarnLevel
	case "error":
		level = logrus.ErrorLevel
	case "fatal":
		level = logrus.FatalLevel
	default:
		err = fmt.Errorf("invalid log level: %s", lvl)
	}

	return
}

func DefaultLogger() *Log {
	return logger
}
