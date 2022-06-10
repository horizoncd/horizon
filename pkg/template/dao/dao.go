package dao

import (
	"context"

	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/template/models"
	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, template *models.Template) (*models.Template, error)
	List(ctx context.Context) ([]models.Template, error)
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

func (d dao) List(ctx context.Context) ([]models.Template, error) {

	var templates []models.Template
	result := d.db.WithContext(ctx).Raw(common.TemplateQuery).Scan(&templates)
	if result.Error != nil {
		return nil, result.Error
	}
	return templates, nil
}
