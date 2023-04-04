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

package idp

import (
	"time"

	"github.com/horizoncd/horizon/pkg/models"
)

type AuthInfo struct {
	ID          uint   `json:"id"`
	AuthURL     string `json:"authURL"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
}

type IdentityProvider struct {
	ID                      uint                           `json:"id"`
	DisplayName             string                         `json:"displayName,omitempty"`
	Name                    string                         `json:"name,omitempty"`
	Avatar                  string                         `json:"avatar,omitempty"`
	AuthorizationEndpoint   string                         `json:"authorizationEndpoint,omitempty"`
	TokenEndpoint           string                         `json:"tokenEndpoint,omitempty"`
	UserinfoEndpoint        string                         `json:"userinfoEndpoint,omitempty"`
	RevocationEndpoint      string                         `json:"revocationEndpoint,omitempty"`
	Issuer                  string                         `json:"issuer,omitempty"`
	Scopes                  string                         `json:"scopes,omitempty"`
	TokenEndpointAuthMethod models.TokenEndpointAuthMethod `json:"tokenEndpointAuthMethod,omitempty"`
	Jwks                    string                         `json:"jwks,omitempty"`
	ClientID                string                         `json:"clientID,omitempty"`
	ClientSecret            string                         `json:"clientSecret,omitempty"`
	CreatedAt               time.Time                      `json:"createdAt"`
	UpdatedAt               time.Time                      `json:"updatedAt"`
}

func ofIDPModel(idp *models.IdentityProvider) *IdentityProvider {
	var method = models.TokenEndpointAuthMethod(models.ClientSecretSentAsPost)
	if idp.TokenEndpointAuthMethod != nil {
		method = *idp.TokenEndpointAuthMethod
	}
	return &IdentityProvider{
		ID:                      idp.ID,
		DisplayName:             idp.DisplayName,
		Name:                    idp.Name,
		Avatar:                  idp.Avatar,
		AuthorizationEndpoint:   idp.AuthorizationEndpoint,
		TokenEndpoint:           idp.TokenEndpoint,
		UserinfoEndpoint:        idp.UserinfoEndpoint,
		RevocationEndpoint:      idp.RevocationEndpoint,
		Issuer:                  idp.Issuer,
		Scopes:                  idp.Scopes,
		TokenEndpointAuthMethod: method,
		Jwks:                    idp.Jwks,
		ClientID:                idp.ClientID,
		ClientSecret:            idp.ClientSecret,
		CreatedAt:               idp.CreatedAt,
		UpdatedAt:               idp.UpdatedAt,
	}
}

func ofIDPModels(idps []*models.IdentityProvider) []*IdentityProvider {
	res := make([]*IdentityProvider, 0, len(idps))
	for _, idp := range idps {
		res = append(res, ofIDPModel(idp))
	}
	return res
}

type CreateIDPRequest struct {
	UpdateIDPRequest
}

func (r *CreateIDPRequest) toModel() *models.IdentityProvider {
	idp := &models.IdentityProvider{
		DisplayName:             r.DisplayName,
		Name:                    r.Name,
		Avatar:                  r.Avatar,
		AuthorizationEndpoint:   r.AuthorizationEndpoint,
		TokenEndpoint:           r.TokenEndpoint,
		UserinfoEndpoint:        r.UserinfoEndpoint,
		RevocationEndpoint:      r.RevocationEndpoint,
		Issuer:                  r.Issuer,
		Scopes:                  r.Scopes,
		SigningAlgs:             r.SigningAlgs,
		TokenEndpointAuthMethod: &r.TokenEndpointAuthMethod,
		Jwks:                    r.Jwks,
		ClientID:                r.ClientID,
		ClientSecret:            r.ClientSecret,
	}
	return idp
}

type UpdateIDPRequest struct {
	DisplayName             string                         `json:"displayName"`
	Name                    string                         `json:"name"`
	Avatar                  string                         `json:"avatar,omitempty"`
	AuthorizationEndpoint   string                         `json:"authorizationEndpoint"`
	TokenEndpoint           string                         `json:"tokenEndpoint"`
	UserinfoEndpoint        string                         `json:"userinfoEndpoint,omitempty"`
	RevocationEndpoint      string                         `json:"revocationEndpoint,omitempty"`
	Issuer                  string                         `json:"issuer"`
	Scopes                  string                         `json:"scopes"`
	SigningAlgs             string                         `json:"signingAlgs,omitempty"`
	TokenEndpointAuthMethod models.TokenEndpointAuthMethod `json:"tokenEndpointAuthMethod,omitempty"`
	Jwks                    string                         `json:"jwks,omitempty"`
	ClientID                string                         `json:"clientID"`
	ClientSecret            string                         `json:"clientSecret"`
}

func (r *UpdateIDPRequest) toModel() *models.IdentityProvider {
	idp := &models.IdentityProvider{
		DisplayName:             r.DisplayName,
		Name:                    r.Name,
		Avatar:                  r.Avatar,
		AuthorizationEndpoint:   r.AuthorizationEndpoint,
		TokenEndpoint:           r.TokenEndpoint,
		UserinfoEndpoint:        r.UserinfoEndpoint,
		RevocationEndpoint:      r.RevocationEndpoint,
		Issuer:                  r.Issuer,
		Scopes:                  r.Scopes,
		SigningAlgs:             r.SigningAlgs,
		TokenEndpointAuthMethod: &r.TokenEndpointAuthMethod,
		Jwks:                    r.Jwks,
		ClientID:                r.ClientID,
		ClientSecret:            r.ClientSecret,
	}
	return idp
}

type Discovery struct {
	FromURL string `json:"fromUrl"`
}

type DiscoveryConfig struct {
	AuthorizationEndpoint string `json:"authorizationEndpoint"`
	TokenEndpoint         string `json:"tokenEndpoint"`
	Issuer                string `json:"issuer"`
}
