package store

import (
	goerrors "errors"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/oauth/models"
	"golang.org/x/net/context"

	"gorm.io/gorm"
)

type DbTokenStore struct {
	db *gorm.DB
}

var _ TokenStore = &DbTokenStore{}

func (d *DbTokenStore) Create(ctx context.Context, token *models.Token) error {
	result := d.db.Create(token)
	return result.Error
}

func (d *DbTokenStore) Get(ctx context.Context, code string) (*models.Token, error) {
	var token models.Token
	result := d.db.Raw(common.TokenGetByCode, code).First(&token)
	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.TokenInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.TokenInDB, result.Error.Error())
	}
	return &token, nil
}

func (d *DbTokenStore) DeleteByCode(ctx context.Context, code string) error {
	result := d.db.Exec(common.DeleteByCode, code)
	return result.Error
}

func (d *DbTokenStore) DeleteByClientID(ctx context.Context, clientID string) error {
	result := d.db.Exec(common.DeleteByClientID, clientID)
	return result.Error
}
