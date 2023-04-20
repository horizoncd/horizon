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
