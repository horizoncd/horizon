package log

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	rlog "github.com/sirupsen/logrus"
)

type key string

const (
	traceKey = key("trace")
)
const (
	fileKey = "File"
	funcKey = "Func"
)

func Key() string {
	return string(traceKey)
}

func getCallerInfo() (file, function string) {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		panic("Could not get context info for logger!")
	}
	filename := file + ":" + strconv.Itoa(line)
	funcname := runtime.FuncForPC(pc).Name()
	fn := funcname[strings.LastIndex(funcname, ".")+1:]
	return filename, fn
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
	filename, fn := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).WithField(funcKey, fn).Info(args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename, fn := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).WithField(funcKey, fn).Infof(format, args...)
}

func Error(ctx context.Context, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename, fn := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).WithField(funcKey, fn).Error(args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename, fn := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).WithField(funcKey, fn).Errorf(format, args...)
}

func Warning(ctx context.Context, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename, fn := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).WithField(funcKey, fn).Warning(args...)
}

func Warningf(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename, fn := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).WithField(funcKey, fn).Warningf(format, args...)
}

func Debug(ctx context.Context, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename, fn := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).WithField(funcKey, fn).Debug(args...)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(traceKey)).(string)
	if !ok {
		val = ""
	}
	filename, fn := getCallerInfo()
	rlog.WithField(string(traceKey), val).WithField(fileKey, filename).WithField(funcKey, fn).Debugf(format, args...)
}
