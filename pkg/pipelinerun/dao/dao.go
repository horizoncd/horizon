package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"gorm.io/gorm"
)

type DAO interface {
	// Create create a pipelinerun
	Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error)
	GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error)
	GetByClusterID(ctx context.Context, clusterID uint,
		canRollback bool, query q.Query) (int, []*models.Pipelinerun, error)
	// DeleteByID delete pipelinerun by id
	DeleteByID(ctx context.Context, pipelinerunID uint) error
	DeleteByClusterID(ctx context.Context, clusterID uint) error
	UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error
	GetLatestByClusterIDAndAction(ctx context.Context, clusterID uint, action string) (*models.Pipelinerun, error)
	GetLatestByClusterIDAndActionAndStatus(ctx context.Context, clusterID uint,
		action, status string) (*models.Pipelinerun, error)
	UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error
	GetLatestSuccessByClusterID(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
	GetFirstCanRollbackPipelinerun(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
}

type dao struct{ db *gorm.DB }

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d *dao) Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error) {
	result := d.db.WithContext(ctx).Create(pipelinerun)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.PipelinerunInDB, result.Error.Error())
	}

	return pipelinerun, result.Error
}

func (d *dao) GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error) {
	var pr models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetByID, pipelinerunID).Scan(&pr)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pr, nil
}

func (d *dao) DeleteByClusterID(ctx context.Context, clusterID uint) error {
	result := d.db.WithContext(ctx).Exec(common.PipelinerunDeleteByClusterID, clusterID)

	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.PipelinerunInDB, result.Error.Error())
	}

	return result.Error
}

func (d *dao) DeleteByID(ctx context.Context, pipelinerunID uint) error {
	result := d.db.WithContext(ctx).Exec(common.PipelinerunDeleteByID, pipelinerunID)

	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.PipelinerunInDB, result.Error.Error())
	}

	return result.Error
}

func (d *dao) UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error {
	result := d.db.WithContext(ctx).Exec(common.PipelinerunUpdateConfigCommitByID, commit, pipelinerunID)

	if result.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	return result.Error
}

func (d *dao) GetLatestByClusterIDAndAction(ctx context.Context,
	clusterID uint, action string) (*models.Pipelinerun, error) {
	var pipelinerun models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetLatestByClusterIDAndAction,
		clusterID, action).Scan(&pipelinerun)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}

func (d *dao) GetLatestByClusterIDAndActionAndStatus(ctx context.Context,
	clusterID uint, action string, status string) (*models.Pipelinerun, error) {
	var pipelinerun models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetLatestByClusterIDAndActionAndStatus, clusterID,
		action, status).Scan(&pipelinerun)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}

func (d *dao) GetLatestSuccessByClusterID(ctx context.Context, clusterID uint) (*models.Pipelinerun, error) {
	var pipelinerun models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetLatestSuccessByClusterID, clusterID).Scan(&pipelinerun)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}

func (d *dao) UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error {
	res := d.db.WithContext(ctx).Exec(common.PipelinerunUpdateResultByID, result.Result, result.S3Bucket,
		result.LogObject, result.PrObject, result.StartedAt, result.FinishedAt, pipelinerunID)

	if res.Error != nil {
		return herrors.NewErrGetFailed(herrors.PipelinerunInDB, res.Error.Error())
	}
	return res.Error
}

func (d *dao) GetByClusterID(ctx context.Context, clusterID uint,
	canRollback bool, query q.Query) (int, []*models.Pipelinerun, error) {
	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize

	var pipelineruns []*models.Pipelinerun
	queryScript := common.PipelinerunGetByClusterID
	countScript := common.PipelinerunGetByClusterIDTotalCount
	if canRollback {
		// remove the first canRollback pipelinerun
		offset++
		queryScript = common.PipelinerunCanRollbackGetByClusterID
		countScript = common.PipelinerunCanRollbackGetByClusterIDTotalCount
	}
	result := d.db.WithContext(ctx).Raw(queryScript,
		clusterID, limit, offset).Scan(&pipelineruns)
	if result.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	var total int
	result = d.db.WithContext(ctx).Raw(countScript,
		clusterID).Scan(&total)

	if total < 0 {
		total = 0
	}

	if result.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}

	return total, pipelineruns, result.Error
}

func (d *dao) GetFirstCanRollbackPipelinerun(ctx context.Context, clusterID uint) (*models.Pipelinerun, error) {
	var pipelinerun models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetFirstCanRollbackByClusterID, clusterID).Scan(&pipelinerun)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}
