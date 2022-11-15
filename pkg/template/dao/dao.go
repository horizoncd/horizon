package dao

import (
	"context"
	"fmt"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	amodels "g.hz.netease.com/horizon/pkg/application/models"
	cmodel "g.hz.netease.com/horizon/pkg/cluster/models"
	dbsql "g.hz.netease.com/horizon/pkg/common"
	hctx "g.hz.netease.com/horizon/pkg/context"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/template/models"
	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, template *models.Template) (*models.Template, error)
	ListTemplate(ctx context.Context) ([]*models.Template, error)
	ListByGroupID(ctx context.Context, groupID uint) ([]*models.Template, error)
	DeleteByID(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.Template, error)
	GetByName(ctx context.Context, name string) (*models.Template, error)
	GetRefOfApplication(ctx context.Context, id uint) ([]*amodels.Application, uint, error)
	GetRefOfCluster(ctx context.Context, id uint) ([]*cmodel.Cluster, uint, error)
	UpdateByID(ctx context.Context, id uint, template *models.Template) error
	ListByGroupIDs(ctx context.Context, ids []uint) ([]*models.Template, error)
	ListByIDs(ctx context.Context, ids []uint) ([]*models.Template, error)
	ListV2(ctx context.Context, query *q.Query, gorupIDs ...uint) (int, []*models.Template, error)
}

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

type dao struct{ db *gorm.DB }

func (d dao) Create(ctx context.Context, template *models.Template) (*models.Template, error) {
	result := d.db.WithContext(ctx).Create(template)
	return template, result.Error
}

func (d dao) ListTemplate(ctx context.Context) ([]*models.Template, error) {
	var templates []*models.Template
	result := d.db.Raw(dbsql.TemplateList).Scan(&templates)
	if result.Error != nil {
		return nil, result.Error
	}
	return templates, nil
}

func (d dao) ListByGroupID(ctx context.Context, groupID uint) ([]*models.Template, error) {
	var templates []*models.Template
	result := d.db.Raw(dbsql.TemplateListByGroup, groupID).Scan(&templates)
	if result.Error != nil {
		return nil, result.Error
	}
	return templates, nil
}

func (d dao) DeleteByID(ctx context.Context, id uint) error {
	if res := d.db.Exec(dbsql.TemplateDelete, id); res.Error != nil {
		return perror.Wrap(herrors.NewErrDeleteFailed(herrors.TemplateInDB, res.Error.Error()),
			fmt.Sprintf("failed to delete template, id = %d", id))
	}
	return nil
}

func (d dao) GetByID(ctx context.Context, id uint) (*models.Template, error) {
	var template models.Template
	res := d.db.Raw(dbsql.TemplateQueryByID, id).First(&template)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, perror.Wrap(herrors.NewErrNotFound(herrors.TemplateInDB, res.Error.Error()),
				fmt.Sprintf("failed to find template: id = %d", id))
		}
		return nil, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
			fmt.Sprintf("failed to get template: id = %d", id))
	}
	return &template, nil
}

func (d dao) GetByName(ctx context.Context, name string) (*models.Template, error) {
	var template models.Template
	res := d.db.Raw(dbsql.TemplateQueryByName, name).First(&template)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, perror.Wrap(herrors.NewErrNotFound(herrors.TemplateInDB, res.Error.Error()),
				fmt.Sprintf("failed to find template: name = %s", name))
		}
		return nil, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
			fmt.Sprintf("failed to get template: name = %s", name))
	}
	return &template, nil
}

func (d dao) GetRefOfApplication(ctx context.Context, id uint) ([]*amodels.Application, uint, error) {
	onlyRefCount, ok := ctx.Value(hctx.TemplateOnlyRefCount).(bool)
	var (
		applications []*amodels.Application
		total        uint
	)
	res := d.db.Raw(dbsql.TemplateRefCountOfApplication, id).Scan(&total)
	if res.Error != nil {
		return nil, 0, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
			fmt.Sprintf("failed to get ref count of application: %s", res.Error.Error()))
	}

	if !ok || !onlyRefCount {
		res = d.db.Raw(dbsql.TemplateRefOfApplication, id).Scan(&applications)
		if res.Error != nil {
			return nil, 0, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
				fmt.Sprintf("failed to get ref of application: %s", res.Error.Error()))
		}
	}
	return applications, total, nil
}

func (d dao) GetRefOfCluster(ctx context.Context, id uint) ([]*cmodel.Cluster, uint, error) {
	onlyRefCount, ok := ctx.Value(hctx.TemplateOnlyRefCount).(bool)
	var (
		clusters []*cmodel.Cluster
		total    uint
	)
	res := d.db.Raw(dbsql.TemplateRefCountOfCluster, id).Scan(&total)
	if res.Error != nil {
		return nil, 0, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
			fmt.Sprintf("failed to get ref count of cluster: %s", res.Error.Error()))
	}

	if !ok || !onlyRefCount {
		res = d.db.Raw(dbsql.TemplateRefOfCluster, id).Scan(&clusters)
		if res.Error != nil {
			return nil, 0, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
				fmt.Sprintf("failed to get ref of cluster: %s", res.Error.Error()))
		}
	}
	return clusters, total, nil
}

func (d dao) UpdateByID(ctx context.Context, templateID uint, template *models.Template) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		var oldTemplate models.Template
		res := tx.Raw(dbsql.TemplateQueryByID, templateID).Scan(&oldTemplate)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return perror.Wrap(herrors.NewErrNotFound(herrors.TemplateInDB, res.Error.Error()),
				fmt.Sprintf("not found template with templateID = %d", templateID))
		}

		oldTemplate.UpdatedBy = template.UpdatedBy
		if template.Repository != "" {
			oldTemplate.Repository = template.Repository
		}
		if template.Description != "" {
			oldTemplate.Description = template.Description
		}
		if template.OnlyOwner != nil {
			oldTemplate.OnlyOwner = template.OnlyOwner
		}
		return tx.Model(&oldTemplate).Updates(oldTemplate).Error
	})
}

func (d dao) ListByGroupIDs(ctx context.Context, ids []uint) ([]*models.Template, error) {
	templates := make([]*models.Template, 0)
	res := d.db.Where("group_id in ?", ids).Find(&templates)
	if res.Error != nil {
		return nil, perror.Wrapf(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
			"failed to get template:\n"+
				"template ids = %v\n err = %v", ids, res.Error)
	}
	return templates, nil
}

func (d dao) ListByIDs(ctx context.Context, ids []uint) ([]*models.Template, error) {
	templates := make([]*models.Template, 0)
	res := d.db.Where("id in ?", ids).Find(&templates)
	if res.Error != nil {
		return nil, perror.Wrapf(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
			"failed to get template:\n"+
				"template ids = %v\n err = %v", ids, res.Error)
	}
	return templates, nil
}

func (d dao) ListV2(ctx context.Context, query *q.Query, groupIDs ...uint) (int, []*models.Template, error) {
	var (
		templates []*models.Template
		total     int64
	)

	statement := d.db.WithContext(ctx).
		Table("tb_template as t").
		Select("t.*")

	genSQL := func(statement *gorm.DB, query *q.Query) *gorm.DB {
		for k, v := range query.Keywords {
			switch k {
			case common.ParamGroupID:
				statement = statement.Where("t.group_id = ?", v)
			case common.TemplateQueryName:
				statement = statement.Where("t.name like ?", fmt.Sprintf("%%%v%%", v))
			case common.TemplateQueryByGroup:
				if _, ok := v.(uint); ok {
					statement = statement.Where("t.group_id = ?", v)
				} else if _, ok = v.([]uint); ok {
					statement = statement.Where("t.group_id in ?", v)
				}
			case common.TemplateQueryByUser:
				statement = statement.
					Joins("join tb_member as m on m.resource_id = t.id").
					Where("m.resource_type = ?", common.ResourceTemplate).
					Where("m.member_type = '0'").
					Where("m.deleted_ts = 0").
					Where("m.membername_id = ?", v)
			}
		}
		statement = statement.Where("t.deleted_ts = 0")
		return statement
	}

	if query != nil {
		statement = genSQL(statement, query)

		if len(groupIDs) > 0 &&
			query.Keywords != nil &&
			query.Keywords[common.TemplateQueryByUser] != nil {
			statementGroup := d.db.WithContext(ctx).
				Table("tb_template as t").
				Select("t.*")

			delete(query.Keywords, common.TemplateQueryByUser)
			statementGroup = genSQL(statementGroup, query)

			statementGroup = statementGroup.Where("group_id in ?", groupIDs)
			statement = d.db.Raw("? union ?", statement, statementGroup)
		}
	}

	res := d.db.Raw("select count(distinct id) from (?) as templates", statement).Debug().Scan(&total)

	if res.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error())
	}

	statement = d.db.Raw("select distinct * from (?) as apps order by updated_at desc limit ? offset ?",
		statement, query.Limit(), query.Offset())
	res = statement.Scan(&templates)
	if res.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error())
	}

	return int(total), templates, nil
}
