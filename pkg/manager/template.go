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

	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/dao"
	"github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/template/manager/manager_mock.go -package=mock_manager
type TemplateManager interface {
	// Create template
	Create(ctx context.Context, template *models.Template) (*models.Template, error)
	ListV2(ctx context.Context, query *q.Query, groupIDs ...uint) ([]*models.Template, error)
	// ListTemplate returns all template
	ListTemplate(ctx context.Context) ([]*models.Template, error)
	// ListByGroupID lists all template by group ID
	ListByGroupID(ctx context.Context, groupID uint) ([]*models.Template, error)
	// DeleteByID deletes template by ID
	DeleteByID(ctx context.Context, id uint) error
	// GetByID gets a template by ID
	GetByID(ctx context.Context, id uint) (*models.Template, error)
	// GetByName gets a template by name
	GetByName(ctx context.Context, name string) (*models.Template, error)
	GetRefOfApplication(ctx context.Context, id uint) ([]*models.Application, uint, error)
	GetRefOfCluster(ctx context.Context, id uint) ([]*models.Cluster, uint, error)
	UpdateByID(ctx context.Context, id uint, template *models.Template) error
	ListByGroupIDs(ctx context.Context, ids []uint) ([]*models.Template, error)
	ListByIDs(ctx context.Context, ids []uint) ([]*models.Template, error)
}

func NewTemplateManager(db *gorm.DB) TemplateManager {
	return &templateManager{templateDAO: dao.NewTemplateDAO(db)}
}

type templateManager struct {
	templateDAO dao.TemplateDAO
}

func (m *templateManager) Create(ctx context.Context, template *models.Template) (*models.Template, error) {
	return m.templateDAO.Create(ctx, template)
}

func (m *templateManager) ListTemplate(ctx context.Context) ([]*models.Template, error) {
	return m.templateDAO.ListTemplate(ctx)
}

func (m *templateManager) ListByGroupID(ctx context.Context, groupID uint) ([]*models.Template, error) {
	return m.templateDAO.ListByGroupID(ctx, groupID)
}

func (m *templateManager) DeleteByID(ctx context.Context, id uint) error {
	return m.templateDAO.DeleteByID(ctx, id)
}

func (m *templateManager) GetByID(ctx context.Context, id uint) (*models.Template, error) {
	return m.templateDAO.GetByID(ctx, id)
}

func (m *templateManager) GetByName(ctx context.Context, name string) (*models.Template, error) {
	return m.templateDAO.GetByName(ctx, name)
}

func (m *templateManager) GetRefOfApplication(ctx context.Context, id uint) ([]*models.Application, uint, error) {
	return m.templateDAO.GetRefOfApplication(ctx, id)
}

func (m *templateManager) GetRefOfCluster(ctx context.Context, id uint) ([]*models.Cluster, uint, error) {
	return m.templateDAO.GetRefOfCluster(ctx, id)
}

func (m *templateManager) UpdateByID(ctx context.Context, id uint, template *models.Template) error {
	return m.templateDAO.UpdateByID(ctx, id, template)
}

func (m *templateManager) ListByGroupIDs(ctx context.Context, ids []uint) ([]*models.Template, error) {
	return m.templateDAO.ListByGroupIDs(ctx, ids)
}

func (m *templateManager) ListByIDs(ctx context.Context, ids []uint) ([]*models.Template, error) {
	return m.templateDAO.ListByIDs(ctx, ids)
}

func (m *templateManager) ListV2(ctx context.Context, query *q.Query, groupIDs ...uint) ([]*models.Template, error) {
	return m.templateDAO.ListV2(ctx, query, groupIDs...)
}
