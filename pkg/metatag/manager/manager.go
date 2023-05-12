package manager

import (
	"context"
	"github.com/horizoncd/horizon/pkg/metatag/dao"
	"github.com/horizoncd/horizon/pkg/metatag/models"
	"gorm.io/gorm"
)

type Manager interface {
	CreateMetatags(ctx context.Context, metatags []*models.Metatag) error
	GetMetatagKeys(ctx context.Context) ([]string, error)
	GetMetatagsByKey(ctx context.Context, key string) ([]*models.Metatag, error)
}

type manager struct {
	dao dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{dao: dao.NewDAO(db)}
}

func (s *manager) CreateMetatags(ctx context.Context, metatags []*models.Metatag) error {
	return s.dao.CreateMetatags(ctx, metatags)
}

func (s *manager) GetMetatagKeys(ctx context.Context) ([]string, error) {
	return s.dao.GetMetatagKeys(ctx)
}

func (s *manager) GetMetatagsByKey(ctx context.Context, key string) ([]*models.Metatag, error) {
	return s.dao.GetMetatagsByKey(ctx, key)
}
