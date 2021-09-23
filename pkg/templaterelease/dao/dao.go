package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/templaterelease/models"
)

type DAO interface {
	Create(ctx context.Context, templateRelease *models.TemplateRelease) (*models.TemplateRelease, error)
	ListByTemplateName(ctx context.Context, templateName string) ([]models.TemplateRelease, error)
	GetByTemplateNameAndRelease(ctx context.Context, templateName, release string) (*models.TemplateRelease, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d dao) Create(ctx context.Context, templateRelease *models.TemplateRelease) (*models.TemplateRelease, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(templateRelease)
	return templateRelease, result.Error
}

func (d dao) ListByTemplateName(ctx context.Context, templateName string) ([]models.TemplateRelease, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var trs []models.TemplateRelease
	result := db.Raw("select * from template_release where template_name = ?", templateName).Scan(&trs)
	if result.Error != nil {
		return nil, result.Error
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
	result := db.Raw("select * from template_release where template_name = ? and name = ?",
		templateName, release).Scan(&tr)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &tr, nil
}
