package log

import (
	"context"
	"fmt"

	rlog "github.com/sirupsen/logrus"
)

type key string

const (
	traceKey = key("trace")
)

func Key() string {
	return string(traceKey)
}

func WithContext(parent context.Context, traceID string) context.Context {
	return context.WithValue(parent, Key(), fmt.Sprintf("%v", traceID)) // nolint
}

func WithFiled(ctx context.Context, key string, value interface{}) *rlog.Entry {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	return rlog.WithField(string(traceKey), val).WithField(key, value)
}

func Info(ctx context.Context, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	rlog.WithField(string(traceKey), val).Info(args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	rlog.WithField(string(traceKey), val).Infof(format, args...)
}

func Error(ctx context.Context, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	rlog.WithField(string(traceKey), val).Error(args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	rlog.WithField(string(traceKey), val).Errorf(format, args...)
}

func Warning(ctx context.Context, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	rlog.WithField(string(traceKey), val).Warning(args...)
}

func Warningf(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	rlog.WithField(string(traceKey), val).Warningf(format, args...)
}

func Debug(ctx context.Context, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	rlog.WithField(string(traceKey), val).Debug(args...)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	rlog.WithField(string(traceKey), val).Debugf(format, args...)
}
