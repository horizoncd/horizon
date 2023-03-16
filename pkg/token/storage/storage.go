package storage

import (
	"context"
	goerrors "errors"

	"github.com/horizoncd/horizon/pkg/common"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	"github.com/horizoncd/horizon/pkg/token/models"
	"gorm.io/gorm"
)

type storage struct {
	db *gorm.DB
}

func NewStorage(db *gorm.DB) Storage {
	return &storage{db: db}
}

func (d *storage) Create(ctx context.Context, token *models.Token) (*models.Token, error) {
	result := d.db.WithContext(ctx).Create(token)
	return token, result.Error
}

func (d *storage) GetByID(ctx context.Context, id uint) (*models.Token, error) {
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

func (d *storage) GetByCode(ctx context.Context, code string) (*models.Token, error) {
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

func (d *storage) DeleteByID(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteTokenByID, id)
	return result.Error
}

func (d *storage) DeleteByCode(ctx context.Context, code string) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteByCode, code)
	return result.Error
}

func (d *storage) DeleteByClientID(ctx context.Context, clientID string) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteByClientID, clientID)
	return result.Error
}
