package idp

import (
	"time"

	"g.hz.netease.com/horizon/pkg/idp/models"
)

type AuthInfo struct {
	AuthURL     string `json:"authURL"`
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
