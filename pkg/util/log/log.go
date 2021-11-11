package log

import (
	"context"
	"fmt"
	"runtime"
	"strconv"

	rlog "github.com/sirupsen/logrus"
)

type key string

const (
	traceKey = key("trace")
)
const (
	fileKey = "File"
)

func Key() string {
	return string(traceKey)
}

func getCallerInfo() (fileline string) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		panic("Could not get context info for logger!")
	}
	filename := file + ":" + strconv.Itoa(line)

	return filename
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
	filename := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).Info(args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).Infof(format, args...)
}

func Error(ctx context.Context, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).Error(args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).Errorf(format, args...)
}

func Warning(ctx context.Context, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).Warning(args...)
}

func Warningf(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).Warningf(format, args...)
}

func Debug(ctx context.Context, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).Debug(args...)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).Debugf(format, args...)
}
