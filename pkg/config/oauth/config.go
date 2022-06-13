package oauth

import (
	"time"

	"g.hz.netease.com/horizon/pkg/rbac/types"
)

type Scopes struct {
	DefaultScopes []string     `yaml:"defaultScope"`
	Roles         []types.Role `yaml:"roles"`
}
type Server struct {
	OauthHTMLLocation     string        `yaml:"oauthHTMLLocation"`
	AuthorizeCodeExpireIn time.Duration `yaml:"authorizeCodeExpireIn"`
	AccessTokenExpireIn   time.Duration `yaml:"accessTokenExpireIn"`
}
