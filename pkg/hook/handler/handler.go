package handler

import (
	"g.hz.netease.com/horizon/pkg/cmdb"
	"g.hz.netease.com/horizon/pkg/hook"
)

type EventHandler interface {
	Process(event *hook.Event) error
}

func NewCMDBEventHandler(controller cmdb.Controller) EventHandler {
	return &CMDBEventHandler{ctl: controller}
}

type CMDBEventHandler struct {
	ctl cmdb.Controller
}

func (h *CMDBEventHandler) Process(event *hook.Event) error {
	return nil
}
