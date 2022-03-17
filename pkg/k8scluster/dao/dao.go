package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/k8scluster/models"
)

type DAO interface {
	// Create a k8sCluster
	Create(ctx context.Context, k8sCluster *models.K8SCluster) (*models.K8SCluster, error)
	// ListAll list all k8sClusters
	ListAll(ctx context.Context) ([]*models.K8SCluster, error)
	GetByServer(ctx context.Context, server string) (*models.K8SCluster, error)
}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(ctx context.Context, k8sCluster *models.K8SCluster) (*models.K8SCluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(k8sCluster)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.K8SClusterInDB, result.Error.Error())
	}

	return k8sCluster, result.Error
}

func (d *dao) ListAll(ctx context.Context) ([]*models.K8SCluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var k8sClusters []*models.K8SCluster
	result := db.Raw(common.K8SClusterListAll).Scan(&k8sClusters)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.K8SClusterInDB, result.Error.Error())
	}

	return k8sClusters, result.Error
}

func (d *dao) GetByServer(ctx context.Context, server string) (*models.K8SCluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var k8sCluster models.K8SCluster
	result := db.Raw(common.K8SClusterGetByServer, server).Scan(&k8sCluster)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.K8SClusterInDB, result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &k8sCluster, nil
}
