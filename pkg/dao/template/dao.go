package template

import (
	"context"

	"g.hz.netease.com/horizon/pkg/dao/common"
	"g.hz.netease.com/horizon/pkg/lib/orm"
)

type DAO interface {
	Create(ctx context.Context, template *Template) (*Template, error)
	List(ctx context.Context) ([]Template, error)
}

// newDAO returns an instance of the default DAO
func newDAO() DAO {
	return &dao{}
}

type dao struct{}

func (d dao) Create(ctx context.Context, template *Template) (*Template, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(template)
	return template, result.Error
}

func (d dao) List(ctx context.Context) ([]Template, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var templates []Template
	result := db.Raw(common.TemplateQuery).Scan(&templates)
	if result.Error != nil {
		return nil, result.Error
	}
	return templates, nil
}
