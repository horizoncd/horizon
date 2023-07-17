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
	"github.com/horizoncd/horizon/pkg/cluster/metrics/tekton"
	"github.com/horizoncd/horizon/pkg/pipelinerun/pipeline/dao"
	"github.com/horizoncd/horizon/pkg/pipelinerun/pipeline/models"
	"github.com/horizoncd/horizon/pkg/server/global"
)

const (
	_application = "application"
	_cluster     = "cluster"
	_updateAt    = "updated_at"
)

type Manager interface {
	Create(ctx context.Context, results *tekton.PipelineResults, data *global.HorizonMetaData) error
	ListPipelineStats(ctx context.Context, application, cluster string, pageNumber, pageSize int) (
		[]*models.PipelineStats, int64, error)
}

type manager struct {
	dao dao.DAO
}

func (m manager) ListPipelineStats(ctx context.Context, application, cluster string, pageNumber, pageSize int) (
	[]*models.PipelineStats, int64, error) {
	query := q.New(q.KeyWords{
		_application: application,
	})
	if cluster != "" {
		query.Keywords[_cluster] = cluster
	}
	query.PageNumber = pageNumber
	query.PageSize = pageSize
	// sort by updated_at desc default，let newer items be in head
	s := q.NewSort(_updateAt, true)
	query.Sorts = []*q.Sort{s}

	return m.dao.ListPipelineStats(ctx, query)
}

func (m manager) Create(ctx context.Context, results *tekton.PipelineResults, data *global.HorizonMetaData) error {
	return m.dao.Create(ctx, results, data)
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}
