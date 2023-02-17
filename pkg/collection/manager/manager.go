package manager

import (
	"context"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/collection/dao"
	"github.com/horizoncd/horizon/pkg/collection/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"gorm.io/gorm"
)

type Manager interface {
	Create(ctx context.Context, collection *models.Collection) (*models.Collection, error)
	DeleteByResource(ctx context.Context, userID uint, resourceID uint,
		resourceType models.CollectionResourceType) (*models.Collection, error)
	List(ctx context.Context, resourceType models.CollectionResourceType, ids []uint) ([]models.Collection, error)
}

type manager struct {
	dao dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}

func (m *manager) Create(ctx context.Context, collection *models.Collection) (*models.Collection, error) {
	_, err := m.dao.GetByResource(ctx, collection.UserID, collection.ResourceID, collection.ResourceType)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			return m.dao.Create(ctx, collection)
		}
		return nil, err
	}
	return collection, nil
}

func (m *manager) DeleteByResource(ctx context.Context, userID uint, resourceID uint,
	resourceType models.CollectionResourceType) (*models.Collection, error) {
	_, err := m.dao.GetByResource(ctx, userID, resourceID, resourceType)
	if err != nil {
		return nil, err
	}
	return m.dao.DeleteByResource(ctx, userID, resourceID, resourceType)
}

func (m *manager) List(ctx context.Context, resourceType models.CollectionResourceType,
	ids []uint) ([]models.Collection, error) {
	return m.dao.List(ctx, resourceType, ids)
}
