package idp

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/core/common"
	idpconst "github.com/horizoncd/horizon/pkg/core/common/idp"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/idp/manager"
	"github.com/horizoncd/horizon/pkg/idp/utils"
	"github.com/horizoncd/horizon/pkg/param"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	usermodel "github.com/horizoncd/horizon/pkg/user/models"
	linkmanager "github.com/horizoncd/horizon/pkg/userlink/manager"
	"golang.org/x/oauth2"
)

var (
	providerKey = "provider"
	linkKey     = "link"
)

type Controller interface {
	ListAuthEndpoints(ctx context.Context, redirectURL string) ([]*AuthInfo, error)
	List(ctx context.Context) ([]*IdentityProvider, error)
	GetByID(ctx context.Context, id uint) (*IdentityProvider, error)
	LoginOrLink(ctx context.Context, code string, state string, redirectURL string) (*usermodel.User, error)
	Create(c context.Context, createParam *CreateIDPRequest) (*IdentityProvider, error)
	Delete(c context.Context, idpID uint) error
	Update(c context.Context, id uint, updateParam *UpdateIDPRequest) (*IdentityProvider, error)
	GetDiscovery(ctx context.Context, s Discovery) (*DiscoveryConfig, error)
}

type controller struct {
	idpManager  manager.Manager
	userManager usermanager.Manager
	linkManager linkmanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		idpManager:  param.IdpManager,
		userManager: param.UserManager,
		linkManager: param.UserLinksManager,
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
		info := &AuthInfo{ID: idp.ID, Name: idp.Name, DisplayName: idp.DisplayName}
		conf, err = utils.MakeOuath2Config(ctx, idp, oidc.ScopeOpenID)
		if err != nil {
			return nil, err
		}

		state := url.Values{providerKey: []string{idp.Name}}
		conf.RedirectURL = redirectURL
		info.AuthURL = conf.AuthCodeURL(
			base64.StdEncoding.EncodeToString([]byte(state.Encode())),
			oauth2.AccessTypeOnline)

		res = append(res, info)
	}
	return res, nil
}

func (c *controller) List(ctx context.Context) ([]*IdentityProvider, error) {
	idps, err := c.idpManager.List(ctx)
	if err != nil {
		return nil, err
	}

	return ofIDPModels(idps), nil
}

func (c *controller) LoginOrLink(ctx context.Context,
	code string, state string, redirectURL string) (*usermodel.User, error) {
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

	var claims *utils.Claims
	claims, err = utils.HandleOIDC(ctx, idp, code, redirectURL)
	if err != nil {
		return nil, err
	}

	currentUser, _ := common.UserFromContext(ctx)
	var user *usermodel.User
	if v, ok := stateMap[linkKey]; ok && len(v) == 1 && v[0] == "true" && currentUser != nil {
		// for linking
		user, err = c.userManager.GetUserByID(ctx, currentUser.GetID())
		if err != nil {
			return nil, err
		}
		_, err = c.linkManager.CreateLink(ctx, user.ID, idp.ID, claims, true)
		if err != nil {
			return nil, err
		}
	} else {
		if link, err := c.linkManager.GetByIDPAndSub(ctx, idp.ID, claims.Sub); err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
				return nil, err
			}
			// for register
			name := strings.SplitN(claims.Email, "@", 2)[0]
			if claims.Name == "" {
				claims.Name = name
			}
			user, err = c.userManager.Create(ctx, &usermodel.User{
				Name:     name,
				FullName: claims.Name,
				Email:    claims.Email,
			})
			if err != nil {
				return nil, err
			}
			_, err = c.linkManager.CreateLink(ctx, user.ID, idp.ID, claims, false)
			if err != nil {
				return nil, err
			}
		} else {
			// for signing in
			user, _ = c.userManager.GetUserByID(ctx, link.UserID)
			if user != nil {
				if user.Banned {
					return nil, perror.Wrapf(herrors.ErrForbidden,
						"user is banned")
				}
			}
		}
	}
	return user, nil
}

func (c *controller) GetByID(ctx context.Context, id uint) (*IdentityProvider, error) {
	idp, err := c.idpManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ofIDPModel(idp), nil
}

func (c *controller) Create(ctx context.Context,
	createParam *CreateIDPRequest) (*IdentityProvider, error) {
	idp := createParam.toModel()

	_, err := c.idpManager.GetByCondition(ctx,
		q.Query{Keywords: map[string]interface{}{idpconst.QueryName: idp.Name}})
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return nil, err
		}
	} else {
		return nil, perror.Wrapf(herrors.ErrNameConflict, "name = %v", idp.Name)
	}
	idp, err = c.idpManager.Create(ctx, idp)
	if err != nil {
		return nil, err
	}
	return ofIDPModel(idp), err
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
	return ofIDPModel(idp), nil
}

func (c *controller) GetDiscovery(ctx context.Context, s Discovery) (*DiscoveryConfig, error) {
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
