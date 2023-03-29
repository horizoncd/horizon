package horizonapp

import (
	"github.com/horizoncd/horizon/pkg/oauth/models"
	"golang.org/x/net/context"
)

type PostRequest struct {
	AppName                  string
	Desc                     string
	HomeURL                  string
	AuthorizationCallbackURL string
	WebHook                  string
	Permissions              []Permission
}

type BasicInfo struct {
	AppID   uint // TODO: if use uuid here
	AppName string
	Desc    string
	HomeURL string

	ClientID                 string
	AuthorizationCallbackURL string
	WebHook                  string
}

type Permission struct {
	Resource string
	Scope    []string
}

type Controller interface {
	Create(ctx context.Context, groupID uint, body *PostRequest) (*BasicInfo, error)
	Get(ctx context.Context, appID uint) (*BasicInfo, error)
	List(ctx context.Context, groupID uint) ([]BasicInfo, error)
	Delete(ctx context.Context, appID uint) (*BasicInfo, error)
	Update(ctx context.Context, info *BasicInfo) (*BasicInfo, error)

	CreateOrUpdatePermission(ctx context.Context, appID uint, permissions []Permission) ([]Permission, error)
	GetPermission(ctx context.Context, appID uint) ([]Permission, error)

	CreateClientSecret(ctx context.Context, appID uint) (secret *models.OauthClientSecret, err error)
	ListClientSecret(ctx context.Context, appID uint) ([]models.OauthClientSecret, error)
	DeleteClientSecret(ctx context.Context, appID uint, secretID uint) (*models.OauthClientSecret, error)
}
