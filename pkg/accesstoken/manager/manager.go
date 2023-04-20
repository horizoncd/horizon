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

package manager

import (
	"context"

	"gorm.io/gorm"

	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/accesstoken/dao"
	"github.com/horizoncd/horizon/pkg/accesstoken/models"
)

type Manager interface {
	ListAccessTokensByResource(context.Context, string, uint, *q.Query) ([]*models.AccessToken, int, error)
	ListPersonalAccessTokens(context.Context, *q.Query) ([]*models.AccessToken, int, error)
}

type manager struct {
	dao dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}

func (m *manager) ListAccessTokensByResource(ctx context.Context, resourceType string,
	resourceID uint, query *q.Query) ([]*models.AccessToken, int, error) {
	return m.dao.ListAccessTokensByResource(ctx, resourceType, resourceID, query)
}

func (m *manager) ListPersonalAccessTokens(ctx context.Context, query *q.Query) ([]*models.AccessToken, int, error) {
	return m.dao.ListPersonalAccessTokens(ctx, query)
}
