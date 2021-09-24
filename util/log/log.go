package log

import (
	"context"
	"fmt"

	"k8s.io/klog"
)

type key string

const (
	logKey = key("log")

	_depth = 1
)

func Key() string {
	return string(logKey)
}

func WithContext(parent context.Context, traceID string) context.Context {
	return context.WithValue(parent, Key(), fmt.Sprintf("[%v] ", traceID)) // nolint
}

func Info(ctx context.Context, args ...interface{}) {
	InfoDepth(ctx, _depth+1, args...)
}

func InfoDepth(ctx context.Context, depth int, args ...interface{}) {
	val, ok := ctx.Value(string(logKey)).(string)
	if !ok {
		klog.InfoDepth(depth, args...)
		return
	}

	msg := fmt.Sprint(args...)
	klog.InfoDepth(depth, val, msg)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(logKey)).(string)
	if !ok {
		msg := fmt.Sprintf(format, args...)
		klog.InfoDepth(_depth, msg)
		return
	}

	format = fmt.Sprintf("%v%v", val, format)
	msg := fmt.Sprintf(format, args...)
	klog.InfoDepth(_depth, msg)
}

func Error(ctx context.Context, args ...interface{}) {
	val, ok := ctx.Value(string(logKey)).(string)
	if !ok {
		klog.ErrorDepth(_depth, args...)
		return
	}

	msg := fmt.Sprint(args...)
	klog.ErrorDepth(_depth, val, msg)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	val, ok := ctx.Value(string(logKey)).(string)
	if !ok {
		msg := fmt.Sprintf(format, args...)
		klog.ErrorDepth(_depth, msg)
		return
	}

	format = fmt.Sprintf("%v%v", val, format)
	msg := fmt.Sprintf(format, args...)
	klog.ErrorDepth(_depth, msg)
}

func Flush() {
	klog.Flush()
}
