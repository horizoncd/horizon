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

package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/models"
	"golang.org/x/oauth2"
)

type Claims struct {
	Sub   string `json:"sub"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func MakeOuath2Config(ctx context.Context, idp *models.IdentityProvider,
	extraScopes ...string) (*oauth2.Config, error) {
	if idp.ClientID != "" &&
		idp.ClientSecret != "" &&
		idp.AuthorizationEndpoint != "" &&
		idp.TokenEndpoint != "" {
		conf := &oauth2.Config{
			ClientID:     idp.ClientID,
			ClientSecret: idp.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  idp.AuthorizationEndpoint,
				TokenURL: idp.TokenEndpoint},
			Scopes: append(extraScopes, strings.Split(idp.Scopes, " ")...),
		}
		return conf, nil
	}

	return nil, perror.Wrapf(herrors.ErrParamInvalid,
		"lack infos when creating oauth2 config:\n"+
			"clientID = %s\n"+
			"clientSecret is empty: %v\n"+
			"authEndpoint = %s\n"+
			"tokenEndpoint = %s\n"+
			"scopes = %s",
		idp.ClientID,
		idp.ClientSecret == "",
		idp.AuthorizationEndpoint,
		idp.TokenEndpoint,
		idp.Scopes,
	)
}

func MakeOidcProvider(ctx context.Context, idp *models.IdentityProvider) (*oidc.Provider, error) {
	var (
		provider *oidc.Provider
		err      error
	)

	if idp.AuthorizationEndpoint != "" &&
		idp.TokenEndpoint != "" &&
		idp.UserinfoEndpoint != "" &&
		idp.Jwks != "" &&
		idp.SigningAlgs != "" {
		provider = (&oidc.ProviderConfig{
			IssuerURL:   idp.Issuer,
			AuthURL:     idp.AuthorizationEndpoint,
			TokenURL:    idp.TokenEndpoint,
			UserInfoURL: idp.UserinfoEndpoint,
			JWKSURL:     idp.Jwks,
			Algorithms:  strings.Split(idp.SigningAlgs, " "),
		}).NewProvider(ctx)
		return provider, nil
	}

	if idp.Issuer == "" {
		return nil, perror.Wrapf(herrors.ErrParamInvalid,
			"Issuer is empty and some other info is required "+
				"when creating oidc provider: \n"+
				"AuthEndpoint = %s\n"+
				"TokenEndpoint = %s\n"+
				"UserInfoEndpoint = %s\n"+
				"JwksEnpoint = %s\n"+
				"Algorithms = %s",
			idp.AuthorizationEndpoint,
			idp.TokenEndpoint,
			idp.UserinfoEndpoint,
			idp.Jwks,
			strings.Split(idp.SigningAlgs, " "),
		)
	}

	provider, err = oidc.NewProvider(ctx, idp.Issuer)
	if err != nil {
		return nil, perror.Wrap(
			herrors.NewErrGetFailed(herrors.ProviderFromDiscovery, err.Error()),
			fmt.Sprintf("failed to get provider from issuer: %s", idp.Issuer))
	}

	return provider, nil
}

// HandleOIDC gets user's email and name by code & token
func HandleOIDC(ctx context.Context, idp *models.IdentityProvider,
	code string, redirect ...string) (*Claims, error) {
	conf, err := MakeOuath2Config(ctx, idp)
	if err != nil {
		return nil, err
	}
	if len(redirect) > 0 {
		conf.RedirectURL = redirect[0]
	}

	provider, err := MakeOidcProvider(ctx, idp)
	if err != nil {
		return nil, err
	}

	// exchange token with code
	token, err := conf.Exchange(ctx, code, oauth2.AccessTypeOnline)
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.IdentityProviderInDB, err.Error()),
			"failed to get token:\n"+
				"code = %v\n err = %v", code, err)
	}

	// fetch userinfo by token
	userinfo, err := provider.UserInfo(ctx, conf.TokenSource(ctx, token))
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrParamInvalid,
			"get userinfo failed:\n"+
				"token = %v\n err = %v", token, err)
	}

	var claims Claims
	if err := userinfo.Claims(&claims); err != nil {
		return nil, perror.Wrapf(herrors.ErrParamInvalid,
			"failed to parse claims:\n"+
				"err = %v", err)
	}

	return &claims, nil
}
