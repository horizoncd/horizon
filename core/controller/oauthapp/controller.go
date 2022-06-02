package oauthapp

import (
	"time"

	"g.hz.netease.com/horizon/pkg/oauth/manager"
	"g.hz.netease.com/horizon/pkg/oauth/models"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"golang.org/x/net/context"
)

type CreateOauthAPPRequest struct {
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	HomeURL     string `json:"homeURL"`
	RedirectURL string `json:"redirectURL"`
}

type APPBasicInfo struct {
	AppID       uint   `json:"appID"`
	AppName     string `json:"appName"`
	Decs        string `json:"decs"`
	HomeURL     string `json:"homeURL"`
	ClientID    string `json:"clientID"`
	RedirectURL string `json:"redirectURL"`
}

type Controller interface {
	Create(ctx context.Context, groupID uint, request CreateOauthAPPRequest) (*APPBasicInfo, error)
	Get(ctx context.Context, clientID string) (*APPBasicInfo, error)
	List(ctx context.Context, groupID uint) ([]APPBasicInfo, error)
	Update(ctx context.Context, info APPBasicInfo) (*APPBasicInfo, error)
	Delete(ctx context.Context, clientID string) error

	CreateSecret(ctx context.Context, clientID string) (*SecretBasic, error)
	DeleteSecret(ctx context.Context, ClientID string, clientSecretID uint) error
	ListSecret(ctx context.Context, ClientID string) ([]SecretBasic, error)
}

var _ Controller = &controller{}

func NewController(authManager manager.Manager,
	userManager usermanager.Manager) Controller {
	return &controller{
		oauthManager: authManager,
		userManager:  userManager,
	}
}

type controller struct {
	oauthManager manager.Manager
	userManager  usermanager.Manager
}

type SecretBasic struct {
	ID           uint      `json:"id"`
	ClientID     string    `json:"clientID"`
	ClientSecret string    `json:"clientSecret"`
	CreatedAt    time.Time `json:"createdAt"`
	CreatedBy    string    `json:"createdBy"`
}

func (c *controller) ofClientSecret(ctx context.Context, secret *models.OauthClientSecret) (*SecretBasic, error) {
	user, err := c.userManager.GetUserByID(ctx, secret.CreatedBy)
	if err != nil {
		return nil, err
	}
	return &SecretBasic{
		ID:           secret.ID,
		ClientID:     secret.ClientID,
		ClientSecret: secret.ClientSecret,
		CreatedAt:    secret.CreatedAt,
		CreatedBy:    user.Name,
	}, nil
}
func (c *controller) CreateSecret(ctx context.Context, clientID string) (*SecretBasic, error) {
	const op = "oauth app controller  CreateSecret"
	defer wlog.Start(ctx, op).StopPrint()
	secret, err := c.oauthManager.CreateSecret(ctx, clientID)
	if err != nil {
		return nil, err
	}
	secretBasic, err := c.ofClientSecret(ctx, secret)
	if err != nil {
		return nil, err
	}
	return secretBasic, nil
}

func (c *controller) DeleteSecret(ctx context.Context, ClientID string, clientSecretID uint) error {
	const op = "oauth app controller  DeleteSecret"
	defer wlog.Start(ctx, op).StopPrint()
	return c.oauthManager.DeleteSecret(ctx, ClientID, clientSecretID)
}

func (c *controller) ListSecret(ctx context.Context, ClientID string) ([]SecretBasic, error) {
	const op = "oauth app controller  ListSecret"
	defer wlog.Start(ctx, op).StopPrint()
	secrets, err := c.oauthManager.ListSecret(ctx, ClientID)
	if err != nil {
		return nil, err
	}
	translate := func() ([]SecretBasic, error) {
		var secretBasics = make([]SecretBasic, 0)
		for _, secret := range secrets {
			basic, err := c.ofClientSecret(ctx, &secret)
			if err != nil {
				return nil, err
			}
			secretBasics = append(secretBasics, *basic)
		}
		return secretBasics, nil
	}
	return translate()
}

func (c *controller) Create(ctx context.Context, groupID uint, request CreateOauthAPPRequest) (*APPBasicInfo, error) {
	const op = "oauth app controller  Create"
	defer wlog.Start(ctx, op).StopPrint()

	// TODO: check if have the permission to create
	oauthApp, err := c.oauthManager.CreateOauthApp(ctx, &manager.CreateOAuthAppReq{
		Name:        request.Name,
		RedirectURI: request.RedirectURL,
		HomeURL:     request.HomeURL,
		Desc:        request.Desc,
		OwnerType:   models.GroupOwnerType,
		OwnerID:     groupID,
		APPType:     models.DirectOAuthAPP,
	})
	if err != nil {
		return nil, err
	}
	resp := &APPBasicInfo{
		AppID:       oauthApp.ID,
		AppName:     oauthApp.Name,
		Decs:        oauthApp.Desc,
		HomeURL:     oauthApp.Desc,
		ClientID:    oauthApp.ClientID,
		RedirectURL: oauthApp.RedirectURL,
	}
	return resp, err
}

func (c *controller) Get(ctx context.Context, clientID string) (*APPBasicInfo, error) {
	const op = "oauth app controller  Get"
	defer wlog.Start(ctx, op).StopPrint()

	oauthApp, err := c.oauthManager.GetOAuthApp(ctx, clientID)
	if err != nil {
		return nil, err
	}
	resp := &APPBasicInfo{
		AppID:       oauthApp.ID,
		AppName:     oauthApp.Name,
		Decs:        oauthApp.Desc,
		HomeURL:     oauthApp.HomeURL,
		ClientID:    oauthApp.ClientID,
		RedirectURL: oauthApp.RedirectURL,
	}
	return resp, err
}

func (c *controller) List(ctx context.Context, groupID uint) ([]APPBasicInfo, error) {
	const op = "oauth  app controller  List"
	defer wlog.Start(ctx, op).StopPrint()

	apps, err := c.oauthManager.ListOauthApp(ctx, models.GroupOwnerType, groupID)
	if err != nil {
		return nil, err
	}
	var appInfos = make([]APPBasicInfo, 0)
	for _, app := range apps {
		appInfos = append(appInfos, APPBasicInfo{
			AppID:       app.ID,
			AppName:     app.Name,
			Decs:        app.Desc,
			HomeURL:     app.HomeURL,
			ClientID:    app.ClientID,
			RedirectURL: app.RedirectURL,
		})
	}
	return appInfos, nil
}

func (c *controller) Update(ctx context.Context, info APPBasicInfo) (*APPBasicInfo, error) {
	const op = "oauth  app controller  Update"
	defer wlog.Start(ctx, op).StopPrint()

	app, err := c.oauthManager.UpdateOauthApp(ctx, info.ClientID, manager.UpdateOauthAppReq{
		Name:        info.AppName,
		HomeURL:     info.HomeURL,
		RedirectURI: info.RedirectURL,
		Desc:        info.Decs,
	})
	if err != nil {
		return nil, err
	}
	return &APPBasicInfo{
		AppID:       app.ID,
		AppName:     app.Name,
		Decs:        app.Desc,
		HomeURL:     app.HomeURL,
		ClientID:    app.ClientID,
		RedirectURL: app.RedirectURL,
	}, nil
}

func (c *controller) Delete(ctx context.Context, clientID string) error {
	return c.oauthManager.DeleteOAuthApp(ctx, clientID)
}
