package oauth

import (
	"time"

	"g.hz.netease.com/horizon/pkg/rbac/types"
)

type Scopes struct {
	DefaultScopes []string     `yaml:"DefaultScope"`
	Roles         []types.Role `yaml:"Roles"`
}
type Server struct {
	OauthHTMLLocation     string        `yaml:"oauthHTMLLocation"`
	AuthorizeCodeExpireIn time.Duration `yaml:"authorizeCodeExpireIn"`
	AccessTokenExpireIn   time.Duration `yaml:"accessTokenExpireIn"`
}
