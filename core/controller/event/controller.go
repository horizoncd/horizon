package event

import (
	"context"

	eventmanager "github.com/horizoncd/horizon/pkg/event/manager"
	"github.com/horizoncd/horizon/pkg/log/wlog"
	"github.com/horizoncd/horizon/pkg/param"
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
