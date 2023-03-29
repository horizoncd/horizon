package store

import (
	goerrors "errors"

	"github.com/horizoncd/horizon/pkg/common"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	"github.com/horizoncd/horizon/pkg/oauth/models"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

type DAO interface {
	CreateApp(ctx context.Context, client models.OauthApp) error
	GetApp(ctx context.Context, clientID string) (*models.OauthApp, error)
	DeleteApp(ctx context.Context, clientID string) error
	ListApp(ctx context.Context, ownerType models.OwnerType, ownerID uint) ([]models.OauthApp, error)
	UpdateApp(ctx context.Context, clientID string, app models.OauthApp) (*models.OauthApp, error)
	CreateSecret(ctx context.Context, secret *models.OauthClientSecret) (*models.OauthClientSecret, error)
	DeleteSecret(ctx context.Context, clientID string, clientSecretID uint) error
	DeleteSecretByClientID(ctx context.Context, clientID string) error
	ListSecret(ctx context.Context, clientID string) ([]models.OauthClientSecret, error)
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

type dao struct {
	db *gorm.DB
}

func (d *dao) CreateApp(ctx context.Context, client models.OauthApp) error {
	result := d.db.WithContext(ctx).Save(&client)
	return result.Error
}
func (d *dao) GetApp(ctx context.Context, clientID string) (*models.OauthApp, error) {
	var client models.OauthApp
	result := d.db.WithContext(ctx).Raw(common.GetOauthAppByClientID, clientID).First(&client)
	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.OAuthInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.OAuthInDB, result.Error.Error())
	}
	return &client, nil
}

func (d *dao) UpdateApp(ctx context.Context,
	clientID string, app models.OauthApp) (*models.OauthApp, error) {
	var appInDb models.OauthApp
	var err error
	if err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Raw(common.GetOauthAppByClientID, clientID).Scan(&appInDb)
		if result.Error != nil {
			return herrors.NewErrGetFailed(herrors.OAuthInDB, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return herrors.NewErrNotFound(herrors.OAuthInDB, "row affected = 0s")
		}
		appInDb.Name = app.Name
		appInDb.HomeURL = app.HomeURL
		appInDb.RedirectURL = app.RedirectURL
		appInDb.Desc = app.Desc
		appInDb.UpdatedBy = app.UpdatedBy
		if err := tx.Save(&appInDb).Error; err != nil {
			return herrors.NewErrUpdateFailed(herrors.OAuthInDB, err.Error())
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &appInDb, err
}

func (d *dao) DeleteApp(ctx context.Context, clientID string) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteOauthAppByClientID, clientID)
	return result.Error
}

func (d *dao) ListApp(ctx context.Context, ownerType models.OwnerType,
	ownerID uint) ([]models.OauthApp, error) {
	var oauthApps []models.OauthApp
	if result := d.db.WithContext(ctx).Raw(common.SelectOauthAppByOwner, ownerType,
		ownerID).Scan(&oauthApps); result.Error != nil {
		return nil, result.Error
	}
	return oauthApps, nil
}

func (d *dao) CreateSecret(ctx context.Context,
	secret *models.OauthClientSecret) (*models.OauthClientSecret, error) {
	if result := d.db.WithContext(ctx).Save(secret); result.Error != nil {
		return nil, result.Error
	}
	return secret, nil
}
func (d *dao) DeleteSecretByClientID(ctx context.Context, clientID string) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteClientSecretByClientID, clientID)
	return result.Error
}
func (d *dao) DeleteSecret(ctx context.Context, clientID string, clientSecretID uint) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteClientSecret, clientID, clientSecretID)
	return result.Error
}
func (d *dao) ListSecret(ctx context.Context, clientID string) ([]models.OauthClientSecret, error) {
	var secrets []models.OauthClientSecret
	result := d.db.WithContext(ctx).Raw(common.ClientSecretSelectAll, clientID).Scan(&secrets)
	if result.Error != nil {
		return nil, result.Error
	}
	return secrets, nil
}
