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
	ctx   context.Context
	start time.Time
	op    string
}

func Start(ctx context.Context, op string) Log {
	return Log{op: op, ctx: ctx, start: time.Now()}
}

func (l Log) Stop(end func() string) {
	if err := recover(); err != nil {
		log.Error(l.ctx, string(debug.Stack()))

		panic(err)
	}
	duration := time.Since(l.start)

	str := end()
	if str == Success {
		log.WithFiled(l.ctx, "op", l.op).WithField("duration", duration).Info(Success)
	} else {
		log.WithFiled(l.ctx, "op", l.op).WithField("duration", duration).Errorf(end())
	}
}

func ByErr(err error) string {
	if err == nil {
		return Success
	}
	return err.Error()
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
