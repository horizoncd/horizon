package member

import (
	"context"

	"g.hz.netease.com/horizon/pkg/member/dao"
	"g.hz.netease.com/horizon/pkg/member/models"
)

var (
	// Mgr is the global member manager
	Mgr = New()
)

type Manager interface {
	Create(ctx context.Context, member *models.Member) (*models.Member, error)

	GetByUserName(ctx context.Context, resourceType models.ResourceType,
		resourceID uint, userName string) (*models.Member, error)
	// TODO(tom) GetByGroupID

	UpdateByID(ctx context.Context, id uint, member *models.Member) (*models.Member, error)

	DeleteMember(ctx context.Context, memberID uint) error

	ListDirectMember(ctx context.Context, resourceType models.ResourceType,
		resourceID uint) ([]models.Member, error)
}

type manager struct {
	dao dao.DAO
}

func New() Manager {
	return &manager{dao: dao.New()}
}

func (m *manager) Create(ctx context.Context, member *models.Member) (*models.Member, error) {
	return m.dao.Create(ctx, member)
}

func (m *manager) GetByUserName(ctx context.Context, resourceType models.ResourceType, resourceID uint,
	userName string) (*models.Member, error) {
	return m.dao.Get(ctx, resourceType, resourceID, models.MemberUser, userName)
}

func (m *manager) UpdateByID(ctx context.Context, memberID uint, member *models.Member) (*models.Member, error) {
	return m.dao.UpdateByID(ctx, memberID, member)
}

func (m *manager) DeleteMember(ctx context.Context, memberID uint) error {
	return m.dao.Delete(ctx, memberID)
}

func (m *manager) ListDirectMember(ctx context.Context, resourceType models.ResourceType,
	resourceID uint) ([]models.Member, error) {
	return m.dao.ListDirectMember(ctx, resourceType, resourceID)
}
