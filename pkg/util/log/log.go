// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"context"
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
	return context.WithValue(parent, Key(), traceID) // nolint
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
