package event

import (
	"context"

	eventmanager "g.hz.netease.com/horizon/pkg/event/manager"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	ListSupportEvents(ctx context.Context) map[string]string
}

type controller struct {
	eventMgr eventmanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		eventMgr: param.EventManager,
	}
}

func (c *controller) ListSupportEvents(ctx context.Context) map[string]string {
	const op = "event controller: list supported events"
	defer wlog.Start(ctx, op).StopPrint()

	return c.eventMgr.ListSupportEvents()
}
