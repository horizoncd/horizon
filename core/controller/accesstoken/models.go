package accesstoken

import (
	"time"

	usermodels "g.hz.netease.com/horizon/pkg/user/models"
)

const (
	RobotEmailSuffix = "@noreply.com"
	NeverExpire      = "never"
	ExpiresAtFormat  = "2006-01-02"
)

type CreatePersonalAccessTokenRequest struct {
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	ExpiresAt string   `json:"expiresAt"`
}

type CreateResourceAccessTokenRequest struct {
	CreatePersonalAccessTokenRequest
	Role string `json:"role"`
}

type PersonalAccessToken struct {
	CreatePersonalAccessTokenRequest
	ID        uint                  `json:"id"`
	CreatedAt time.Time             `json:"createdAt"`
	CreatedBy *usermodels.UserBasic `json:"createdBy"`
}

type ResourceAccessToken struct {
	CreateResourceAccessTokenRequest
	ID        uint                  `json:"id"`
	CreatedAt time.Time             `json:"createdAt"`
	CreatedBy *usermodels.UserBasic `json:"createdBy"`
}

type CreatePersonalAccessTokenResponse struct {
	PersonalAccessToken
	Token string `json:"token"`
}

type CreateResourceAccessTokenResponse struct {
	ResourceAccessToken
	Token string `json:"token"`
}
