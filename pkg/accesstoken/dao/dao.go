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

package dao

import (
	"context"
	"fmt"

	"github.com/horizoncd/horizon/pkg/token/generator"
	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/accesstoken/models"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
)

type dao struct {
	db *gorm.DB
}

type DAO interface {
	ListAccessTokensByResource(ctx context.Context, resourceType string, resourceID uint,
		query *q.Query) ([]*models.AccessToken, int, error)
	ListPersonalAccessTokens(ctx context.Context, query *q.Query) ([]*models.AccessToken, int, error)
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{
		db: db,
	}
}

func (d *dao) ListAccessTokensByResource(ctx context.Context, resourceType string, resourceID uint,
	query *q.Query) ([]*models.AccessToken, int, error) {
	var (
		pageSize   = common.DefaultPageSize
		pageNumber = common.DefaultPageNumber
		tokens     []*models.AccessToken
		total      int64
	)
	if query != nil {
		if query.PageSize > 1 {
			pageSize = query.PageSize
		}
		if query.PageNumber > 0 {
			pageNumber = query.PageNumber
		}
	}
	limit := pageSize
	offset := (pageNumber - 1) * pageSize

	result := d.db.WithContext(ctx).Table("tb_token as t").
		Joins("join tb_user as u on t.user_id = u.id").
		Joins("join tb_member as m on u.id = m.membername_id").
		Where("u.user_type = ?", usermodels.UserTypeRobot).
		Where("m.resource_type = ?", resourceType).
		Where("m.resource_id = ?", resourceID).
		Select("t.*, m.role as role").Offset(offset).Limit(limit).Scan(&tokens).Offset(0).Limit(-1).Count(&total)
	return tokens, int(total), result.Error
}

func (d *dao) ListPersonalAccessTokens(ctx context.Context, query *q.Query) ([]*models.AccessToken, int, error) {
	var (
		pageSize   = common.DefaultPageSize
		pageNumber = common.DefaultPageNumber
		tokens     []*models.AccessToken
		total      int64
	)
	if query != nil {
		if query.PageSize > 1 {
			pageSize = query.PageSize
		}
		if query.PageNumber > 0 {
			pageNumber = query.PageNumber
		}
	}
	limit := pageSize
	offset := (pageNumber - 1) * pageSize
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, int(total), err
	}

	result := d.db.WithContext(ctx).Table("tb_token").
		Where("user_id = ?", currentUser.GetID()).
		Where("code like ?", fmt.Sprintf("%s%%", generator.AccessTokenPrefix)).
		Offset(offset).Limit(limit).
		Find(&tokens).Offset(0).Limit(-1).Count(&total)
	return tokens, int(total), result.Error
}
