package store

import (
	goerrors "errors"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/oauth/models"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

type DBOauthAppStore struct {
	db *gorm.DB
}

func NewOauthAppStore(db *gorm.DB) OauthAppStore {
	return &DBOauthAppStore{db: db}
}

var _ OauthAppStore = &DBOauthAppStore{}

func (d *DBOauthAppStore) CreateApp(ctx context.Context, client models.OauthApp) error {
	result := d.db.Save(&client)
	return result.Error
}
func (d *DBOauthAppStore) GetApp(ctx context.Context, clientID string) (*models.OauthApp, error) {
	var client models.OauthApp
	result := d.db.Raw(common.GetOauthAppByClientID, clientID).First(&client)
	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.OAuthInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.OAuthInDB, result.Error.Error())
	}
	return &client, nil
}
func (d *DBOauthAppStore) DeleteApp(ctx context.Context, clientID string) error {
	result := d.db.Exec(common.DeleteOauthAppByClientID, clientID)
	return result.Error
}

func (d *DBOauthAppStore) CreateSecret(ctx context.Context,
	secret *models.OauthClientSecret) (*models.OauthClientSecret, error) {
	if result := d.db.Save(secret); result.Error != nil {
		return nil, result.Error
	}
	return secret, nil
}
func (d *DBOauthAppStore) DeleteSecretByClientID(ctx context.Context, clientID string) error {
	result := d.db.Exec(common.DeleteClientSecretByClientID, clientID)
	return result.Error
}
func (d *DBOauthAppStore) DeleteSecret(ctx context.Context, clientID string, clientSecretID uint) error {
	result := d.db.Exec(common.DeleteClientSecret, clientID, clientSecretID)
	return result.Error
}
func (d *DBOauthAppStore) ListSecret(ctx context.Context, clientID string) ([]models.OauthClientSecret, error) {
	var secrets []models.OauthClientSecret
	result := d.db.Raw(common.ClientSecretSelectAll, clientID).Scan(&secrets)
	if result.Error != nil {
		return nil, result.Error
	}
	return secrets, nil
}
