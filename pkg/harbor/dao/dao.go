package dao

import (
	"context"

	he "g.hz.netease.com/horizon/core/errors"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/harbor/models"
)

type DAO interface {
	// Create a harbor
	Create(ctx context.Context, harbor *models.Harbor) (*models.Harbor, error)
	// GetByID get by id
	GetByID(ctx context.Context, id uint) (*models.Harbor, error)
	// ListAll list all harbors
	ListAll(ctx context.Context) ([]*models.Harbor, error)
}

type dao struct{}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

func (d *dao) Create(ctx context.Context, harbor *models.Harbor) (*models.Harbor, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(harbor)

	if result.Error != nil {
		return nil, he.NewErrGetFailed(he.HarborInDB, result.Error.Error())
	}

	return harbor, result.Error
}

func (d *dao) GetByID(ctx context.Context, id uint) (*models.Harbor, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var harbor models.Harbor
	result := db.Raw(common.HarborGetByID, id).First(&harbor)

	if result.Error != nil {
		return nil, he.NewErrGetFailed(he.HarborInDB, result.Error.Error())
	}

	return &harbor, result.Error
}

func (d *dao) ListAll(ctx context.Context) ([]*models.Harbor, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var harbors []*models.Harbor
	result := db.Raw(common.HarborListAll).Scan(&harbors)

	if result.Error != nil {
		return nil, he.NewErrGetFailed(he.HarborInDB, result.Error.Error())
	}

	return harbors, result.Error
}
