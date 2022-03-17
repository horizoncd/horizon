package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/pipelinerun/models"
)

type DAO interface {
	// Create create a pipelinerun
	Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error)
	GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error)
	GetByClusterID(ctx context.Context, clusterID uint,
		canRollback bool, query q.Query) (int, []*models.Pipelinerun, error)
	// DeleteByID delete pipelinerun by id
	DeleteByID(ctx context.Context, pipelinerunID uint) error
	UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error
	GetLatestByClusterIDAndAction(ctx context.Context, clusterID uint, action string) (*models.Pipelinerun, error)
	UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error
	GetLatestSuccessByClusterID(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
	GetFirstCanRollbackPipelinerun(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
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

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.PipelinerunInDB, result.Error.Error())
	}

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
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
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

	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.PipelinerunInDB, result.Error.Error())
	}

	return result.Error
}

func (d *dao) UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Exec(common.PipelinerunUpdateConfigCommitByID, commit, pipelinerunID)

	if result.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	return result.Error
}

func (d *dao) GetLatestByClusterIDAndAction(ctx context.Context,
	clusterID uint, action string) (*models.Pipelinerun, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var pipelinerun models.Pipelinerun
	result := db.Raw(common.PipelinerunGetLatestByClusterIDAndAction, clusterID, action).Scan(&pipelinerun)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}

func (d *dao) GetLatestSuccessByClusterID(ctx context.Context, clusterID uint) (*models.Pipelinerun, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var pipelinerun models.Pipelinerun
	result := db.Raw(common.PipelinerunGetLatestSuccessByClusterID, clusterID).Scan(&pipelinerun)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}

func (d *dao) UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	res := db.Exec(common.PipelinerunUpdateResultByID, result.Result, result.S3Bucket,
		result.LogObject, result.PrObject, result.StartedAt, result.FinishedAt, pipelinerunID)

	if res.Error != nil {
		return herrors.NewErrGetFailed(herrors.PipelinerunInDB, res.Error.Error())
	}
	return res.Error
}

func (d *dao) GetByClusterID(ctx context.Context, clusterID uint,
	canRollback bool, query q.Query) (int, []*models.Pipelinerun, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, nil, err
	}
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
	result := db.Raw(queryScript,
		clusterID, limit, offset).Scan(&pipelineruns)
	if result.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	var total int
	result = db.Raw(countScript,
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
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var pipelinerun models.Pipelinerun
	result := db.Raw(common.PipelinerunGetFirstCanRollbackByClusterID, clusterID).Scan(&pipelinerun)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}
