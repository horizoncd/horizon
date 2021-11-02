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
	// GetLastPipelinerunWithConfigCommit get last pipelinerun which config_commit is not empty
	GetLastPipelinerunWithConfigCommit(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
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

func (d *dao) GetLastPipelinerunWithConfigCommit(ctx context.Context, clusterID uint) (*models.Pipelinerun, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var pipelinerun models.Pipelinerun
	result := db.Raw(common.PipelineRunGetLastConfigCommit, clusterID).Scan(&pipelinerun)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}
