package log

import (
	"context"
)

const loggerKey = "logger"

func LoggerKey() string {
	return loggerKey
}

func GetLogger(ctx context.Context) *Log {
	if ctx == nil {
		return logger
	}

	l := ctx.Value(LoggerKey())
	if l == nil {
		return logger
	}

	return l.(*Log)
}
