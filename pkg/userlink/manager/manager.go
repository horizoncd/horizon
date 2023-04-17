package manager

import (
	"context"

	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/idp/utils"
	"github.com/horizoncd/horizon/pkg/userlink/dao"
	"github.com/horizoncd/horizon/pkg/userlink/models"
	"gorm.io/gorm"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/userlink/manager/manager_mock.go -package=mock_manager
type Manager interface {
	CreateLink(ctx context.Context, uid uint, idpID uint, claims *utils.Claims, deletable bool) (*models.UserLink, error)
	ListByUserID(ctx context.Context, uid uint) ([]*models.UserLink, error)
	GetByID(ctx context.Context, id uint) (*models.UserLink, error)
	GetByIDPAndSub(ctx context.Context, idpID uint, sub string) (*models.UserLink, error)
	DeleteByID(ctx context.Context, id uint) error
}

type manager struct {
	dao dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{dao: dao.NewDAO(db)}
}

func (m *manager) CreateLink(ctx context.Context, uid uint,
	idpID uint, claims *utils.Claims, deletable bool,
) (*models.UserLink, error) {
	if claims == nil {
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "claims is required")
	}
	link := models.UserLink{
		Sub:       claims.Sub,
		IdpID:     idpID,
		UserID:    uid,
		Name:      claims.Name,
		Email:     claims.Email,
		Deletable: deletable,
	}
	return m.dao.CreateLink(ctx, &link)
}

func (m *manager) ListByUserID(ctx context.Context, uid uint) ([]*models.UserLink, error) {
	return m.dao.ListByUserID(ctx, uid)
}

func (m *manager) GetByID(ctx context.Context, id uint) (*models.UserLink, error) {
	return m.dao.GetByID(ctx, id)
}

func (m *manager) GetByIDPAndSub(ctx context.Context,
	idpID uint, sub string,
) (*models.UserLink, error) {
	return m.dao.GetByIDPAndSub(ctx, idpID, sub)
}

func (m *manager) DeleteByID(ctx context.Context, id uint) error {
	return m.dao.DeleteByID(ctx, id)
}
