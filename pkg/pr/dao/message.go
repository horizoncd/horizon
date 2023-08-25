package dao

import (
	"context"

	"gorm.io/gorm"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/pr/models"
)

type PRMessageDAO interface {
	// Create create a PRMessage
	Create(ctx context.Context, prMessage *models.PRMessage) (*models.PRMessage, error)
	// List list PRMessage order by created_at asc
	List(ctx context.Context, pipelineRunID uint, query *q.Query) (int, []*models.PRMessage, error)
}

type prMessageDAO struct{ db *gorm.DB }

func NewPRMessageDAO(db *gorm.DB) PRMessageDAO {
	return &prMessageDAO{db: db}
}

func (d *prMessageDAO) Create(ctx context.Context, prMessage *models.PRMessage) (*models.PRMessage, error) {
	result := d.db.WithContext(ctx).Create(prMessage)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.PRMessageInDB, result.Error.Error())
	}

	return prMessage, result.Error
}

func (d *prMessageDAO) List(ctx context.Context, pipelineRunID uint, query *q.Query) (int, []*models.PRMessage, error) {
	var (
		total      int64
		prMessages []*models.PRMessage
	)

	sql := d.db.WithContext(ctx).Table("tb_pr_msg").
		Where("pipeline_run_id = ?", pipelineRunID).
		Order("created_at asc")
	sql.Count(&total)
	result := sql.Limit(query.Limit()).Offset(query.Offset()).Find(&prMessages)
	if result.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.PRMessageInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return 0, []*models.PRMessage{}, nil
	}
	return int(result.RowsAffected), prMessages, nil
}
