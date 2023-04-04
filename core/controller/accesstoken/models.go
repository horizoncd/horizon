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

package accesstoken

import (
	"time"

	usermodels "github.com/horizoncd/horizon/pkg/models"
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
