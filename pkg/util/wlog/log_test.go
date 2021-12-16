package wlog

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/stretchr/testify/assert"
)

func TestLogOK(t *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")

	var err error

	const op = "app: create application"
	defer Start(ctx, op).Stop(func() string { return ByErr(err) })
	log.Info(ctx, "hello world")

	Start(ctx, "test: stopPrint").StopPrint()
}

func TestPanic(t *testing.T) {
	var err error
	ctx := log.WithContext(context.Background(), "traceId")

	defer func() {
		err := recover()
		assert.NotNil(t, err)
		log.Error(ctx, err)
	}()

	const op = "app: create application"
	defer Start(ctx, op).Stop(func() string { return ByErr(err) })

	doPanic()
}

func doPanic() {
	var v *int
	*v = 10
}

func TestLogError(t *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")

	var err error

	const op = "app: create application"
	defer Start(ctx, op).Stop(func() string { return ByErr(err) })

	err = errors.New("unknown error")
	log.Info(ctx, "hello world")
}

func TestResponse(t *testing.T) {
	ctx := log.WithContext(context.Background(), "traceId")
	resp := &http.Response{
		Body: ioutil.NopCloser(strings.NewReader("123")),
	}
	Response(ctx, resp)
}
