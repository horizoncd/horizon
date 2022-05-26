package oauthapp

import (
	"g.hz.netease.com/horizon/pkg/oauth/manager"
	"g.hz.netease.com/horizon/pkg/oauth/models"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"golang.org/x/net/context"
)

type CreateOauthAPPRequest struct {
	Name        string
	Desc        string
	HomeURL     string
	RedirectURL string
}

type APPBasicInfo struct {
	AppID       uint
	AppName     string
	Decs        string
	HomeURL     string
	ClientID    string
	RedirectURL string
}

type Controller interface {
	Create(ctx context.Context, groupID uint, request CreateOauthAPPRequest) (*APPBasicInfo, error)
	Get(ctx context.Context, clientID string) (*APPBasicInfo, error)
	List(ctx context.Context, groupID uint) ([]APPBasicInfo, error)
	Update(ctx context.Context, info APPBasicInfo) (*APPBasicInfo, error)
	Delete(ctx context.Context, clientID string) error
}

func NewController(authManager manager.Manager) Controller {
	return &controller{oauthManager: authManager}
}

type controller struct {
	oauthManager manager.Manager
}

func (c controller) Create(ctx context.Context, groupID uint, request CreateOauthAPPRequest) (*APPBasicInfo, error) {
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

func (c controller) Get(ctx context.Context, clientID string) (*APPBasicInfo, error) {
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
		HomeURL:     oauthApp.Desc,
		ClientID:    oauthApp.ClientID,
		RedirectURL: oauthApp.RedirectURL,
	}
	return resp, err
}

func (c controller) List(ctx context.Context, groupID uint) ([]APPBasicInfo, error) {
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

func (c controller) Update(ctx context.Context, info APPBasicInfo) (*APPBasicInfo, error) {
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

func (c controller) Delete(ctx context.Context, clientID string) error {
	return c.oauthManager.DeleteOAuthApp(ctx, clientID)
}

var _ Controller = &controller{}
