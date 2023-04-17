package dao

import (
	"context"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/common"
	"github.com/horizoncd/horizon/pkg/pipelinerun/models"
	"gorm.io/gorm"
)

type DAO interface {
	// Create creates a Pipelinerun object.
	Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error)
	// GetByID gets a Pipelinerun object by its ID.
	GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error)
	// GetByCIEventID gets a Pipelinerun object by its CI event ID.
	GetByCIEventID(ctx context.Context, ciEventID string) (*models.Pipelinerun, error)
	// GetByClusterID gets a specified number of Pipelinerun objects by their cluster ID.
	GetByClusterID(ctx context.Context, clusterID uint,
		canRollback bool, query q.Query) (int, []*models.Pipelinerun, error)
	// DeleteByID deletes a Pipelinerun object by its ID.
	DeleteByID(ctx context.Context, pipelinerunID uint) error
	// DeleteByClusterID deletes a Pipelinerun object by its cluster ID.
	DeleteByClusterID(ctx context.Context, clusterID uint) error
	// UpdateConfigCommitByID updates the configuration of a Pipelinerun object by its ID.
	UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error
	// GetLatestByClusterIDAndAction gets the latest Pipelinerun object by its cluster ID and action.
	GetLatestByClusterIDAndAction(ctx context.Context, clusterID uint, action string) (*models.Pipelinerun, error)
	// GetLatestByClusterIDAndActionAndStatus gets the latest Pipelinerun object by its cluster ID, action, and status.
	GetLatestByClusterIDAndActionAndStatus(ctx context.Context,
		clusterID uint, action, status string) (*models.Pipelinerun, error)
	// UpdateStatusByID updates the status of a Pipelinerun object by its ID.
	UpdateStatusByID(ctx context.Context, pipelinerunID uint, result models.PipelineStatus) error
	// UpdateCIEventIDByID updates the CI event ID of a Pipelinerun object by its ID.
	UpdateCIEventIDByID(ctx context.Context, pipelinerunID uint, ciEventID string) error
	// UpdateResultByID updates the result of a Pipelinerun object by its ID.
	UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error
	// GetLatestSuccessByClusterID gets the latest successful Pipelinerun object by its cluster ID.
	GetLatestSuccessByClusterID(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
	// GetFirstCanRollbackPipelinerun gets the first Pipelinerun object that can be rolled back.
	GetFirstCanRollbackPipelinerun(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
}

type dao struct {
	db *gorm.DB
}

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

func (d *dao) GetByCIEventID(ctx context.Context, ciEventID string) (*models.Pipelinerun, error) {
	var pr models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetByCIEventID, ciEventID).Scan(&pr)
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
	clusterID uint, action string,
) (*models.Pipelinerun, error) {
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
	clusterID uint, action string, status string,
) (*models.Pipelinerun, error) {
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

func (d *dao) UpdateStatusByID(ctx context.Context, pipelinerunID uint, status models.PipelineStatus) error {
	res := d.db.WithContext(ctx).Exec(common.PipelinerunUpdateStatusByID, status, pipelinerunID)
	if res.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.PipelinerunInDB, res.Error.Error())
	}
	return res.Error
}

func (d *dao) UpdateCIEventIDByID(ctx context.Context, pipelinerunID uint, ciEventID string) error {
	res := d.db.WithContext(ctx).Exec(common.PipelinerunUpdateCIEventIDByID, ciEventID, pipelinerunID)
	if res.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.PipelinerunInDB, res.Error.Error())
	}
	return res.Error
}

func (d *dao) UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error {
	res := d.db.WithContext(ctx).Exec(common.PipelinerunUpdateResultByID, result.Result, result.S3Bucket,
		result.LogObject, result.PrObject, result.StartedAt, result.FinishedAt, pipelinerunID)

	if res.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.PipelinerunInDB, res.Error.Error())
	}
	return res.Error
}

func (d *dao) GetByClusterID(ctx context.Context, clusterID uint,
	canRollback bool, query q.Query,
) (int, []*models.Pipelinerun, error) {
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
