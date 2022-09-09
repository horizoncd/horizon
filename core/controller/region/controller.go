package region

import (
	"context"
	"fmt"

	"g.hz.netease.com/horizon/pkg/grafana"
	"g.hz.netease.com/horizon/pkg/param"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	"g.hz.netease.com/horizon/pkg/region/models"
)

const (
	GrafanaDatasourceCMNamePrefix = "grafana-datasource"
	PrometheusDatasourceType      = "prometheus"
)

type Controller interface {
	ListRegions(ctx context.Context) ([]*Region, error)
	Create(ctx context.Context, request *CreateRegionRequest) (uint, error)
	UpdateByID(ctx context.Context, id uint, request *UpdateRegionRequest) error
	DeleteByID(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*Region, error)
}

func NewController(param *param.Param) Controller {
	return &controller{
		regionMgr:  param.RegionMgr,
		grafanaSvc: param.GrafanaService,
	}
}

type controller struct {
	regionMgr  regionmanager.Manager
	grafanaSvc grafana.Service
}

func (c controller) GetByID(ctx context.Context, id uint) (*Region, error) {
	region, err := c.regionMgr.GetRegionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ofRegionEntity(region), nil
}

func (c controller) DeleteByID(ctx context.Context, id uint) error {
	region, err := c.GetByID(ctx, id)
	if err != nil {
		return err
	}
	err = c.grafanaSvc.DeletePrometheusDatasourceConfigMap(ctx, formatGrafanaDatasourceCMName(region.Name))
	if err != nil {
		return err
	}

	return c.regionMgr.DeleteByID(ctx, id)
}

func (c controller) UpdateByID(ctx context.Context, id uint, request *UpdateRegionRequest) error {
	err := c.grafanaSvc.UpdatePrometheusDatasourceConfigMap(ctx, formatGrafanaDatasourceCMName(request.Name),
		formatGrafanaDatasourceCMLabels(),
		&grafana.DataSource{
			Name: request.Name,
			Type: PrometheusDatasourceType,
			URL:  request.PrometheusURL,
		})
	if err != nil {
		return err
	}

	err = c.regionMgr.UpdateByID(ctx, id, &models.Region{
		DisplayName:   request.DisplayName,
		Server:        request.Server,
		Certificate:   request.Certificate,
		IngressDomain: request.IngressDomain,
		PrometheusURL: request.PrometheusURL,
		HarborID:      request.HarborID,
		Disabled:      request.Disabled,
	})
	if err != nil {
		err := c.grafanaSvc.DeletePrometheusDatasourceConfigMap(ctx, formatGrafanaDatasourceCMName(request.Name))
		if err != nil {
			return err
		}
		return err
	}

	return nil
}

func (c controller) Create(ctx context.Context, request *CreateRegionRequest) (uint, error) {
	err := c.grafanaSvc.CreatePrometheusDatasourceConfigMap(ctx, formatGrafanaDatasourceCMName(request.Name),
		formatGrafanaDatasourceCMLabels(),
		&grafana.DataSource{
			Name: request.Name,
			Type: PrometheusDatasourceType,
			URL:  request.PrometheusURL,
		})
	if err != nil {
		return 0, err
	}

	create, err := c.regionMgr.Create(ctx, &models.Region{
		Name:          request.Name,
		DisplayName:   request.DisplayName,
		Server:        request.Server,
		Certificate:   request.Certificate,
		IngressDomain: request.IngressDomain,
		PrometheusURL: request.PrometheusURL,
		HarborID:      request.HarborID,
	})
	if err != nil {
		err := c.grafanaSvc.DeletePrometheusDatasourceConfigMap(ctx, formatGrafanaDatasourceCMName(request.Name))
		if err != nil {
			return 0, err
		}
		return 0, err
	}

	return create.ID, nil
}

func (c controller) ListRegions(ctx context.Context) ([]*Region, error) {
	entities, err := c.regionMgr.ListRegionEntities(ctx)
	if err != nil {
		return nil, err
	}
	return ofRegionEntities(entities), nil
}

func formatGrafanaDatasourceCMName(name string) string {
	return fmt.Sprintf("%s-%s", GrafanaDatasourceCMNamePrefix, name)
}
func formatGrafanaDatasourceCMLabels() map[string]string {
	// todo read from config
	return map[string]string{}
}
