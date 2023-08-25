package manager

import (
	"context"

	"gorm.io/gorm"

	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/pr/dao"
	"github.com/horizoncd/horizon/pkg/pr/models"
)

type PRMessageManager interface {
	// Create create a PRMessage
	Create(ctx context.Context, prMessage *models.PRMessage) (*models.PRMessage, error)
	// List list PRMessage order by created_at asc
	List(ctx context.Context, pipelineRunID uint, query *q.Query) (int, []*models.PRMessage, error)
}

type prMessageManager struct {
	dao dao.PRMessageDAO
}

func NewPRMessageManager(db *gorm.DB) PRMessageManager {
	return &prMessageManager{
		dao: dao.NewPRMessageDAO(db),
	}
}

func (m *prMessageManager) Create(ctx context.Context, prMessage *models.PRMessage) (*models.PRMessage, error) {
	return m.dao.Create(ctx, prMessage)
}

func (m *prMessageManager) List(ctx context.Context, pipelineRunID uint,
	query *q.Query) (int, []*models.PRMessage, error) {
	return m.dao.List(ctx, pipelineRunID, query)
}
