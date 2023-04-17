package wlog

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/util/log"
)

func TestLogOK(_ *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")

	const op = "app: create application"
	defer Start(ctx, op).StopPrint()
	log.Info(ctx, "hello world")

	Start(ctx, "test: stopPrint").StopPrint()
}

func TestPanic(_ *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")

	const op = "app: create application"
	defer Start(ctx, op).StopPrint()

	doPanic()
}

func TestPanic2(_ *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")

	const op = "app: create application"
	defer Start(ctx, op).StopPrint()

	doPanic()
}

func doPanic() {
	var v *int
	*v = 10
}

func TestLogError(_ *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")

	const op = "app: create application"
	defer Start(ctx, op).StopPrint()

	// err = errors.New("unknown error")
	log.Info(ctx, "hello world")
}

func TestResponse(_ *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")
	resp := &http.Response{
		Body: ioutil.NopCloser(strings.NewReader("123")),
	}
	common.Response(ctx, resp)
}
