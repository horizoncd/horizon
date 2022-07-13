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
	"g.hz.netease.com/horizon/pkg/templaterelease/models"
	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, templateRelease *models.TemplateRelease) (*models.TemplateRelease, error)
	ListByTemplateName(ctx context.Context, templateName string) ([]*models.TemplateRelease, error)
	ListByTemplateID(ctx context.Context, id uint) ([]*models.TemplateRelease, error)
	GetByTemplateNameAndRelease(ctx context.Context, templateName, release string) (*models.TemplateRelease, error)
	GetByID(ctx context.Context, releaseID uint) (*models.TemplateRelease, error)
	GetRefOfApplication(ctx context.Context, id uint) ([]*amodels.Application, uint, error)
	GetRefOfCluster(ctx context.Context, id uint) ([]*cmodel.Cluster, uint, error)
	UpdateByID(ctx context.Context, releaseID uint, release *models.TemplateRelease) error
	DeleteByID(ctx context.Context, id uint) error
}

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

type dao struct{ db *gorm.DB }

func (d dao) Create(ctx context.Context, templateRelease *models.TemplateRelease) (*models.TemplateRelease, error) {
	result := d.db.WithContext(ctx).Create(templateRelease)

	if result.Error != nil {
		return nil, herrors.NewErrCreateFailed(herrors.GroupInDB, result.Error.Error())
	}
	return templateRelease, nil
}

func (d dao) ListByTemplateName(ctx context.Context, templateName string) ([]*models.TemplateRelease, error) {
	var trs []*models.TemplateRelease
	result := d.db.WithContext(ctx).Raw(common.TemplateReleaseQueryByTemplateName, templateName).Scan(&trs)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.TemplateReleaseInDB, result.Error.Error())
	}
	return trs, nil
}

func (d dao) ListByTemplateID(ctx context.Context, templateID uint) ([]*models.TemplateRelease, error) {
	var trs []*models.TemplateRelease
	result := d.db.Raw(common.TemplateReleaseListByTemplateID, templateID).Scan(&trs)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.TemplateReleaseInDB, result.Error.Error())
	}
	return trs, nil
}

func (d dao) GetByTemplateNameAndRelease(ctx context.Context,
	templateName, release string) (*models.TemplateRelease, error) {
	var tr models.TemplateRelease
	result := d.db.WithContext(ctx).Raw(common.TemplateReleaseQueryByTemplateNameAndName,
		templateName, release).First(&tr)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.TemplateReleaseInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.TemplateReleaseInDB, result.Error.Error())
	}

	return &tr, nil
}

func (d dao) GetByID(ctx context.Context, releaseID uint) (*models.TemplateRelease, error) {
	var tr models.TemplateRelease
	result := d.db.Raw(common.TemplateReleaseQueryByID,
		releaseID).First(&tr)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.TemplateReleaseInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.TemplateReleaseInDB, result.Error.Error())
	}

	return &tr, nil
}
func (d dao) GetRefOfApplication(ctx context.Context, id uint) ([]*amodels.Application, uint, error) {
	onlyRefCount, ok := ctx.Value(hctx.TemplateOnlyRefCount).(bool)
	var (
		applications []*amodels.Application
		total        uint
	)
	res := d.db.Raw(common.TemplateReleaseRefCountOfApplication, id).Scan(&total)
	if res.Error != nil {
		return nil, 0, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
			fmt.Sprintf("failed to get ref count of application: %s", res.Error.Error()))
	}

	if !ok || !onlyRefCount {
		res = d.db.Raw(common.TemplateReleaseRefOfApplication, id).Scan(&applications)
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
	res := d.db.Raw(common.TemplateReleaseRefCountOfCluster, id).Scan(&total)
	if res.Error != nil {
		return nil, 0, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
			fmt.Sprintf("failed to get ref count of clsuter: %s", res.Error.Error()))
	}

	if !ok || !onlyRefCount {
		res = d.db.Raw(common.TemplateReleaseRefOfCluster, id).Scan(&clusters)
		if res.Error != nil {
			return nil, 0, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, res.Error.Error()),
				fmt.Sprintf("failed to get ref of clsuter: %s", res.Error.Error()))
		}
	}
	return clusters, total, nil
}

func (d dao) UpdateByID(ctx context.Context, releaseID uint, release *models.TemplateRelease) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		var oldRelease models.TemplateRelease
		res := tx.Raw(common.TemplateReleaseQueryByID, releaseID).Scan(&oldRelease)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return perror.Wrap(herrors.NewErrNotFound(herrors.TemplateInDB,
				"not found"),
				fmt.Sprintf("not found template with releaseID = %d", releaseID))
		}

		oldRelease.UpdatedBy = release.UpdatedBy
		if release.Description != "" {
			oldRelease.Description = release.Description
		}
		if release.Recommended != nil {
			oldRelease.Recommended = release.Recommended
		}
		if release.OnlyAdmin != nil {
			oldRelease.OnlyAdmin = release.OnlyAdmin
		}
		return tx.Model(&oldRelease).Updates(oldRelease).Error
	})
}

func (d dao) DeleteByID(ctx context.Context, id uint) error {
	if res := d.db.Exec(common.TemplateReleaseDelete, id); res.Error != nil {
		return perror.Wrap(herrors.NewErrDeleteFailed(herrors.TemplateInDB, res.Error.Error()),
			fmt.Sprintf("failed to delete template, id = %d", id))
	}
	return nil
}
