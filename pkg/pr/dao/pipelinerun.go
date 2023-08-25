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

	"gorm.io/gorm"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/common"
	"github.com/horizoncd/horizon/pkg/pr/models"
)

type PipelineRunDAO interface {
	// Create create a pipelinerun
	Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error)
	GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error)
	GetByCIEventID(ctx context.Context, ciEventID string) (*models.Pipelinerun, error)
	GetByClusterID(ctx context.Context, clusterID uint,
		canRollback bool, query q.Query) (int, []*models.Pipelinerun, error)
	// DeleteByID delete pipelinerun by id
	DeleteByID(ctx context.Context, pipelinerunID uint) error
	DeleteByClusterID(ctx context.Context, clusterID uint) error
	UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error
	GetLatestByClusterIDAndActions(ctx context.Context, clusterID uint, action ...string) (*models.Pipelinerun, error)
	GetLatestByClusterIDAndActionAndStatus(ctx context.Context, clusterID uint,
		action, status string) (*models.Pipelinerun, error)
	UpdateStatusByID(ctx context.Context, pipelinerunID uint, result models.PipelineStatus) error
	UpdateCIEventIDByID(ctx context.Context, pipelinerunID uint, ciEventID string) error
	UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error
	GetLatestSuccessByClusterID(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
	GetFirstCanRollbackPipelinerun(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
	UpdateColumns(ctx context.Context, id uint, columns map[string]interface{}) error
}

type pipelinerunDAO struct{ db *gorm.DB }

func NewPipelineRunDAO(db *gorm.DB) PipelineRunDAO {
	return &pipelinerunDAO{db: db}
}

func (d *pipelinerunDAO) Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error) {
	result := d.db.WithContext(ctx).Create(pipelinerun)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.PipelinerunInDB, result.Error.Error())
	}

	return pipelinerun, result.Error
}

func (d *pipelinerunDAO) GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error) {
	var pr models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetByID, pipelinerunID).Scan(&pr)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pr, nil
}

func (d *pipelinerunDAO) GetByCIEventID(ctx context.Context, ciEventID string) (*models.Pipelinerun, error) {
	var pr models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetByCIEventID, ciEventID).Scan(&pr)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pr, nil
}

func (d *pipelinerunDAO) DeleteByClusterID(ctx context.Context, clusterID uint) error {
	result := d.db.WithContext(ctx).Exec(common.PipelinerunDeleteByClusterID, clusterID)

	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.PipelinerunInDB, result.Error.Error())
	}

	return result.Error
}

func (d *pipelinerunDAO) DeleteByID(ctx context.Context, pipelinerunID uint) error {
	result := d.db.WithContext(ctx).Exec(common.PipelinerunDeleteByID, pipelinerunID)

	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.PipelinerunInDB, result.Error.Error())
	}

	return result.Error
}

func (d *pipelinerunDAO) UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error {
	result := d.db.WithContext(ctx).Exec(common.PipelinerunUpdateConfigCommitByID, commit, pipelinerunID)

	if result.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	return result.Error
}

func (d *pipelinerunDAO) GetLatestByClusterIDAndActions(ctx context.Context,
	clusterID uint, actions ...string) (*models.Pipelinerun, error) {
	var pipelinerun models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetLatestByClusterIDAndActions,
		clusterID, actions).Scan(&pipelinerun)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}

func (d *pipelinerunDAO) GetLatestByClusterIDAndActionAndStatus(ctx context.Context,
	clusterID uint, action string, status string) (*models.Pipelinerun, error) {
	var pipelinerun models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetLatestByClusterIDAndActionAndStatus, clusterID,
		action, status).Scan(&pipelinerun)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}

func (d *pipelinerunDAO) GetLatestSuccessByClusterID(ctx context.Context, clusterID uint) (*models.Pipelinerun, error) {
	var pipelinerun models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetLatestSuccessByClusterID, clusterID).Scan(&pipelinerun)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}

func (d *pipelinerunDAO) UpdateStatusByID(ctx context.Context, pipelinerunID uint, status models.PipelineStatus) error {
	return d.UpdateColumns(ctx, pipelinerunID, map[string]interface{}{"status": string(status)})
}

func (d *pipelinerunDAO) UpdateCIEventIDByID(ctx context.Context, pipelinerunID uint, ciEventID string) error {
	return d.UpdateColumns(ctx, pipelinerunID, map[string]interface{}{"ci_event_id": ciEventID})
}

func (d *pipelinerunDAO) UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error {
	res := d.db.WithContext(ctx).Exec(common.PipelinerunUpdateResultByID, result.Result, result.S3Bucket,
		result.LogObject, result.PrObject, result.StartedAt, result.FinishedAt, pipelinerunID)

	if res.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.PipelinerunInDB, res.Error.Error())
	}
	return res.Error
}

func (d *pipelinerunDAO) GetByClusterID(ctx context.Context, clusterID uint,
	canRollback bool, query q.Query) (int, []*models.Pipelinerun, error) {
	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize

	var pipelineruns []*models.Pipelinerun
	queryScript := common.PipelinerunGetByClusterID
	countScript := common.PipelinerunGetByClusterIDTotalCount
	if canRollback {
		// remove the first canRollback pipelinerun
		offset++
		queryScript = common.PipelinerunCanRollbackGetByClusterID
		countScript = common.PipelinerunCanRollbackGetByClusterIDTotalCount
	}
	result := d.db.WithContext(ctx).Raw(queryScript,
		clusterID, limit, offset).Scan(&pipelineruns)
	if result.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	var total int
	result = d.db.WithContext(ctx).Raw(countScript,
		clusterID).Scan(&total)

	if total < 0 {
		total = 0
	}

	if result.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}

	return total, pipelineruns, result.Error
}

func (d *pipelinerunDAO) GetFirstCanRollbackPipelinerun(ctx context.Context,
	clusterID uint) (*models.Pipelinerun, error) {
	var pipelinerun models.Pipelinerun
	result := d.db.WithContext(ctx).Raw(common.PipelinerunGetFirstCanRollbackByClusterID, clusterID).Scan(&pipelinerun)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.PipelinerunInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pipelinerun, nil
}

func (d *pipelinerunDAO) UpdateColumns(ctx context.Context, id uint, columns map[string]interface{}) error {
	res := d.db.WithContext(ctx).Model(models.Pipelinerun{}).
		Where("id = ?", id).Updates(columns)
	if res.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.PipelinerunInDB, res.Error.Error())
	}
	return res.Error
}
