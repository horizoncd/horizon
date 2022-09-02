package idp

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/idp/manager"
	"g.hz.netease.com/horizon/pkg/idp/utils"
	"g.hz.netease.com/horizon/pkg/param"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	usermodel "g.hz.netease.com/horizon/pkg/user/models"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	providerKey = "provider"

	redirectKey = "redirect"
)

type Controller interface {
	ListAuthEndpoints(ctx context.Context, redirectURL string) ([]*AuthInfo, error)
	ListIDPs(ctx context.Context) ([]*IdentityProvider, error)
	Login(ctx context.Context, code string, state string) (*usermodel.User, error)
}

type controller struct {
	idpManager  manager.Manager
	userManager usermanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		idpManager:  param.IdpManager,
		userManager: param.UserManager,
	}
}

func (c *controller) ListAuthEndpoints(ctx context.Context, redirectURL string) ([]*AuthInfo, error) {
	idps, err := c.idpManager.ListIDP(ctx)
	if err != nil {
		return nil, err
	}

	var (
		conf *oauth2.Config
		res  = make([]*AuthInfo, 0)
	)
	for _, idp := range idps {
		info := &AuthInfo{DisplayName: idp.DisplayName}
		conf, err = utils.MakeOuath2Config(ctx, idp, oidc.ScopeOpenID)
		if err != nil {
			return nil, err
		}

		state := url.Values{providerKey: []string{idp.Name}}
		conf.RedirectURL = redirectURL
		info.AuthURL = conf.AuthCodeURL(base64.StdEncoding.EncodeToString([]byte(state.Encode())))

		res = append(res, info)
	}
	return res, nil
}

func (c *controller) ListIDPs(ctx context.Context) ([]*IdentityProvider, error) {
	idps, err := c.idpManager.ListIDP(ctx)
	if err != nil {
		return nil, err
	}

	return ConvertIDPs(idps), nil
}

func (c *controller) Login(ctx context.Context, code string, state string) (*usermodel.User, error) {
	bts, err := base64.StdEncoding.DecodeString(state)
	if err != nil {
		return nil, perror.Wrapf(
			herrors.ErrParamInvalid,
			"state is invalid:\n"+
				"state = %s\n err = %v", state, err)
	}

	state = string(bts)
	stateMap, err := url.ParseQuery(state)
	if err != nil {
		return nil, perror.Wrapf(
			herrors.ErrParamInvalid,
			"state is invalid:\n"+
				"state = %s\n err = %v", state, err)
	}

	providerName, ok := stateMap[providerKey]
	if !ok || len(providerName) < 1 {
		return nil, perror.Wrapf(
			herrors.ErrParamInvalid,
			"no identity name in state:\n"+
				"state = %v", stateMap)
	}

	idp, err := c.idpManager.GetProviderByName(ctx, providerName[0])
	if err != nil {
		return nil, err
	}

	redirect := stateMap[redirectKey]

	var claims *utils.Claims
	claims, err = utils.HandleOIDC(ctx, idp, code, redirect...)
	if err != nil {
		return nil, err
	}

	user, _ := c.userManager.GetUserByEmail(ctx, claims.Email)
	if user == nil {
		user, err = c.userManager.Create(ctx, &usermodel.User{
			Name:     strings.SplitN(claims.Email, "@", 2)[0],
			FullName: claims.Name,
			Email:    claims.Email,
			OIDCType: idp.Name,
		})
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}
