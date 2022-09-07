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
	List(ctx context.Context) ([]*IdentityProvider, error)
	GetByID(ctx context.Context, id uint) (*IdentityProvider, error)
	Login(ctx context.Context, code string, state string) (*usermodel.User, error)
	Create(c context.Context, createParam *CreateIDPRequest) (*IdentityProvider, error)
	Delete(c context.Context, idpID uint) error
	Update(c context.Context, u uint, updateParam *UpdateIDPRequest) (*IdentityProvider, error)
	GetDiscovery(ctx context.Context, s *Discovery) (*DiscoveryConfig, error)
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
	idps, err := c.idpManager.List(ctx)
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

func (c *controller) List(ctx context.Context) ([]*IdentityProvider, error) {
	idps, err := c.idpManager.List(ctx)
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

func (c *controller) GetByID(ctx context.Context, id uint) (*IdentityProvider, error) {
	idp, err := c.idpManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ConvertIDP(idp), nil
}

func (c *controller) Create(ctx context.Context,
	createParam *CreateIDPRequest) (*IdentityProvider, error) {
	idp := createParam.toModel()
	idp, err := c.idpManager.Create(ctx, idp)
	if err != nil {
		return nil, err
	}
	return ConvertIDP(idp), err
}

func (c *controller) Delete(ctx context.Context, idpID uint) error {
	if _, err := c.idpManager.GetByID(ctx, idpID); err != nil {
		return err
	}
	return c.idpManager.Delete(ctx, idpID)
}

func (c *controller) Update(ctx context.Context,
	id uint, updateParam *UpdateIDPRequest) (*IdentityProvider, error) {
	updateIDP := updateParam.toModel()
	idp, err := c.idpManager.Update(ctx, id, updateIDP)
	if err != nil {
		return nil, err
	}
	return ConvertIDP(idp), nil
}

func (c *controller) GetDiscovery(ctx context.Context, s *Discovery) (*DiscoveryConfig, error) {
	issuer := strings.TrimSuffix(
		strings.TrimSuffix(s.FromURL, "/"),
		"/.well-known/openid-configuration",
	)
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, perror.Wrapf(
			herrors.ErrHTTPRequestFailed,
			"failed to get discovery:\n"+
				"url = %s\nerr = %v", s.FromURL, err)
	}

	endpoint := provider.Endpoint()
	return &DiscoveryConfig{
		AuthorizationEndpoint: endpoint.AuthURL,
		TokenEndpoint:         endpoint.TokenURL,
		Issuer:                issuer,
	}, nil
}
