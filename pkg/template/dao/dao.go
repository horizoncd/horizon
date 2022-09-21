package dao

import (
	"context"
	"fmt"

	herrors "g.hz.netease.com/horizon/core/errors"
	amodels "g.hz.netease.com/horizon/pkg/application/models"
	cmodel "g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/common"
	hctx "g.hz.netease.com/horizon/pkg/context"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/template/models"
	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, template *models.Template) (*models.Template, error)
	List(ctx context.Context) ([]*models.Template, error)
	ListByGroupID(ctx context.Context, groupID uint) ([]*models.Template, error)
	DeleteByID(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.Template, error)
	GetByName(ctx context.Context, name string) (*models.Template, error)
	GetRefOfApplication(ctx context.Context, id uint) ([]*amodels.Application, uint, error)
	GetRefOfCluster(ctx context.Context, id uint) ([]*cmodel.Cluster, uint, error)
	UpdateByID(ctx context.Context, id uint, template *models.Template) error
	ListByGroupIDs(ctx context.Context, ids []uint) ([]*models.Template, error)
	ListByIDs(ctx context.Context, ids []uint) ([]*models.Template, error)
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

func (d dao) List(ctx context.Context) ([]*models.Template, error) {
	var templates []*models.Template
	result := d.db.Raw(common.TemplateList).Scan(&templates)
	if result.Error != nil {
		return nil, result.Error
	}
	return templates, nil
}

func (d dao) ListByGroupID(ctx context.Context, groupID uint) ([]*models.Template, error) {
	var templates []*models.Template
	result := d.db.Raw(common.TemplateListByGroup, groupID).Scan(&templates)
	if result.Error != nil {
		return nil, result.Error
	}
	return templates, nil
}

func (d dao) DeleteByID(ctx context.Context, id uint) error {
	if res := d.db.Exec(common.TemplateDelete, id); res.Error != nil {
		return perror.Wrap(herrors.NewErrDeleteFailed(herrors.TemplateInDB, res.Error.Error()),
			fmt.Sprintf("failed to delete template, id = %d", id))
	}
	return nil
}

func (d dao) GetByID(ctx context.Context, id uint) (*models.Template, error) {
	var template models.Template
	res := d.db.Raw(common.TemplateQueryByID, id).First(&template)
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
	res := d.db.Raw(common.TemplateQueryByName, name).First(&template)
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
	res := d.db.Raw(common.TemplateRefCountOfApplication, id).Scan(&total)
	if res.Error != nil {
		return nil, 0, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
			fmt.Sprintf("failed to get ref count of application: %s", res.Error.Error()))
	}

	if !ok || !onlyRefCount {
		res = d.db.Raw(common.TemplateRefOfApplication, id).Scan(&applications)
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
	res := d.db.Raw(common.TemplateRefCountOfCluster, id).Scan(&total)
	if res.Error != nil {
		return nil, 0, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
			fmt.Sprintf("failed to get ref count of cluster: %s", res.Error.Error()))
	}

	if !ok || !onlyRefCount {
		res = d.db.Raw(common.TemplateRefOfCluster, id).Scan(&clusters)
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
		res := tx.Raw(common.TemplateQueryByID, templateID).Scan(&oldTemplate)
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
