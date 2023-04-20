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

package horizonapp

import (
	"github.com/horizoncd/horizon/pkg/oauth/models"
	"golang.org/x/net/context"
)

type PostRequest struct {
	AppName                  string
	Desc                     string
	HomeURL                  string
	AuthorizationCallbackURL string
	WebHook                  string
	Permissions              []Permission
}

type BasicInfo struct {
	AppID   uint // TODO: if use uuid here
	AppName string
	Desc    string
	HomeURL string

	ClientID                 string
	AuthorizationCallbackURL string
	WebHook                  string
}

type Permission struct {
	Resource string
	Scope    []string
}

type Controller interface {
	Create(ctx context.Context, groupID uint, body *PostRequest) (*BasicInfo, error)
	Get(ctx context.Context, appID uint) (*BasicInfo, error)
	List(ctx context.Context, groupID uint) ([]BasicInfo, error)
	Delete(ctx context.Context, appID uint) (*BasicInfo, error)
	Update(ctx context.Context, info *BasicInfo) (*BasicInfo, error)

	CreateOrUpdatePermission(ctx context.Context, appID uint, permissions []Permission) ([]Permission, error)
	GetPermission(ctx context.Context, appID uint) ([]Permission, error)

	CreateClientSecret(ctx context.Context, appID uint) (secret *models.OauthClientSecret, err error)
	ListClientSecret(ctx context.Context, appID uint) ([]models.OauthClientSecret, error)
	DeleteClientSecret(ctx context.Context, appID uint, secretID uint) (*models.OauthClientSecret, error)
}
