package slo

import (
	"context"
	"fmt"

	"g.hz.netease.com/horizon/pkg/config/grafana"
)

type Controller interface {
	GetDashboards(ctx context.Context, env string) *Dashboards
}

type controller struct {
	GrafanaSLO grafana.SLO
}

func NewController(grafanaSLO grafana.SLO) Controller {
	return &controller{GrafanaSLO: grafanaSLO}
}

func (c *controller) GetDashboards(_ context.Context, env string) *Dashboards {
	return &Dashboards{
		Overview: fmt.Sprintf(c.GrafanaSLO.OverviewDashboard, env),
		API:      c.GrafanaSLO.APIDashboard,
		Pipeline: fmt.Sprintf(c.GrafanaSLO.PipelineDashboard, env),
	}
}
