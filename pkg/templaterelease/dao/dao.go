package dao

import (
	"context"

	he "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
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
func NewDAO() DAO {
	return &dao{}
}

type dao struct{}

func (d dao) Create(ctx context.Context, templateRelease *models.TemplateRelease) (*models.TemplateRelease, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(templateRelease)

	if result.Error != nil {
		return nil, he.NewErrCreateFailed(he.GroupInDB, result.Error.Error())
	}
	return templateRelease, result.Error
}

func (d dao) ListByTemplateName(ctx context.Context, templateName string) ([]*models.TemplateRelease, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var trs []*models.TemplateRelease
	result := db.Raw(common.TemplateReleaseQueryByTemplateName, templateName).Scan(&trs)
	if result.Error != nil {
		return nil, he.NewErrGetFailed(he.TemplateReleaseInDB, result.Error.Error())
	}
	return trs, nil
}

func (d dao) GetByTemplateNameAndRelease(ctx context.Context,
	templateName, release string) (*models.TemplateRelease, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var tr models.TemplateRelease
	result := db.Raw(common.TemplateReleaseQueryByTemplateNameAndName,
		templateName, release).First(&tr)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &tr, he.NewErrNotFound(he.TemplateReleaseInDB, result.Error.Error())
		}
		return &tr, he.NewErrGetFailed(he.TemplateReleaseInDB, result.Error.Error())
	}

	return &tr, result.Error
}
