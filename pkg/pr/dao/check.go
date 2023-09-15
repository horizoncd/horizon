package dao

import (
	"context"
	"fmt"

	"github.com/horizoncd/horizon/lib/q"
	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/pr/models"
)

type CheckDAO interface {
	// Create create a check
	Create(ctx context.Context, check *models.Check) (*models.Check, error)
	// Update check run
	UpdateByID(ctx context.Context, checkRunID uint, newCheckRun *models.CheckRun) error
	// GetByResource get checks by resource
	GetByResource(ctx context.Context, resources ...common.Resource) ([]*models.Check, error)
	GetCheckRunByID(ctx context.Context, checkRunID uint) (*models.CheckRun, error)
	ListCheckRuns(ctx context.Context, query *q.Query) ([]*models.CheckRun, error)
	CreateCheckRun(ctx context.Context, run *models.CheckRun) (*models.CheckRun, error)
}

type checkDAO struct{ db *gorm.DB }

func NewCheckDAO(db *gorm.DB) CheckDAO {
	return &checkDAO{db: db}
}

func (d *checkDAO) Create(ctx context.Context, check *models.Check) (*models.Check, error) {
	result := d.db.WithContext(ctx).Debug().Create(check)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.CheckInDB, result.Error.Error())
	}

	return check, result.Error
}

func (d *checkDAO) UpdateByID(ctx context.Context, checkRunID uint, newCheckRun *models.CheckRun) error {
	result := d.db.WithContext(ctx).Model(&models.CheckRun{}).Debug().Where("id = ?", checkRunID).Updates(newCheckRun)

	if result.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.CheckInDB, result.Error.Error())
	}

	return result.Error
}

func (d *checkDAO) GetByResource(ctx context.Context, resources ...common.Resource) ([]*models.Check, error) {
	var checks []*models.Check
	sql := d.db.WithContext(ctx).Debug()
	if len(resources) == 0 {
		return []*models.Check{}, nil
	}

	sql = sql.Where(d.db.Where("resource_type = ?", resources[0].Type).Where("resource_id = ?", resources[0].ResourceID))
	for _, resource := range resources[1:] {
		sql = sql.Or(d.db.Where("resource_type = ?", resource.Type).Where("resource_id = ?", resource.ResourceID))
	}

	result := sql.Find(&checks)

	if result.RowsAffected == 0 {
		return []*models.Check{}, nil
	}
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.CheckInDB, result.Error.Error())
	}
	return checks, nil
}

func (d *checkDAO) ListCheckRuns(ctx context.Context, query *q.Query) ([]*models.CheckRun, error) {
	var checkRuns []*models.CheckRun

	statement := d.db.WithContext(ctx)
	if query != nil {
		for k, v := range query.Keywords {
			switch k {
			case common.CheckrunQueryFilter:
				statement = statement.Where("name like ?", fmt.Sprintf("%%%v%%", v))
			case common.CheckrunQueryByStatus:
				status := models.String2CheckRunStatus(v.(string))
				statement = statement.Where("status = ?", status)
			case common.CheckrunQueryByPipelinerunID:
				statement = statement.Where("pipeline_run_id = ?", v)
			case common.CheckrunQueryByCheckID:
				statement = statement.Where("check_id = ?", v)
			}
		}
	}
	result := statement.Debug().Find(&checkRuns)

	if result.RowsAffected == 0 {
		return []*models.CheckRun{}, nil
	}
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.CheckRunInDB, result.Error.Error())
	}
	return checkRuns, nil
}

func (d *checkDAO) CreateCheckRun(ctx context.Context, run *models.CheckRun) (*models.CheckRun, error) {
	result := d.db.WithContext(ctx).Create(run)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.CheckRunInDB, result.Error.Error())
	}

	return run, result.Error
}

func (d *checkDAO) GetCheckRunByID(ctx context.Context, checkRunID uint) (*models.CheckRun, error) {
	var checkRun models.CheckRun
	result := d.db.WithContext(ctx).Where("id = ?", checkRunID).Find(&checkRun)

	if result.RowsAffected == 0 {
		return nil, herrors.NewErrNotFound(herrors.CheckRunInDB, "check run not found")
	}
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.CheckRunInDB, result.Error.Error())
	}
	return &checkRun, nil
}
