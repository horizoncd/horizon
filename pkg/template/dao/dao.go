package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/template/models"
)

type DAO interface {
	Create(ctx context.Context, template *models.Template) (*models.Template, error)
	List(ctx context.Context) ([]models.Template, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d dao) Create(ctx context.Context, template *models.Template) (*models.Template, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(template)
	return template, result.Error
}

func (d dao) List(ctx context.Context) ([]models.Template, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var templates []models.Template
	result := db.Raw("select * from template").Scan(&templates)
	if result.Error != nil {
		return nil, result.Error
	}
	return templates, nil
}
