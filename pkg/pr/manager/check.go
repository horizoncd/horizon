package manager

import (
	"context"

	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/pr/dao"
	"github.com/horizoncd/horizon/pkg/pr/models"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/pipelinerun/manager/mock_check_manager.go
type CheckManager interface {
	// Create create a check
	Create(ctx context.Context, check *models.Check) (*models.Check, error)
	// Update check run
	UpdateByID(ctx context.Context, checkRunID uint, newCheckRun *models.CheckRun) error
	// GetByResource get checks by resource
	GetByResource(ctx context.Context, resources ...common.Resource) ([]*models.Check, error)
	GetCheckRunByID(ctx context.Context, checkRunID uint) (*models.CheckRun, error)
	ListCheckRuns(ctx context.Context, pipelinerunID uint) ([]*models.CheckRun, error)
	CreateCheckRun(ctx context.Context, checkRun *models.CheckRun) (*models.CheckRun, error)
}

type checkManager struct {
	dao dao.CheckDAO
}

func NewCheckManager(db *gorm.DB) CheckManager {
	return &checkManager{
		dao: dao.NewCheckDAO(db),
	}
}

func (m *checkManager) Create(ctx context.Context, check *models.Check) (*models.Check, error) {
	return m.dao.Create(ctx, check)
}

func (m *checkManager) UpdateByID(ctx context.Context, checkRunID uint, newCheckRun *models.CheckRun) error {
	return m.dao.UpdateByID(ctx, checkRunID, newCheckRun)
}

func (m *checkManager) GetByResource(ctx context.Context, resources ...common.Resource) ([]*models.Check, error) {
	return m.dao.GetByResource(ctx, resources...)
}

func (m *checkManager) GetCheckRunByID(ctx context.Context, checkRunID uint) (*models.CheckRun, error) {
	return m.dao.GetCheckRunByID(ctx, checkRunID)
}

func (m *checkManager) ListCheckRuns(ctx context.Context, pipelinerunID uint) ([]*models.CheckRun, error) {
	return m.dao.ListCheckRuns(ctx, pipelinerunID)
}

func (m *checkManager) CreateCheckRun(ctx context.Context, checkRun *models.CheckRun) (*models.CheckRun, error) {
	return m.dao.CreateCheckRun(ctx, checkRun)
}
