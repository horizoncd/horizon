package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/common"
)

type DAO interface {
	Create(ctx context.Context, cluster *models.Cluster) (*models.Cluster, error)
	GetByName(ctx context.Context, name string) (*models.Cluster, error)
	ListByApplication(ctx context.Context, application string) ([]*models.Cluster, error)
}

type dao struct {
}

func NewDAO() DAO {
	return &dao{}
}

func (d *dao) Create(ctx context.Context, cluster *models.Cluster) (*models.Cluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(cluster)

	return cluster, result.Error
}

func (d *dao) GetByName(ctx context.Context, name string) (*models.Cluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var cluster models.Cluster
	result := db.Raw(common.ClusterQueryByName, name).First(&cluster)

	return &cluster, result.Error
}

func (d *dao) ListByApplication(ctx context.Context, application string) ([]*models.Cluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var clusters []*models.Cluster
	result := db.Raw(common.ClusterQueryByApplication, application).Scan(&clusters)

	return clusters, result.Error
}
