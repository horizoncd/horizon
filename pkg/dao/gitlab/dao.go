package gitlab

import (
	"context"

	"g.hz.netease.com/horizon/pkg/dao/common"
	"g.hz.netease.com/horizon/pkg/lib/orm"
)

type DAO interface {
	Create(ctx context.Context, template *Gitlab) (*Gitlab, error)
	List(ctx context.Context) ([]Gitlab, error)
	GetByName(ctx context.Context, name string) (*Gitlab, error)
}

// newDAO returns an instance of the default DAO
func newDAO() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(ctx context.Context, gitlab *Gitlab) (*Gitlab, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(gitlab)
	return gitlab, result.Error
}

func (d *dao) List(ctx context.Context) ([]Gitlab, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var gitlabs []Gitlab
	result := db.Raw(common.GitlabQuery).Scan(&gitlabs)
	if result.Error != nil {
		return nil, result.Error
	}
	return gitlabs, nil
}

func (d *dao) GetByName(ctx context.Context, name string) (*Gitlab, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var gitlab Gitlab
	result := db.Raw(common.GitlabQueryByName, name).Scan(&gitlab)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &gitlab, nil
}
