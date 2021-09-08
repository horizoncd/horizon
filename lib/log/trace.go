package log

import (
	"context"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type tracer struct{}

var TRACE = &tracer{}

func (t *tracer) Debug(ctx context.Context) func(f func() error) {
	return t.trace(ctx, logrus.DebugLevel)
}

func (t *tracer) Info(ctx context.Context) func(f func() error) {
	return t.trace(ctx, logrus.InfoLevel)
}

func (t *tracer) trace(ctx context.Context, l logrus.Level) func(f func() error) {
	logger := GetLogger(ctx)

	if err := recover(); err != nil {
		logger.Error(string(debug.Stack()))
		panic(err)
	}

	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return func(func() error) {
			logger.Errorf("exit with no caller found")
		}
	}
	fn := runtime.FuncForPC(pc)
	name := fn.Name()
	file = file[strings.LastIndex(file, "/")+1:]
	logger = logger.WithField("file", file+":"+strconv.Itoa(line))
	logger.Logf(l, "enter: %v", name)
	return func(f func() error) {
		err := f()
		if err == nil {
			logger.Logf(l, "exit: %v successfully", name)
		} else {
			logger.Errorf("exit: %v with error: %v", name, err)
		}
	}
}
