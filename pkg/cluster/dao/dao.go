package dao

import (
	"context"
	goerrors "errors"
	"fmt"
	"time"

	querycommon "g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	clustertagmodels "g.hz.netease.com/horizon/pkg/clustertag/models"
	"g.hz.netease.com/horizon/pkg/common"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	"gorm.io/gorm"
)

var (
	columnInTable = map[string]string{
		querycommon.Template:        "`c`.`template`",
		querycommon.TemplateRelease: "`c`.`template_release`",
	}
)

type DAO interface {
	Create(ctx context.Context, cluster *models.Cluster,
		clusterTags []*clustertagmodels.ClusterTag, extraMembers map[*usermodels.User]string) (*models.Cluster, error)
	GetByID(ctx context.Context, id uint) (*models.Cluster, error)
	GetByName(ctx context.Context, clusterName string) (*models.Cluster, error)
	UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error)
	DeleteByID(ctx context.Context, id uint) error
	ListByApplicationAndEnvs(ctx context.Context, applicationID uint, environments []string,
		filter string, query *q.Query) (int, []*models.ClusterWithEnvAndRegion, error)
	ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.Cluster, error)
	CheckClusterExists(ctx context.Context, cluster string) (bool, error)
	ListByNameFuzzily(context.Context, string, string, *q.Query) (int, []*models.ClusterWithEnvAndRegion, error)
	ListUserAuthorizedByNameFuzzily(ctx context.Context, environment,
		name string, applicationIDs []uint, userInfo uint, query *q.Query) (int, []*models.ClusterWithEnvAndRegion, error)
}

type dao struct {
}

func NewDAO() DAO {
	return &dao{}
}

func (d *dao) Create(ctx context.Context, cluster *models.Cluster,
	clusterTags []*clustertagmodels.ClusterTag, extraMembers map[*usermodels.User]string) (*models.Cluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(cluster).Error; err != nil {
			return herrors.NewErrInsertFailed(herrors.ClusterInDB, err.Error())
		}
		// insert records to member table
		members := make([]*membermodels.Member, 0)

		// the owner who created this cluster
		members = append(members, &membermodels.Member{
			ResourceType: membermodels.TypeApplicationCluster,
			ResourceID:   cluster.ID,
			Role:         role.Owner,
			MemberType:   membermodels.MemberUser,
			MemberNameID: currentUser.GetID(),
			GrantedBy:    currentUser.GetID(),
		})

		// the extra owners
		for extraMember, roleOfMember := range extraMembers {
			if extraMember.ID == currentUser.GetID() {
				continue
			}
			members = append(members, &membermodels.Member{
				ResourceType: membermodels.TypeApplicationCluster,
				ResourceID:   cluster.ID,
				Role:         roleOfMember,
				MemberType:   membermodels.MemberUser,
				MemberNameID: extraMember.ID,
				GrantedBy:    currentUser.GetID(),
			})
		}

		result := tx.Create(members)
		if result.Error != nil {
			return herrors.NewErrInsertFailed(herrors.ClusterInDB, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return herrors.NewErrInsertFailed(herrors.ClusterInDB, "create member error")
		}

		if len(clusterTags) == 0 {
			return nil
		}
		for i := 0; i < len(clusterTags); i++ {
			clusterTags[i].ClusterID = cluster.ID
		}

		result = tx.Create(clusterTags)
		if result.Error != nil {
			return herrors.NewErrInsertFailed(herrors.ClusterInDB, result.Error.Error())
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

	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.ClusterInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
	}

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
		return nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return nil, herrors.NewErrNotFound(herrors.ClusterInDB, fmt.Sprintf("no cluster named %s found", clusterName))
	}

	return &cluster, nil
}

func (d *dao) UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var clusterInDB models.Cluster
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 1. get application in db first
		result := tx.Raw(common.ClusterQueryByID, id).Scan(&clusterInDB)
		if result.Error != nil {
			return herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return herrors.NewErrNotFound(herrors.ClusterInDB, "row affected = 0")
		}
		// 2. update value
		clusterInDB.Description = cluster.Description
		clusterInDB.GitURL = cluster.GitURL
		clusterInDB.GitSubfolder = cluster.GitSubfolder
		clusterInDB.GitBranch = cluster.GitBranch
		clusterInDB.TemplateRelease = cluster.TemplateRelease
		clusterInDB.Status = cluster.Status
		clusterInDB.EnvironmentRegionID = cluster.EnvironmentRegionID

		// 3. save cluster after updated
		if err := tx.Save(&clusterInDB).Error; err != nil {
			return herrors.NewErrInsertFailed(herrors.ClusterInDB, result.Error.Error())
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return &clusterInDB, nil
}

func (d *dao) DeleteByID(ctx context.Context, id uint) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Exec(common.ClusterDeleteByID, time.Now().Unix(), currentUser.GetID(), id)

	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.ClusterInDB, result.Error.Error())
	}

	return nil
}

func (d *dao) ListByApplicationAndEnvs(ctx context.Context, applicationID uint, environments []string,
	filter string, query *q.Query) (int, []*models.ClusterWithEnvAndRegion, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, nil, err
	}

	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize

	like := "%" + filter + "%"
	var clusters []*models.ClusterWithEnvAndRegion

	var result *gorm.DB
	if len(environments) > 0 {
		result = db.Raw(common.ClusterQueryByApplicationAndEnvs, applicationID,
			environments, like, limit, offset).Scan(&clusters)
	} else {
		result = db.Raw(common.ClusterQueryByApplication, applicationID, like, limit, offset).Scan(&clusters)
	}

	if result.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
	}

	var count int
	if len(environments) > 0 {
		result = db.Raw(common.ClusterCountByApplicationAndEnvs, applicationID, environments, like).Scan(&count)
	} else {
		result = db.Raw(common.ClusterCountByApplication, applicationID, like).Scan(&count)
	}
	if result.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
	}

	return count, clusters, nil
}

func (d *dao) ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.Cluster, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var clusters []*models.Cluster
	result := db.Raw(common.ClusterListByApplicationID, applicationID).Scan(&clusters)

	if result.Error != nil {
		return nil, herrors.NewErrListFailed(herrors.ClusterInDB, result.Error.Error())
	}

	return clusters, nil
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
	whereCond, whereValues := orm.FormatFilterExp(query, columnInTable)
	var (
		clusters []*models.ClusterWithEnvAndRegion
		count    int
		result   *gorm.DB
	)
	if environment != "" {
		whereValuesForRecord := append([]interface{}(nil), whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, environment, like, limit, offset)

		result = db.Raw(fmt.Sprintf(common.ClusterQueryByEnvNameFuzzily, whereCond),
			whereValuesForRecord...).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}

		whereValuesForCount := append([]interface{}(nil), whereValues...)
		whereValuesForCount = append(whereValuesForCount, environment, like)

		result = db.Raw(fmt.Sprintf(common.ClusterCountByEnvNameFuzzily, whereCond), whereValuesForCount...).Scan(&count)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
	} else {
		whereValuesForRecord := append([]interface{}(nil), whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, like, limit, offset)

		result = db.Raw(fmt.Sprintf(common.ClusterQueryByNameFuzzily, whereCond),
			whereValuesForRecord...).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}

		whereValuesForCount := append([]interface{}(nil), whereValues...)
		whereValuesForCount = append(whereValuesForCount, like)

		result = db.Raw(fmt.Sprintf(common.ClusterCountByNameFuzzily, whereCond), whereValuesForCount...).Scan(&count)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
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
		return false, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

func (d *dao) ListUserAuthorizedByNameFuzzily(ctx context.Context, environment,
	name string, applicationIDs []uint, userInfo uint, query *q.Query) (int, []*models.ClusterWithEnvAndRegion, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, nil, err
	}

	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize

	like := "%" + name + "%"
	whereCond, whereValues := orm.FormatFilterExp(query, columnInTable)
	var (
		clusters []*models.ClusterWithEnvAndRegion
		count    int
		result   *gorm.DB
	)

	if len(environment) == 0 {
		whereValuesForRecord := append([]interface{}(nil), whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, userInfo, like)
		whereValuesForRecord = append(whereValuesForRecord, whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, applicationIDs, like, limit, offset)
		result = db.Raw(fmt.Sprintf(common.ClusterQueryByUserAndNameFuzzily, whereCond, whereCond),
			whereValuesForRecord...).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}

		whereValuesForCount := append([]interface{}(nil), whereValues...)
		whereValuesForCount = append(whereValuesForCount, userInfo, like)
		whereValuesForCount = append(whereValuesForCount, whereValues...)
		whereValuesForCount = append(whereValuesForCount, applicationIDs, like)
		result = db.Raw(fmt.Sprintf(common.ClusterCountByUserAndNameFuzzily, whereCond, whereCond),
			whereValuesForCount...).Scan(&count)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
	} else {
		whereValuesForRecord := append([]interface{}(nil), whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, userInfo, environment, like)
		whereValuesForRecord = append(whereValuesForRecord, whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, applicationIDs, environment, like, limit, offset)
		result = db.Raw(fmt.Sprintf(common.ClusterQueryByUserAndEnvAndNameFuzzily, whereCond, whereCond),
			whereValuesForRecord...).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}

		whereValuesForCount := append([]interface{}(nil), whereValues...)
		whereValuesForCount = append(whereValuesForCount, userInfo, environment, like)
		whereValuesForCount = append(whereValuesForCount, whereValues...)
		whereValuesForCount = append(whereValuesForCount, applicationIDs, environment, like)
		result = db.Raw(fmt.Sprintf(common.ClusterCountByUserAndEnvAndNameFuzzily, whereCond, whereCond),
			whereValuesForCount...).Scan(&count)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
	}

	return count, clusters, nil
}
