// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oauth

import (
	"net/http"
	"time"

	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/oauth/manager"
	oauthmodel "github.com/horizoncd/horizon/pkg/oauth/models"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/token/generator"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"golang.org/x/net/context"
)

type AuthorizeReq struct {
	ClientID     string
	Scope        string
	RedirectURL  string
	State        string
	UserIdentity uint

	Request *http.Request
}

type AuthorizeCodeResponse struct {
	Code        string
	RedirectURL string
	State       string
}

type BaseTokenReq struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string

	Request *http.Request
}

type AccessTokenReq struct {
	BaseTokenReq
	Code string
}

type RefreshTokenReq struct {
	BaseTokenReq
	RefreshToken string
}

type AccessTokenResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    time.Duration `json:"expires_in"`
	Scope        string        `json:"scope"`
	TokenType    string        `json:"token_type"`
}

type Controller interface {
	// GenAuthorizeCode oauth  Authorization GenOauthTokensRequest ref:rfc6750
	GenAuthorizeCode(ctx context.Context, req *AuthorizeReq) (*AuthorizeCodeResponse, error)
	// GenAccessToken Access Token GenOauthTokensRequest,ref:rfc6750
	GenAccessToken(ctx context.Context, req *AccessTokenReq) (*AccessTokenResponse, error)
	RefreshToken(ctx context.Context, req *RefreshTokenReq) (*AccessTokenResponse, error)
}

func NewController(param *param.Param) Controller {
	return &controller{oauthManager: param.OauthManager}
}

var _ Controller = &controller{}

type controller struct {
	oauthManager manager.Manager
}

func (c *controller) GenAuthorizeCode(ctx context.Context, req *AuthorizeReq) (*AuthorizeCodeResponse, error) {
	const op = "oauth controller: GenAuthorizeCode"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. TODO: check if the scope is ok (now horizon app do not need provide scope)
	// 2. gen authorization Code
	authToken, err := c.oauthManager.GenAuthorizeCode(ctx, &manager.AuthorizeGenerateRequest{
		ClientID:     req.ClientID,
		RedirectURL:  req.RedirectURL,
		State:        req.State,
		Scope:        req.Scope,
		UserIdentify: req.UserIdentity,
		Request:      req.Request,
	})
	if err != nil {
		return nil, err
	}
	resp := &AuthorizeCodeResponse{
		Code:        authToken.Code,
		RedirectURL: authToken.RedirectURI,
		State:       req.State,
	}
	return resp, nil
}

func (c *controller) getAccessTokenGenerator(ctx context.Context,
	clientID string) (generator.AccessTokenCodeGenerator, error) {
	app, err := c.oauthManager.GetOAuthApp(ctx, clientID)
	if err != nil {
		return nil, err
	}
	var gen generator.AccessTokenCodeGenerator
	switch app.AppType {
	case oauthmodel.HorizonOAuthAPP:
		gen = generator.NewHorizonAppUserToServerAccessGenerator()
	case oauthmodel.DirectOAuthAPP:
		gen = generator.NewOauthAccessGenerator()
	default:
		return nil, perror.Wrapf(herrors.ErrOAuthInternal,
			"appType Not Supported, appType = %d", app.AppType)
	}
	return gen, nil
}

func (c *controller) GenAccessToken(ctx context.Context, req *AccessTokenReq) (*AccessTokenResponse, error) {
	accessTokenGenerator, err := c.getAccessTokenGenerator(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}
	refreshTokenGenerator := generator.NewBasicRefreshTokenGenerator()

	tokens, err := c.oauthManager.GenOauthTokens(ctx, &manager.OauthTokensRequest{
		ClientID:              req.ClientID,
		ClientSecret:          req.ClientSecret,
		Code:                  req.Code,
		RedirectURL:           req.RedirectURL,
		Request:               req.Request,
		AccessTokenGenerator:  accessTokenGenerator,
		RefreshTokenGenerator: refreshTokenGenerator,
	})
	if err != nil {
		return nil, err
	}
	return &AccessTokenResponse{
		AccessToken:  tokens.AccessToken.Code,
		RefreshToken: tokens.RefreshToken.Code,
		ExpiresIn:    tokens.AccessToken.ExpiresIn,
		Scope:        tokens.AccessToken.Scope,
		TokenType:    "bearer",
	}, nil
}

func (c *controller) RefreshToken(ctx context.Context,
	req *RefreshTokenReq) (*AccessTokenResponse, error) {
	accessTokenGenerator, err := c.getAccessTokenGenerator(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}
	refreshTokenGenerator := generator.NewBasicRefreshTokenGenerator()

	tokens, err := c.oauthManager.RefreshOauthTokens(ctx, &manager.OauthTokensRequest{
		ClientID:              req.ClientID,
		ClientSecret:          req.ClientSecret,
		RefreshToken:          req.RefreshToken,
		RedirectURL:           req.RedirectURL,
		Request:               req.Request,
		AccessTokenGenerator:  accessTokenGenerator,
		RefreshTokenGenerator: refreshTokenGenerator,
	})
	if err != nil {
		return nil, err
	}
	return &AccessTokenResponse{
		AccessToken:  tokens.AccessToken.Code,
		RefreshToken: tokens.RefreshToken.Code,
		ExpiresIn:    tokens.AccessToken.ExpiresIn,
		Scope:        tokens.AccessToken.Scope,
		TokenType:    "bearer",
	}, nil
}
