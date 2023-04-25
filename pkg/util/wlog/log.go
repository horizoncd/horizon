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
	"runtime/debug"
	"time"

	"github.com/horizoncd/horizon/pkg/util/log"
)

type Log struct {
	ctx   context.Context
	start time.Time
	op    string
}

func Start(ctx context.Context, op string) Log {
	return Log{op: op, ctx: ctx, start: time.Now()}
}

func (l Log) StopPrint() {
	if err := recover(); err != nil {
		log.Error(l.ctx, string(debug.Stack()))
	}
	duration := time.Since(l.start)

	log.WithFiled(l.ctx, "op",
		l.op).WithField("duration", duration.String()).Info("")
}

func (l Log) GetDuration() time.Duration {
	return time.Since(l.start)
}
