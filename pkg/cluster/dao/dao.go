package dao

import (
	"context"
	goerrors "errors"
	"fmt"
	"time"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	sqlcommon "g.hz.netease.com/horizon/pkg/common"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"

	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, cluster *models.Cluster,
		tags []*tagmodels.Tag, extraMembers map[*usermodels.User]string) (*models.Cluster, error)
	GetByID(ctx context.Context, id uint, includeSoftDelete bool) (*models.Cluster, error)
	GetByName(ctx context.Context, clusterName string) (*models.Cluster, error)
	UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error)
	DeleteByID(ctx context.Context, id uint) error
	CheckClusterExists(ctx context.Context, cluster string) (bool, error)
	List(ctx context.Context, query *q.Query, withRegion bool, appIDs ...uint) (int, []*models.ClusterWithRegion, error)
	ListClusterWithExpiry(ctx context.Context, query *q.Query) ([]*models.Cluster, error)
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
			tags[i].ResourceType = common.ResourceCluster
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

func (d *dao) GetByID(ctx context.Context, id uint, includeSoftDelete bool) (*models.Cluster, error) {
	var cluster models.Cluster
	statement := d.db.Unscoped().WithContext(ctx).Where("id = ?", id)
	if !includeSoftDelete {
		statement.Where("deleted_ts = 0")
	}
	result := statement.First(&cluster)
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
		clusterInDB.GitRefType = cluster.GitRefType
		clusterInDB.GitRef = cluster.GitRef
		clusterInDB.TemplateRelease = cluster.TemplateRelease
		clusterInDB.Status = cluster.Status
		clusterInDB.EnvironmentName = cluster.EnvironmentName
		clusterInDB.RegionName = cluster.RegionName
		clusterInDB.ExpireSeconds = cluster.ExpireSeconds

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

func (d *dao) List(ctx context.Context,
	query *q.Query, withRegion bool, appIDs ...uint) (int, []*models.ClusterWithRegion, error) {
	var (
		clusters []*models.ClusterWithRegion
		total    int64
	)

	statement := d.db.WithContext(ctx).
		Table("tb_cluster as c").
		Select("c.*")
	if withRegion {
		statement.
			Select("c.*, r.display_name as region_display_name").
			Joins("join tb_region as r on r.name = c.region_name").
			Where("r.deleted_ts = 0")
	}

	genSQL := func(statement *gorm.DB, query *q.Query) *gorm.DB {
		for k, v := range query.Keywords {
			switch k {
			case common.ClusterQueryTagSelector:
				if tagSelectors, ok := v.([]tagmodels.TagSelector); ok {
					statementSubQuery := d.db.WithContext(ctx)
					statementSubQuery = GenSQLForTagSelector(tagSelectors, statementSubQuery)
					statement.Where("c.id in (?)", statementSubQuery)
				}
			case common.ClusterQueryEnvironment:
				if _, ok := v.(string); ok {
					statement = statement.Where("c.environment_name = ?", v)
				} else {
					statement = statement.Where("c.environment_name in ?", v)
				}
			case common.ParamApplicationID:
				statement = statement.Where("c.application_id = ?", v)
			case common.ClusterQueryName:
				statement = statement.Where("c.name like ?", fmt.Sprintf("%%%v%%", v))
			case common.ClusterQueryByUser:
				statement = statement.
					Joins("join tb_member as m on m.resource_id = c.id").
					Where("m.resource_type = ?", common.ResourceCluster).
					Where("m.member_type = '0'").
					Where("m.deleted_ts = 0").
					Where("m.membername_id = ?", v)
			case common.ClusterQueryByTemplate:
				statement = statement.Where("c.template = ?", v)
			case common.ClusterQueryByRelease:
				statement = statement.Where("c.template_release = ?", v)
			}
		}
		statement = statement.Where("c.deleted_ts = 0")
		return statement
	}

	if query != nil {
		statement = genSQL(statement, query)

		if len(appIDs) > 0 &&
			query.Keywords != nil &&
			query.Keywords[common.ClusterQueryByUser] != nil {
			statementGroup := d.db.WithContext(ctx).
				Table("tb_cluster as c").
				Select("c.*")
			if withRegion {
				statementGroup.
					Select("c.*, r.display_name as region_display_name").
					Joins("join tb_region as r on r.name = c.region_name").
					Where("r.deleted_ts = 0")
			}

			delete(query.Keywords, common.ClusterQueryByUser)
			statementGroup = genSQL(statementGroup, query)

			statementGroup = statementGroup.Where("application_id in ?", appIDs)
			statement = d.db.Raw("? union ?", statement, statementGroup)
		}
	}

	res := d.db.Raw("select count(id) from (?) as clusters", statement).Scan(&total)

	if res.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, res.Error.Error())
	}

	statement = d.db.Raw("select * from (?) as apps order by updated_at desc limit ? offset ?",
		statement, query.Limit(), query.Offset())
	res = statement.Scan(&clusters)
	if res.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.ClusterInDB, res.Error.Error())
	}

	return int(total), clusters, nil
}

func GenSQLForTagSelector(tagSelectors []tagmodels.TagSelector, statement *gorm.DB) *gorm.DB {
	condition := statement.WithContext(context.Background())
	statement = statement.Table("tb_tag as tg").
		Select("tg.resource_id").
		Where("tg.resource_type = ?", common.ResourceCluster).
		Group("tg.resource_id").
		Having("count(tg.id) > ?", len(tagSelectors)-1)

	for i, tag := range tagSelectors {
		switch tag.Operator {
		case tagmodels.Equals:
			value, exist := tag.Values.PopAny()
			if exist {
				if i == 0 {
					statement.Where(condition.Where("tg.tag_key = ?", tag.Key).Where("tg.tag_value = ?", value))
				} else {
					statement.Or(condition.Where("tg.tag_key = ?", tag.Key).Where("tg.tag_value = ?", value))
				}
			}
		case tagmodels.In:
			if i == 0 {
				statement.Where(condition.Where("tg.tag_key = ?", tag.Key).Where("tg.tag_value in ?", tag.Values.List()))
			} else {
				statement.Or(condition.Where("tg.tag_key = ?", tag.Key).Where("tg.tag_value in ?", tag.Values.List()))
			}
		}
	}

	return statement
}

func (d *dao) ListClusterWithExpiry(ctx context.Context,
	query *q.Query) ([]*models.Cluster, error) {
	var clusters []*models.Cluster
	tx := d.db.WithContext(ctx)
	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize
	idThan, ok := query.Keywords[common.IDThan]
	if ok {
		tx = tx.Where("id > ?", idThan)
	}
	result := tx.Where("deleted_ts = ?", 0).Where("status = ?", "").
		Where("expire_seconds > ?", 0).Order("id asc").Limit(limit).Offset(offset).Find(&clusters)
	if result.Error != nil {
		return nil, herrors.NewErrListFailed(herrors.ClusterInDB, result.Error.Error())
	}
	return clusters, nil
}
