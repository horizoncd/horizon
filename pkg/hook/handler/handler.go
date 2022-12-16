package handler

import "github.com/horizoncd/horizon/pkg/hook/hook"

type EventHandler interface {
	Process(event *hook.EventCtx) error
}
