package dao

import (
	"context"
	goerrors "errors"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/common"
	"github.com/horizoncd/horizon/pkg/token/models"
	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, token *models.Token) (*models.Token, error)
	GetByID(ctx context.Context, id uint) (*models.Token, error)
	GetByCode(ctx context.Context, code string) (*models.Token, error)
	DeleteByID(ctx context.Context, id uint) error
	DeleteByCode(ctx context.Context, code string) error
	DeleteByClientID(ctx context.Context, clientID string) error
}

type dao struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d *dao) Create(ctx context.Context, token *models.Token) (*models.Token, error) {
	result := d.db.WithContext(ctx).Create(token)
	return token, result.Error
}

func (d *dao) GetByID(ctx context.Context, id uint) (*models.Token, error) {
	var token models.Token
	result := d.db.WithContext(ctx).Model(token).Where("id = ?", id).First(&token)
	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.TokenInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.TokenInDB, result.Error.Error())
	}
	return &token, nil
}

func (d *dao) GetByCode(ctx context.Context, code string) (*models.Token, error) {
	var token models.Token
	result := d.db.WithContext(ctx).Model(token).Where("code = ?", code).First(&token)
	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.TokenInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.TokenInDB, result.Error.Error())
	}
	return &token, nil
}

func (d *dao) DeleteByID(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteTokenByID, id)
	return result.Error
}

func (d *dao) DeleteByCode(ctx context.Context, code string) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteByCode, code)
	return result.Error
}

func (d *dao) DeleteByClientID(ctx context.Context, clientID string) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteByClientID, clientID)
	return result.Error
}
