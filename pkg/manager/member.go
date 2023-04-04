// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"context"

	"github.com/horizoncd/horizon/pkg/dao"
	"github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/member/manager/mock_manager.go -package=mock_manager
type MemberManager interface {
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

	// DeleteMemberByResourceTypeID Delete a member by memberID
	HardDeleteMemberByResourceTypeID(ctx context.Context, resourceType string, resourceID uint) error

	// DeleteMemberByMemberNameID Delete member by memberNameID
	DeleteMemberByMemberNameID(ctx context.Context, memberNameID uint) error

	// ListDirectMember List the direct member of the resource
	ListDirectMember(ctx context.Context, resourceType models.ResourceType,
		resourceID uint) ([]models.Member, error)

	// ListDirectMemberOnCondition Lists the direct member of the resource on condition
	ListDirectMemberOnCondition(ctx context.Context, resourceType models.ResourceType,
		resourceID uint) ([]models.Member, error)

	// ListResourceOfMemberInfo list the resource id of the specified resourceType and memberInfo
	ListResourceOfMemberInfo(ctx context.Context,
		resourceType models.ResourceType, memberInfo uint) ([]uint, error)

	ListResourceOfMemberInfoByRole(ctx context.Context,
		resourceType models.ResourceType, memberInfo uint, role string) ([]uint, error)

	ListMembersByUserID(ctx context.Context, userID uint) ([]models.Member, error)
}

type memberManager struct {
	dao dao.MemberDAO
}

func NewMemberManager(db *gorm.DB) MemberManager {
	return &memberManager{dao: dao.NewMemberDAO(db)}
}

func (m *memberManager) Create(ctx context.Context, member *models.Member) (*models.Member, error) {
	return m.dao.Create(ctx, member)
}

func (m *memberManager) Get(ctx context.Context, resourceType models.ResourceType,
	resourceID uint, memberType models.MemberType, memberInfo uint) (*models.Member, error) {
	return m.dao.Get(ctx, resourceType, resourceID, memberType, memberInfo)
}

func (m *memberManager) GetByID(ctx context.Context, memberID uint) (*models.Member, error) {
	return m.dao.GetByID(ctx, memberID)
}

func (m *memberManager) UpdateByID(ctx context.Context, memberID uint, role string) (*models.Member, error) {
	return m.dao.UpdateByID(ctx, memberID, role)
}

func (m *memberManager) DeleteMember(ctx context.Context, memberID uint) error {
	return m.dao.Delete(ctx, memberID)
}

func (m *memberManager) DeleteMemberByMemberNameID(ctx context.Context, memberNameID uint) error {
	return m.dao.DeleteByMemberNameID(ctx, memberNameID)
}

func (m *memberManager) HardDeleteMemberByResourceTypeID(ctx context.Context,
	resourceType string, resourceID uint) error {
	return m.dao.HardDelete(ctx, resourceType, resourceID)
}

func (m *memberManager) ListDirectMember(ctx context.Context, resourceType models.ResourceType,
	resourceID uint) ([]models.Member, error) {
	return m.dao.ListDirectMember(ctx, resourceType, resourceID)
}

func (m *memberManager) ListDirectMemberOnCondition(ctx context.Context, resourceType models.ResourceType,
	resourceID uint) ([]models.Member, error) {
	return m.dao.ListDirectMemberOnCondition(ctx, resourceType, resourceID)
}

func (m *memberManager) ListResourceOfMemberInfo(ctx context.Context,
	resourceType models.ResourceType, memberInfo uint) ([]uint, error) {
	return m.dao.ListResourceOfMemberInfo(ctx, resourceType, memberInfo)
}
func (m *memberManager) ListResourceOfMemberInfoByRole(ctx context.Context,
	resourceType models.ResourceType, memberInfo uint, role string) ([]uint, error) {
	return m.dao.ListResourceOfMemberInfoByRole(ctx, resourceType, memberInfo, role)
}

func (m *memberManager) ListMembersByUserID(ctx context.Context, userID uint) ([]models.Member, error) {
	return m.dao.ListMembersByUserID(ctx, userID)
}
