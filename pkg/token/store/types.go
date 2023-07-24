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

package store

import (
	"context"

	"github.com/horizoncd/horizon/pkg/token/models"
)

type Store interface {
	Create(ctx context.Context, token *models.Token) (*models.Token, error)
	GetByID(ctx context.Context, id uint) (*models.Token, error)
	GetByCode(ctx context.Context, code string) (*models.Token, error)
	UpdateByID(ctx context.Context, id uint, token *models.Token) error
	DeleteByID(ctx context.Context, id uint) error
	DeleteByCode(ctx context.Context, code string) error
	DeleteByClientID(ctx context.Context, clientID string) error
}
