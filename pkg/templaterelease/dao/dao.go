package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/templaterelease/models"
	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, templateRelease *models.TemplateRelease) (*models.TemplateRelease, error)
	ListByTemplateName(ctx context.Context, templateName string) ([]*models.TemplateRelease, error)
	GetByTemplateNameAndRelease(ctx context.Context, templateName, release string) (*models.TemplateRelease, error)
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
	return templateRelease, result.Error
}

func (d dao) ListByTemplateName(ctx context.Context, templateName string) ([]*models.TemplateRelease, error) {
	var trs []*models.TemplateRelease
	result := d.db.WithContext(ctx).Raw(common.TemplateReleaseQueryByTemplateName, templateName).Scan(&trs)
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
			return &tr, herrors.NewErrNotFound(herrors.TemplateReleaseInDB, result.Error.Error())
		}
		return &tr, herrors.NewErrGetFailed(herrors.TemplateReleaseInDB, result.Error.Error())
	}

	return &tr, result.Error
}
