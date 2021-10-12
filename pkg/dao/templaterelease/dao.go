package templaterelease

import (
	"context"

	"g.hz.netease.com/horizon/pkg/dao/common"
	"g.hz.netease.com/horizon/pkg/lib/orm"
)

type DAO interface {
	Create(ctx context.Context, templateRelease *TemplateRelease) (*TemplateRelease, error)
	ListByTemplateName(ctx context.Context, templateName string) ([]TemplateRelease, error)
	GetByTemplateNameAndRelease(ctx context.Context, templateName, release string) (*TemplateRelease, error)
}

// newDAO returns an instance of the default DAO
func newDAO() DAO {
	return &dao{}
}

type dao struct{}

func (d dao) Create(ctx context.Context, templateRelease *TemplateRelease) (*TemplateRelease, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(templateRelease)
	return templateRelease, result.Error
}

func (d dao) ListByTemplateName(ctx context.Context, templateName string) ([]TemplateRelease, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var trs []TemplateRelease
	result := db.Raw(common.TemplateReleaseQueryByTemplateName, templateName).Scan(&trs)
	if result.Error != nil {
		return nil, result.Error
	}
	return trs, nil
}

func (d dao) GetByTemplateNameAndRelease(ctx context.Context,
	templateName, release string) (*TemplateRelease, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var tr TemplateRelease
	result := db.Raw(common.TemplateReleaseQueryByTemplateNameAndName,
		templateName, release).Scan(&tr)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &tr, nil
}
