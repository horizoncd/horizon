package accesstoken

import (
	"time"
)

const (
	RobotEmailSuffix = "@noreply.com"
	NeverExpire      = "never"
	ExpiresAtFormat  = "2006-01-02"
)

type CreateAccessTokenRequest struct {
	Name      string    `json:"name"`
	Role      string    `json:"role,omitempty"`
	Scopes    []string  `json:"scopes,omitempty"`
	ExpiresAt string    `json:"expiresAt"`
	Resource  *Resource `json:"-"`
}

type AccessToken struct {
	CreateAccessTokenRequest
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	CreatedBy *User     `json:"createdBy"`
}

type CreateAccessTokenResponse struct {
	AccessToken
	Token string `json:"token"`
}

type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Resource struct {
	ResourceType string
	ResourceID   uint
}
