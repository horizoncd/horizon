package wlog

import (
	"context"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"time"

	"g.hz.netease.com/horizon/pkg/util/log"
)

const (
	Success string = "successfully"
)

type Log struct {
	ctx           context.Context
	start         time.Time
	op            string
	ignoredErrors []error
}

func Start(ctx context.Context, op string) Log {
	return Log{op: op, ctx: ctx, start: time.Now()}
}

func (l Log) Exclude(ignoredErrors ...error) Log {
	l.ignoredErrors = ignoredErrors
	return l
}

func (l Log) Stop(err error) {
	if err := recover(); err != nil {
		log.Error(l.ctx, string(debug.Stack()))
	}
	duration := time.Since(l.start)

	if err == nil {
		log.WithFiled(l.ctx, "op",
			l.op).WithField("duration", duration.String()).Info(Success) // nolint
	} else {
		for _, ignoredError := range l.ignoredErrors {
			if err == ignoredError {
				log.WithFiled(l.ctx, "op", l.op).
					WithField("duration", duration.String()).Info(err.Error())
				return
			}
		}
		log.WithFiled(l.ctx, "op",
			l.op).WithField("duration", duration.String()).Errorf(err.Error()) // nolint
	}
}

func (l Log) StopPrint() {
	if err := recover(); err != nil {
		log.Error(l.ctx, string(debug.Stack()))
	}
	duration := time.Since(l.start)

	log.WithFiled(l.ctx, "op",
		l.op).WithField("duration", duration).Info("")
}

func (l Log) GetDuration() time.Duration {
	return time.Since(l.start)
}

func Response(ctx context.Context, resp *http.Response) string {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(ctx, err)
		return err.Error()
	}

	str := string(data)
	log.Info(ctx, str)
	return str
}

func ResponseContent(ctx context.Context, data []byte) {
	log.Info(ctx, string(data))
}
