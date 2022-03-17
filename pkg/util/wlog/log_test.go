package wlog

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/pkg/util/log"
)

func TestLogOK(t *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")

	const op = "app: create application"
	defer Start(ctx, op).StopPrint()
	log.Info(ctx, "hello world")

	Start(ctx, "test: stopPrint").StopPrint()
}

func TestPanic(t *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")

	const op = "app: create application"
	defer Start(ctx, op).StopPrint()

	doPanic()
}

func TestPanic2(t *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")

	const op = "app: create application"
	defer Start(ctx, op).StopPrint()

	doPanic()
}

func doPanic() {
	var v *int
	*v = 10
}

func TestLogError(t *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")

	const op = "app: create application"
	defer Start(ctx, op).StopPrint()

	// err = errors.New("unknown error")
	log.Info(ctx, "hello world")
}

func TestResponse(t *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")
	resp := &http.Response{
		Body: ioutil.NopCloser(strings.NewReader("123")),
	}
	common.Response(ctx, resp)
}
