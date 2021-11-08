package dao

import (
	"context"
	goerrors "errors"
	"net/http"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/common"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/util/errors"

	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, cluster *models.Cluster) (*models.Cluster, error)
	GetByID(ctx context.Context, id uint) (*models.Cluster, error)
	GetByName(ctx context.Context, clusterName string) (*models.Cluster, error)
	UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error)
	ListByApplicationAndEnv(ctx context.Context, applicationID uint, environment,
		filter string, query *q.Query) (int, []*models.ClusterWithEnvAndRegion, error)
	CheckClusterExists(ctx context.Context, cluster string) (bool, error)
	ListByNameFuzzily(context.Context, string, string, *q.Query) (int, []*models.ClusterWithEnvAndRegion, error)
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

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(cluster).Error; err != nil {
			return err
		}
		// insert a record to member table
		member := &membermodels.Member{
			ResourceType: membermodels.TypeApplicationCluster,
			ResourceID:   cluster.ID,
			Role:         role.Owner,
			MemberType:   membermodels.MemberUser,
			MemberNameID: cluster.CreatedBy,
			GrantedBy:    cluster.UpdatedBy,
		}
		result := tx.Create(member)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return goerrors.New("create member error")
		}
		return nil
	})

	return cluster, err
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

func (d *dao) GetByName(ctx context.Context, clusterName string) (*models.Cluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var cluster models.Cluster
	result := db.Raw(common.ClusterQueryByName, clusterName).Scan(&cluster)

	if result.Error != nil {
		return nil, err
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &cluster, nil
}

func (d *dao) UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error) {
	const op = "cluster dao: update by id"

	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var clusterInDB models.Cluster
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 1. get application in db first
		result := tx.Raw(common.ClusterQueryByID, id).Scan(&clusterInDB)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.E(op, http.StatusNotFound)
		}
		// 2. update value
		clusterInDB.Description = cluster.Description
		clusterInDB.GitURL = cluster.GitURL
		clusterInDB.GitSubfolder = cluster.GitSubfolder
		clusterInDB.GitBranch = cluster.GitBranch
		clusterInDB.TemplateRelease = cluster.TemplateRelease
		clusterInDB.UpdatedBy = cluster.UpdatedBy

		// 3. save application after updated
		tx.Save(&clusterInDB)

		return nil
	}); err != nil {
		return nil, err
	}
	return &clusterInDB, nil
}

func (d *dao) ListByApplicationAndEnv(ctx context.Context, applicationID uint, environment,
	filter string, query *q.Query) (int, []*models.ClusterWithEnvAndRegion, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, nil, err
	}

	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize

	like := "%" + filter + "%"
	var clusters []*models.ClusterWithEnvAndRegion
	result := db.Raw(common.ClusterQueryByApplicationAndEnv, applicationID,
		environment, like, limit, offset).Scan(&clusters)

	if result.Error != nil {
		return 0, nil, result.Error
	}

	var count int
	result = db.Raw(common.ClusterCountByApplicationAndEnv, applicationID, environment, like).Scan(&count)
	if result.Error != nil {
		return 0, nil, result.Error
	}

	return count, clusters, nil
}

func (d *dao) ListByNameFuzzily(ctx context.Context, environment, filter string,
	query *q.Query) (int, []*models.ClusterWithEnvAndRegion, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, nil, err
	}

	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize

	like := "%" + filter + "%"
	var (
		clusters []*models.ClusterWithEnvAndRegion
		count    int
		result   *gorm.DB
	)
	if environment != "" {
		result = db.Raw(common.ClusterQueryByEnvNameFuzzily,
			environment, like, limit, offset).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, result.Error
		}

		result = db.Raw(common.ClusterCountByEnvNameFuzzily, environment, like).Scan(&count)
		if result.Error != nil {
			return 0, nil, result.Error
		}
	} else {
		result = db.Raw(common.ClusterQueryByNameFuzzily,
			like, limit, offset).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, result.Error
		}

		result = db.Raw(common.ClusterCountByNameFuzzily, like).Scan(&count)
		if result.Error != nil {
			return 0, nil, result.Error
		}
	}

	return count, clusters, nil
}

func (d *dao) CheckClusterExists(ctx context.Context, cluster string) (bool, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return false, err
	}

	var c models.Cluster
	result := db.Raw(common.ClusterQueryByClusterName, cluster).Scan(&c)

	if result.Error != nil {
		return false, result.Error
	}

	if result.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}
