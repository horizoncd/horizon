package slo

import (
	"context"
	"errors"
	"fmt"

	"g.hz.netease.com/horizon/pkg/config/grafana"
	perrors "g.hz.netease.com/horizon/pkg/errors"
)

var (
	ErrInternalInValid = errors.New("internal not valid")
)

type Controller interface {
	GetOverviewDashboard(ctx context.Context, env string) string
	GetAPIDashboard(ctx context.Context, internal string) (string, error)
	GetPipelineDashboard(ctx context.Context, internal string, env string) (string, error)
}

type controller struct {
	GrafanaSLO grafana.SLO
}

func NewController(grafanaSLO grafana.SLO) Controller {
	return &controller{GrafanaSLO: grafanaSLO}
}

func (c *controller) GetOverviewDashboard(_ context.Context, env string) string {
	// http://grafana.mockserver.org/d/Wz3GSBank/slo-gai-lan?orgId=1&kiosk&theme=light&var-env=test
	return fmt.Sprintf(c.GrafanaSLO.OverviewDashboard, env)
}

func (c *controller) GetAPIDashboard(_ context.Context, internal string) (string, error) {
	availability, ok := c.GrafanaSLO.Availability[internal]
	if ok {
		// http://grafana.mockserver.org/d/tKjaD1-nk/horizon-api-slo?orgId=1&kiosk&theme=light
		// &var-datasource=default&var-range=%s&var-api_availability=%s&var-api_read_rt=%s&var-api_write_rt=%s
		return fmt.Sprintf(c.GrafanaSLO.APIDashboard, internal, availability.APIAvailability,
			c.GrafanaSLO.APIReadRT, c.GrafanaSLO.APIWriteRT), nil
	}
	return "", perrors.WithMessage(ErrInternalInValid, "param not valid")
}

func (c *controller) GetPipelineDashboard(_ context.Context, internal string, env string) (string, error) {
	availability, ok := c.GrafanaSLO.Availability[internal]
	if ok {
		// http://grafana.mockserver.org/d/g40XAtbnk/horizon-pipeline-slo?orgId=1&kiosk&theme=light&var-env=%s
		// &var-range=%s&var-git_availability=%s&var-git_rt=%s&var-image_availability=%s&var-image_rt=%s
		// &var-deploy_availability=0.99&var-deploy_rt=60&var-datasource=default
		return fmt.Sprintf(c.GrafanaSLO.PipelineDashboard, env, internal, availability.GitAvailability, c.GrafanaSLO.GitRT,
			availability.ImageAvailability, c.GrafanaSLO.ImageRT, availability.DeployAvailability, c.GrafanaSLO.DeployRT), nil
	}
	return "", perrors.WithMessage(ErrInternalInValid, "param not valid")
}
