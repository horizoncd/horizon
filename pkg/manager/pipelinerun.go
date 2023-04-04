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
// -package=mock_manager
//
//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/pipelinerun/manager/mock_manager.go
type PipelineRunManager interface {
	Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error)
	GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error)
	GetByCIEventID(ctx context.Context, ciEventID string) (*models.Pipelinerun, error)
	GetByClusterID(ctx context.Context, clusterID uint, canRollback bool,
		query q.Query) (int, []*models.Pipelinerun, error)
	GetFirstCanRollbackPipelinerun(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
	DeleteByID(ctx context.Context, pipelinerunID uint) error
	DeleteByClusterID(ctx context.Context, clusterID uint) error
	UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error
	GetLatestByClusterIDAndActions(ctx context.Context, clusterID uint, actions ...string) (*models.Pipelinerun, error)
	GetLatestByClusterIDAndActionAndStatus(ctx context.Context, clusterID uint, action,
		status string) (*models.Pipelinerun, error)
	GetLatestSuccessByClusterID(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
	UpdateStatusByID(ctx context.Context, pipelinerunID uint, result models.PipelineStatus) error
	UpdateCIEventIDByID(ctx context.Context, pipelinerunID uint, ciEventID string) error
	// UpdateResultByID  update the pipelinerun restore result
	UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error
}

type pipelineRunManager struct {
	dao dao.PipelineRunDAO
}

func NewPipelineRunManager(db *gorm.DB) PipelineRunManager {
	return &pipelineRunManager{
		dao: dao.NewPipelineRunDAO(db),
	}
}

func (m *pipelineRunManager) Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error) {
	return m.dao.Create(ctx, pipelinerun)
}

func (m *pipelineRunManager) GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error) {
	return m.dao.GetByID(ctx, pipelinerunID)
}

func (m *pipelineRunManager) GetByCIEventID(ctx context.Context, ciEventID string) (*models.Pipelinerun, error) {
	return m.dao.GetByCIEventID(ctx, ciEventID)
}

func (m *pipelineRunManager) DeleteByID(ctx context.Context, pipelinerunID uint) error {
	return m.dao.DeleteByID(ctx, pipelinerunID)
}

func (m *pipelineRunManager) DeleteByClusterID(ctx context.Context, clusterID uint) error {
	return m.dao.DeleteByClusterID(ctx, clusterID)
}

func (m *pipelineRunManager) UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error {
	return m.dao.UpdateConfigCommitByID(ctx, pipelinerunID, commit)
}

func (m *pipelineRunManager) GetLatestByClusterIDAndActions(ctx context.Context,
	clusterID uint, actions ...string) (*models.Pipelinerun, error) {
	return m.dao.GetLatestByClusterIDAndActions(ctx, clusterID, actions...)
}

func (m *pipelineRunManager) GetLatestSuccessByClusterID(ctx context.Context,
	clusterID uint) (*models.Pipelinerun, error) {
	return m.dao.GetLatestSuccessByClusterID(ctx, clusterID)
}

func (m *pipelineRunManager) UpdateStatusByID(ctx context.Context,
	pipelinerunID uint, result models.PipelineStatus) error {
	return m.dao.UpdateStatusByID(ctx, pipelinerunID, result)
}

func (m *pipelineRunManager) UpdateCIEventIDByID(ctx context.Context, pipelinerunID uint, ciEventID string) error {
	return m.dao.UpdateCIEventIDByID(ctx, pipelinerunID, ciEventID)
}

func (m *pipelineRunManager) UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error {
	return m.dao.UpdateResultByID(ctx, pipelinerunID, result)
}

func (m *pipelineRunManager) GetByClusterID(ctx context.Context,
	clusterID uint, canRollback bool, query q.Query) (int, []*models.Pipelinerun, error) {
	return m.dao.GetByClusterID(ctx, clusterID, canRollback, query)
}

func (m *pipelineRunManager) GetFirstCanRollbackPipelinerun(ctx context.Context,
	clusterID uint) (*models.Pipelinerun, error) {
	return m.dao.GetFirstCanRollbackPipelinerun(ctx, clusterID)
}

func (m *pipelineRunManager) GetLatestByClusterIDAndActionAndStatus(ctx context.Context, clusterID uint, action,
	status string) (*models.Pipelinerun, error) {
	return m.dao.GetLatestByClusterIDAndActionAndStatus(ctx, clusterID, action, status)
}
