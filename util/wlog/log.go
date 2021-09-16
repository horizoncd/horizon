package wlog

import (
	"context"
	"io/ioutil"
	"net/http"
	"runtime/debug"

	"g.hz.netease.com/horizon/util/log"
)

type Log struct{ ctx context.Context }

func Start(ctx context.Context, op string) Log {
	log.InfoDepth(ctx, 2, op)
	return Log{ctx: ctx}
}

func (l Log) Stop(end func() string) {
	if err := recover(); err != nil {
		log.Error(l.ctx, string(debug.Stack()))

		panic(err)
	}

	log.InfoDepth(l.ctx, 2, end())
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
	log.InfoDepth(ctx, 2, str)
	return str
}

func ResponseContent(ctx context.Context, data []byte) {
	log.InfoDepth(ctx, 2, string(data))
}
