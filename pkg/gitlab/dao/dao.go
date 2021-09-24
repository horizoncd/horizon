package dao

import (
	"context"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/gitlab/models"
)

type DAO interface {
	Create(ctx context.Context, template *models.Gitlab) (*models.Gitlab, error)
	List(ctx context.Context) ([]models.Gitlab, error)
	GetByName(ctx context.Context, name string) (*models.Gitlab, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(ctx context.Context, gitlab *models.Gitlab) (*models.Gitlab, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(gitlab)
	return gitlab, result.Error
}

func (d *dao) List(ctx context.Context) ([]models.Gitlab, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var gitlabs []models.Gitlab
	result := db.Raw(common.GitlabQuery).Scan(&gitlabs)
	if result.Error != nil {
		return nil, result.Error
	}
	return gitlabs, nil
}

func (d *dao) GetByName(ctx context.Context, name string) (*models.Gitlab, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var gitlab models.Gitlab
	result := db.Raw(common.GitlabQueryByName, name).Scan(&gitlab)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &gitlab, nil
}
