package manager

import (
	"context"

	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
	"g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/dao"
	"g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/models"
	"g.hz.netease.com/horizon/pkg/server/global"
	"gorm.io/gorm"
)

const (
	_application = "application"
	_cluster     = "cluster"
	_updateAt    = "updated_at"
)

type Manager interface {
	Create(ctx context.Context, results *metrics.PipelineResults, data *global.HorizonMetaData) error
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

func (m manager) Create(ctx context.Context, results *metrics.PipelineResults, data *global.HorizonMetaData) error {
	return m.dao.Create(ctx, results, data)
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}
