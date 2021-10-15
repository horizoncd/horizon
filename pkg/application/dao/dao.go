package dao

import (
	"context"
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/util/errors"

	"gorm.io/gorm"
)

type DAO interface {
	GetByName(ctx context.Context, name string) (*models.Application, error)
	GetByNamesUnderGroup(ctx context.Context, groupID uint, names []string) ([]*models.Application, error)
	// GetByNameFuzzily get applications that fuzzily matching the given name
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error)
	// CountByGroupID get the count of the records matching the given groupID
	CountByGroupID(ctx context.Context, groupID uint) (int64, error)
	Create(ctx context.Context, application *models.Application) (*models.Application, error)
	UpdateByName(ctx context.Context, name string, application *models.Application) (*models.Application, error)
	DeleteByName(ctx context.Context, name string) error
}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) CountByGroupID(ctx context.Context, groupID uint) (int64, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	var count int64
	result := db.Raw(common.ApplicationCountByGroupID, groupID).Scan(&count)

	return count, result.Error
}

func (d *dao) GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var applications []*models.Application
	result := db.Raw(common.ApplicationQueryByFuzzily, fmt.Sprintf("%%%s%%", name)).Scan(&applications)

	return applications, result.Error
}

func (d *dao) GetByName(ctx context.Context, name string) (*models.Application, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var application models.Application
	result := db.Raw(common.ApplicationQueryByName, name).Scan(&application)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &application, nil
}

func (d *dao) GetByNamesUnderGroup(ctx context.Context, groupID uint, names []string) ([]*models.Application, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var applications []*models.Application
	result := db.Raw(common.ApplicationQueryByNamesUnderGroup, groupID, names).Scan(&applications)

	return applications, result.Error
}

func (d *dao) Create(ctx context.Context, application *models.Application) (*models.Application, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(application)
	return application, result.Error
}

func (d *dao) UpdateByName(ctx context.Context, name string,
	application *models.Application) (*models.Application, error) {
	const op = "application dao: update by name"

	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var applicationInDB models.Application
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 1. get application in db first
		result := tx.Raw(common.ApplicationQueryByName, name).Scan(&applicationInDB)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.E(op, http.StatusNotFound)
		}
		// 2. update value
		applicationInDB.Description = application.Description
		applicationInDB.Priority = application.Priority
		applicationInDB.GitURL = application.GitURL
		applicationInDB.GitSubfolder = application.GitSubfolder
		applicationInDB.GitBranch = application.GitBranch
		applicationInDB.Template = application.Template
		applicationInDB.TemplateRelease = application.TemplateRelease
		applicationInDB.UpdatedBy = application.UpdatedBy
		// 3. save application after updated
		tx.Save(&applicationInDB)

		return nil
	}); err != nil {
		return nil, err
	}
	return &applicationInDB, nil
}

func (d *dao) DeleteByName(ctx context.Context, name string) error {
	const op = "application dao: delete by name"

	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Exec(common.ApplicationDeleteByName, name)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.E(op, http.StatusNotFound)
	}
	return nil
}
