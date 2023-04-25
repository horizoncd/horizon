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

package storage

import (
	"context"
	goerrors "errors"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/common"
	"github.com/horizoncd/horizon/pkg/token/models"
	"gorm.io/gorm"
)

type storage struct {
	db *gorm.DB
}

func NewStorage(db *gorm.DB) Storage {
	return &storage{db: db}
}

func (d *storage) Create(ctx context.Context, token *models.Token) (*models.Token, error) {
	result := d.db.WithContext(ctx).Create(token)
	return token, result.Error
}

func (d *storage) GetByID(ctx context.Context, id uint) (*models.Token, error) {
	var token models.Token
	result := d.db.WithContext(ctx).Model(token).Where("id = ?", id).First(&token)
	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.TokenInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.TokenInDB, result.Error.Error())
	}
	return &token, nil
}

func (d *storage) GetByCode(ctx context.Context, code string) (*models.Token, error) {
	var token models.Token
	result := d.db.WithContext(ctx).Model(token).Where("code = ?", code).First(&token)
	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.TokenInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.TokenInDB, result.Error.Error())
	}
	return &token, nil
}

func (d *storage) DeleteByID(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteTokenByID, id)
	return result.Error
}

func (d *storage) DeleteByCode(ctx context.Context, code string) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteByCode, code)
	return result.Error
}

func (d *storage) DeleteByClientID(ctx context.Context, clientID string) error {
	result := d.db.WithContext(ctx).Exec(common.DeleteByClientID, clientID)
	return result.Error
}
