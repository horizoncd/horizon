package slo

import (
	"context"
	"fmt"

	"g.hz.netease.com/horizon/pkg/config/grafana"
)

type Controller interface {
	GetOverviewDashboard(ctx context.Context, env string) string
	GetAPIDashboard(ctx context.Context) string
	GetPipelineDashboard(ctx context.Context, env string) string
}

type controller struct {
	GrafanaSLO grafana.SLO
}

func NewController(grafanaSLO grafana.SLO) Controller {
	return &controller{GrafanaSLO: grafanaSLO}
}

func (c *controller) GetOverviewDashboard(_ context.Context, env string) string {
	
	return fmt.Sprintf(c.GrafanaSLO.OverviewDashboard, env)
}

func (c *controller) GetAPIDashboard(_ context.Context) string {
	
	return c.GrafanaSLO.APIDashboard
}

func (c *controller) GetPipelineDashboard(_ context.Context, env string) string {
	
	return fmt.Sprintf(c.GrafanaSLO.PipelineDashboard, env)
}
