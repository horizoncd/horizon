package dao

import (
	"context"
	goerrors "errors"
	"fmt"
	"strings"
	"time"

	common "g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	sqlcommon "g.hz.netease.com/horizon/pkg/common"
	perror "g.hz.netease.com/horizon/pkg/errors"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"

	"gorm.io/gorm"
)

var (
	columnInTable = map[string]string{
		common.Template:        "`c`.`template`",
		common.TemplateRelease: "`c`.`template_release`",
	}
)

type DAO interface {
	Create(ctx context.Context, cluster *models.Cluster,
		tags []*tagmodels.Tag, extraMembers map[*usermodels.User]string) (*models.Cluster, error)
	GetByID(ctx context.Context, id uint) (*models.Cluster, error)
	GetByName(ctx context.Context, clusterName string) (*models.Cluster, error)
	UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error)
	DeleteByID(ctx context.Context, id uint) error
	ListByApplicationEnvsTags(ctx context.Context, applicationID uint, environments []string,
		filter string, query *q.Query, ts []tagmodels.TagSelector) (int, []*models.ClusterWithRegion, error)
	ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.Cluster, error)
	CheckClusterExists(ctx context.Context, cluster string) (bool, error)
	ListByNameFuzzily(context.Context, string, string, *q.Query) (int, []*models.ClusterWithRegion, error)
	ListUserAuthorizedByNameFuzzily(ctx context.Context, environment,
		name string, applicationIDs []uint, userInfo uint, query *q.Query) (int, []*models.ClusterWithRegion, error)
}

type dao struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d *dao) Create(ctx context.Context, cluster *models.Cluster,
	tags []*tagmodels.Tag, extraMembers map[*usermodels.User]string) (*models.Cluster, error) {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

		if len(tags) == 0 {
			return nil
		}
		for i := 0; i < len(tags); i++ {
			tags[i].ResourceType = tagmodels.TypeCluster
			tags[i].ResourceID = cluster.ID
		}

		result = tx.Create(tags)
		if result.Error != nil {
			return herrors.NewErrInsertFailed(herrors.ClusterInDB, result.Error.Error())
		}

		return nil
	})

	return cluster, err
}

func (d *dao) GetByID(ctx context.Context, id uint) (*models.Cluster, error) {
	var cluster models.Cluster
	result := d.db.WithContext(ctx).Raw(sqlcommon.ClusterQueryByID, id).First(&cluster)

	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.ClusterInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
	}

	return &cluster, result.Error
}

func (d *dao) GetByName(ctx context.Context, clusterName string) (*models.Cluster, error) {
	var cluster models.Cluster
	result := d.db.WithContext(ctx).Raw(sqlcommon.ClusterQueryByName, clusterName).Scan(&cluster)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return nil, herrors.NewErrNotFound(herrors.ClusterInDB, fmt.Sprintf("no cluster named %s found", clusterName))
	}

	return &cluster, nil
}

func (d *dao) UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error) {
	var clusterInDB models.Cluster
	if err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. get application in db first
		result := tx.Raw(sqlcommon.ClusterQueryByID, id).Scan(&clusterInDB)
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
		clusterInDB.EnvironmentName = cluster.EnvironmentName
		clusterInDB.RegionName = cluster.RegionName

		// 3. save cluster after updated
		if err := tx.Save(&clusterInDB).Error; err != nil {
			return herrors.NewErrUpdateFailed(herrors.ClusterInDB, err.Error())
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return &clusterInDB, nil
}

func (d *dao) DeleteByID(ctx context.Context, id uint) error {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	result := d.db.WithContext(ctx).Exec(sqlcommon.ClusterDeleteByID, time.Now().Unix(), currentUser.GetID(), id)

	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.ClusterInDB, result.Error.Error())
	}

	return nil
}

func (d *dao) ListByApplicationEnvsTags(ctx context.Context, applicationID uint, environments []string,
	filter string, query *q.Query, ts []tagmodels.TagSelector) (int, []*models.ClusterWithRegion, error) {
	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize

	like := "%" + filter + "%"
	var clusters []*models.ClusterWithRegion

	var result *gorm.DB
	var count int

	if len(ts) > 0 {
		// todo: support other operators
		var conditions []string
		var params []interface{}
		params = append(params, applicationID, tagmodels.TypeCluster, like, environments, environments)
		for _, tagSelector := range ts {
			if tagSelector.Operator != tagmodels.Equals && tagSelector.Operator != tagmodels.In {
				return 0, nil, perror.Wrapf(herrors.ErrParamInvalid,
					fmt.Sprintf("this operator %s is not supported yet", tagSelector.Operator))
			}
			conditions = append(conditions, "(tg.tag_key = ? and tg.tag_value in ?)")
			params = append(params, tagSelector.Key, tagSelector.Values.List())
		}
		params = append(params, len(ts), limit, offset)
		tagCondition := strings.Join(conditions, " or ")
		tagCondition = fmt.Sprintf("(%s)", tagCondition)
		result = d.db.WithContext(ctx).Raw(fmt.Sprintf(sqlcommon.ClusterQueryByApplicationAndTags,
			tagCondition), params...).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
		result = d.db.WithContext(ctx).Raw(fmt.Sprintf(sqlcommon.ClusterCountByApplicationAndTags,
			tagCondition), params...).Scan(&count)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
	} else {
		result = d.db.WithContext(ctx).Raw(sqlcommon.ClusterQueryByApplication, applicationID, like,
			environments, environments, limit, offset).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
		result = d.db.WithContext(ctx).Raw(sqlcommon.ClusterCountByApplication, applicationID, like).Scan(&count)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
	}

	return count, clusters, nil
}

func (d *dao) ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.Cluster, error) {
	var clusters []*models.Cluster
	result := d.db.WithContext(ctx).Raw(sqlcommon.ClusterListByApplicationID, applicationID).Scan(&clusters)

	if result.Error != nil {
		return nil, herrors.NewErrListFailed(herrors.ClusterInDB, result.Error.Error())
	}

	return clusters, nil
}

func (d *dao) ListByNameFuzzily(ctx context.Context, environment, filter string,
	query *q.Query) (int, []*models.ClusterWithRegion, error) {
	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize

	like := "%" + filter + "%"
	whereCond, whereValues := orm.FormatFilterExp(query, columnInTable)
	var (
		clusters []*models.ClusterWithRegion
		count    int
		result   *gorm.DB
	)
	if environment != "" {
		whereValuesForRecord := append([]interface{}(nil), whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, environment, like, limit, offset)

		result = d.db.WithContext(ctx).Raw(fmt.Sprintf(sqlcommon.ClusterQueryByEnvNameFuzzily, whereCond),
			whereValuesForRecord...).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}

		whereValuesForCount := append([]interface{}(nil), whereValues...)
		whereValuesForCount = append(whereValuesForCount, environment, like)

		result = d.db.WithContext(ctx).Raw(fmt.Sprintf(sqlcommon.ClusterCountByEnvNameFuzzily, whereCond),
			whereValuesForCount...).Scan(&count)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
	} else {
		whereValuesForRecord := append([]interface{}(nil), whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, like, limit, offset)

		result = d.db.WithContext(ctx).Raw(fmt.Sprintf(sqlcommon.ClusterQueryByNameFuzzily, whereCond),
			whereValuesForRecord...).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}

		whereValuesForCount := append([]interface{}(nil), whereValues...)
		whereValuesForCount = append(whereValuesForCount, like)

		result = d.db.WithContext(ctx).Raw(fmt.Sprintf(sqlcommon.ClusterCountByNameFuzzily, whereCond),
			whereValuesForCount...).Scan(&count)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
	}

	return count, clusters, nil
}

func (d *dao) CheckClusterExists(ctx context.Context, cluster string) (bool, error) {
	var c models.Cluster
	result := d.db.WithContext(ctx).Raw(sqlcommon.ClusterQueryByClusterName, cluster).Scan(&c)

	if result.Error != nil {
		return false, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

func (d *dao) ListUserAuthorizedByNameFuzzily(ctx context.Context, environment,
	name string, applicationIDs []uint, userInfo uint, query *q.Query) (int, []*models.ClusterWithRegion, error) {
	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize

	like := "%" + name + "%"
	whereCond, whereValues := orm.FormatFilterExp(query, columnInTable)
	var (
		clusters []*models.ClusterWithRegion
		count    int
		result   *gorm.DB
	)

	if len(environment) == 0 {
		whereValuesForRecord := append([]interface{}(nil), whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, userInfo, like)
		whereValuesForRecord = append(whereValuesForRecord, whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, applicationIDs, like, limit, offset)
		result = d.db.WithContext(ctx).Raw(fmt.Sprintf(sqlcommon.ClusterQueryByUserAndNameFuzzily, whereCond, whereCond),
			whereValuesForRecord...).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}

		whereValuesForCount := append([]interface{}(nil), whereValues...)
		whereValuesForCount = append(whereValuesForCount, userInfo, like)
		whereValuesForCount = append(whereValuesForCount, whereValues...)
		whereValuesForCount = append(whereValuesForCount, applicationIDs, like)
		result = d.db.WithContext(ctx).Raw(fmt.Sprintf(sqlcommon.ClusterCountByUserAndNameFuzzily, whereCond, whereCond),
			whereValuesForCount...).Scan(&count)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
	} else {
		whereValuesForRecord := append([]interface{}(nil), whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, userInfo, environment, like)
		whereValuesForRecord = append(whereValuesForRecord, whereValues...)
		whereValuesForRecord = append(whereValuesForRecord, applicationIDs, environment, like, limit, offset)
		result = d.db.WithContext(ctx).Raw(fmt.Sprintf(sqlcommon.ClusterQueryByUserAndEnvAndNameFuzzily,
			whereCond, whereCond), whereValuesForRecord...).Scan(&clusters)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}

		whereValuesForCount := append([]interface{}(nil), whereValues...)
		whereValuesForCount = append(whereValuesForCount, userInfo, environment, like)
		whereValuesForCount = append(whereValuesForCount, whereValues...)
		whereValuesForCount = append(whereValuesForCount, applicationIDs, environment, like)
		result = d.db.WithContext(ctx).Raw(fmt.Sprintf(sqlcommon.ClusterCountByUserAndEnvAndNameFuzzily,
			whereCond, whereCond), whereValuesForCount...).Scan(&count)
		if result.Error != nil {
			return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, result.Error.Error())
		}
	}

	return count, clusters, nil
}
