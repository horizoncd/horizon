package handler

import "g.hz.netease.com/horizon/pkg/hook/hook"

type EventHandler interface {
	Process(event *hook.EventCtx) error
}
