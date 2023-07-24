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
	"time"

	"github.com/horizoncd/horizon/pkg/rbac/types"
)

type Scopes struct {
	DefaultScopes []string     `yaml:"defaultScope"`
	Roles         []types.Role `yaml:"roles"`
}
type Server struct {
	OauthHTMLLocation     string        `yaml:"oauthHTMLLocation"`
	AuthorizeCodeExpireIn time.Duration `yaml:"authorizeCodeExpireIn"`
	AccessTokenExpireIn   time.Duration `yaml:"accessTokenExpireIn"`
	RefreshTokenExpireIn  time.Duration `yaml:"refreshTokenExpireIn"`
}
