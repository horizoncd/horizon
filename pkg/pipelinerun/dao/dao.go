package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/pipelinerun/models"
)

type DAO interface {
	// Create create a pipelinerun
	Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error)
	GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error)
	// DeleteByID delete pipelinerun by id
	DeleteByID(ctx context.Context, pipelinerunID uint) error
	UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error
	GetLatestByClusterID(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
}

type dao struct{}

func NewDAO() DAO {
	return &dao{}
}

func (d *dao) Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(pipelinerun)

	return pipelinerun, result.Error
}

func (d *dao) GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var pr models.Pipelinerun
	result := db.Raw(common.PipelinerunGetByID, pipelinerunID).Scan(&pr)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pr, nil
}

func (d *dao) DeleteByID(ctx context.Context, pipelinerunID uint) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Exec(common.PipelinerunDeleteByID, pipelinerunID)

	return result.Error
}

func (d *dao) UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Exec(common.PipelinerunUpdateConfigCommitByID, commit, pipelinerunID)

	return result.Error
}

func (d *dao) GetLatestByClusterID(ctx context.Context, clusterID uint) (*models.Pipelinerun, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var pipelinerun models.Pipelinerun
	result := db.Raw(common.PipelinerunGetLatestByClusterID, clusterID).Scan(&pipelinerun)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}
