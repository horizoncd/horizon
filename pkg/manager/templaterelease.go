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

	"github.com/horizoncd/horizon/pkg/dao"
	"github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"gorm.io/gorm"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/templaterelease/manager/manager_mock.go -package=mock_manager
type TemplateReleaseManager interface {
	// Create template release
	Create(ctx context.Context, templateRelease *models.TemplateRelease) (*models.TemplateRelease, error)
	// ListByTemplateName list all releases by template name
	ListByTemplateName(ctx context.Context, templateName string) ([]*models.TemplateRelease, error)
	// ListByTemplateID list all releases by template ID
	ListByTemplateID(ctx context.Context, id uint) ([]*models.TemplateRelease, error)
	// GetByTemplateNameAndRelease get release by template name and release name
	GetByTemplateNameAndRelease(ctx context.Context, templateName, release string) (*models.TemplateRelease, error)
	// GetByID gets release by releaseID
	GetByID(ctx context.Context, releaseID uint) (*models.TemplateRelease, error)
	GetRefOfApplication(ctx context.Context, id uint) ([]*models.Application, uint, error)
	GetRefOfCluster(ctx context.Context, id uint) ([]*models.Cluster, uint, error)
	UpdateByID(ctx context.Context, releaseID uint, release *models.TemplateRelease) error
	DeleteByID(ctx context.Context, id uint) error
}

func NewTemplateReleaseManager(db *gorm.DB) TemplateReleaseManager {
	return &templateReleaseManager{dao: dao.NewTemplateReleaseDAO(db)}
}

type templateReleaseManager struct {
	dao dao.TemplateReleaseDAO
}

func (m *templateReleaseManager) Create(ctx context.Context,
	templateRelease *models.TemplateRelease) (*models.TemplateRelease, error) {
	return m.dao.Create(ctx, templateRelease)
}

func (m *templateReleaseManager) ListByTemplateName(ctx context.Context, templateName string) ([]*models.TemplateRelease, error) {
	return m.dao.ListByTemplateName(ctx, templateName)
}
func (m *templateReleaseManager) ListByTemplateID(ctx context.Context, id uint) ([]*models.TemplateRelease, error) {
	return m.dao.ListByTemplateID(ctx, id)
}

func (m *templateReleaseManager) GetByTemplateNameAndRelease(ctx context.Context,
	templateName, release string) (_ *models.TemplateRelease, err error) {
	const op = "template release templateReleaseManager: get by template name and release"
	defer wlog.Start(ctx, op).StopPrint()

	tr, err := m.dao.GetByTemplateNameAndRelease(ctx, templateName, release)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

func (m *templateReleaseManager) GetByID(ctx context.Context,
	releaseID uint) (*models.TemplateRelease, error) {
	return m.dao.GetByID(ctx, releaseID)
}

func (m *templateReleaseManager) GetRefOfApplication(ctx context.Context, id uint) ([]*models.Application, uint, error) {
	return m.dao.GetRefOfApplication(ctx, id)
}
func (m *templateReleaseManager) GetRefOfCluster(ctx context.Context, id uint) ([]*models.Cluster, uint, error) {
	return m.dao.GetRefOfCluster(ctx, id)
}

func (m *templateReleaseManager) UpdateByID(ctx context.Context, releaseID uint, release *models.TemplateRelease) error {
	return m.dao.UpdateByID(ctx, releaseID, release)
}

func (m *templateReleaseManager) DeleteByID(ctx context.Context, id uint) error {
	return m.dao.DeleteByID(ctx, id)
}
