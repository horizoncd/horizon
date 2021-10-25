package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/common"
)

type DAO interface {
	Create(ctx context.Context, cluster *models.Cluster) (*models.Cluster, error)
	GetByID(ctx context.Context, id uint) (*models.Cluster, error)
	ListByApplication(ctx context.Context, applicationID uint) ([]*models.Cluster, error)
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

func (d *dao) GetByID(ctx context.Context, id uint) (*models.Cluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var cluster models.Cluster
	result := db.Raw(common.ClusterQueryByID, id).First(&cluster)

	return &cluster, result.Error
}

func (d *dao) ListByApplication(ctx context.Context, applicationID uint) ([]*models.Cluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var clusters []*models.Cluster
	result := db.Raw(common.ClusterQueryByApplication, applicationID).Scan(&clusters)

	return clusters, result.Error
}
