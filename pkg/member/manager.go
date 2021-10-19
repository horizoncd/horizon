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
	// Create a new member
	Create(ctx context.Context, member *models.Member) (*models.Member, error)

	// Get  return the direct member
	Get(ctx context.Context, resourceType models.ResourceType,
		resourceID uint, memberType models.MemberType, memberInfo uint) (*models.Member, error)

	// GetByID get the member by ID
	GetByID(ctx context.Context, memberID uint) (*models.Member, error)

	// UpdateByID  update a member by memberID
	UpdateByID(ctx context.Context, id uint, role string) (*models.Member, error)

	// DeleteMember Delete a member by memberID
	DeleteMember(ctx context.Context, memberID uint) error

	// ListDirectMember List the direct member of the resource
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

func (m *manager) Get(ctx context.Context, resourceType models.ResourceType,
	resourceID uint, memberType models.MemberType, memberInfo uint) (*models.Member, error) {
	return m.dao.Get(ctx, resourceType, resourceID, memberType, memberInfo)
}

func (m *manager) GetByID(ctx context.Context, memberID uint) (*models.Member, error) {
	return m.dao.GetByID(ctx, memberID)
}

func (m *manager) UpdateByID(ctx context.Context, memberID uint, role string) (*models.Member, error) {
	return m.dao.UpdateByID(ctx, memberID, role)
}

func (m *manager) DeleteMember(ctx context.Context, memberID uint) error {
	return m.dao.Delete(ctx, memberID)
}

func (m *manager) ListDirectMember(ctx context.Context, resourceType models.ResourceType,
	resourceID uint) ([]models.Member, error) {
	return m.dao.ListDirectMember(ctx, resourceType, resourceID)
}
