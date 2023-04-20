package wlog

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/horizoncd/horizon/pkg/log"
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
