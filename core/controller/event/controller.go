package event

import (
	"context"

	eventmanager "g.hz.netease.com/horizon/pkg/event/manager"
	eventmodels "g.hz.netease.com/horizon/pkg/event/models"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	ListEventActions(ctx context.Context) map[string][]string
}

type controller struct {
	eventMgr eventmanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		eventMgr: param.EventManager,
	}
}

func (c *controller) ListEventActions(ctx context.Context) map[string][]string {
	const op = "event controller: list event actions"
	defer wlog.Start(ctx, op).StopPrint()

	eventActions := map[string][]string{
		string(eventmodels.Application): {
			string(eventmodels.Created),
			string(eventmodels.Deleted),
			string(eventmodels.Transferred),
		},
		string(eventmodels.Cluster): {
			string(eventmodels.Created),
			string(eventmodels.Deleted),
			string(eventmodels.BuildDeployed),
			string(eventmodels.Deployed),
			string(eventmodels.Rollbacked),
			string(eventmodels.Freed),
		},
	}
	return eventActions
}
