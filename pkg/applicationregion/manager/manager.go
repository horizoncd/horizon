package manager

import (
	"context"

	"g.hz.netease.com/horizon/pkg/applicationregion/dao"
	"g.hz.netease.com/horizon/pkg/applicationregion/models"
	"gorm.io/gorm"
)

type Manager interface {
	// ListByApplicationID list applicationRegions by applicationID
	ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.ApplicationRegion, error)
	// UpsertByApplicationID upsert application regions
	UpsertByApplicationID(ctx context.Context, applicationID uint, applicationRegions []*models.ApplicationRegion) error
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.ApplicationRegion, error) {
	return m.dao.ListByApplicationID(ctx, applicationID)
}

func (m *manager) UpsertByApplicationID(ctx context.Context,
	applicationID uint, applicationRegions []*models.ApplicationRegion) error {
	return m.dao.UpsertByApplicationID(ctx, applicationID, applicationRegions)
}
