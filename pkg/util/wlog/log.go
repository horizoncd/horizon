package wlog

import (
	"context"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"time"

	"g.hz.netease.com/horizon/pkg/util/log"
)

type Log struct {
	ctx   context.Context
	start time.Time
}

func Start(ctx context.Context, op string) Log {
	log.Info(ctx, op)
	return Log{ctx: ctx, start: time.Now()}
}

func (l Log) Stop(end func() string) {
	if err := recover(); err != nil {
		log.Error(l.ctx, string(debug.Stack()))

		panic(err)
	}
	duration := time.Since(time.Now())
	log.WithFiled(l.ctx, "duration", duration).Info()
}

func ByErr(err error) string {
	if err == nil {
		return "successfully"
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
