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
	AuthURL                 string                         `json:"authURL,omitempty"`
	TokenEndpoint           string                         `json:"tokenEndpoint,omitempty"`
	UserinfoEndpoint        string                         `json:"userinfoEndpoint,omitempty"`
	RevocationEndpoint      string                         `json:"revocationEndpoint,omitempty"`
	Issuer                  string                         `json:"issuer,omitempty"`
	Scopes                  string                         `json:"scopes,omitempty"`
	TokenEndpointAuthMethod models.TokenEndpointAuthMethod `json:"tokenEndpointAuthMethod,omitempty"`
	Jwks                    string                         `json:"jwks,omitempty"`
	ClientID                string                         `json:"clientID,omitempty"`
	ClientSecret            string                         `json:"-"`
	CreatedAt               time.Time                      `json:"createdAt"`
	UpdatedAt               time.Time                      `json:"updatedAt"`
}

func ConvertIDP(idp *models.IdentityProvider) *IdentityProvider {
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
		TokenEndpointAuthMethod: idp.TokenEndpointAuthMethod,
		Jwks:                    idp.Jwks,
		ClientID:                idp.ClientID,
		ClientSecret:            idp.ClientSecret,
		CreatedAt:               idp.CreatedAt,
		UpdatedAt:               idp.UpdatedAt,
	}
}

func ConvertIDPs(idps []*models.IdentityProvider) []*IdentityProvider {
	res := make([]*IdentityProvider, 0, len(idps))
	for _, idp := range idps {
		res = append(res, ConvertIDP(idp))
	}
	return res
}
