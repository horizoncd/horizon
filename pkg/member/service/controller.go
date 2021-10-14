package service

import (
	"context"
	"errors"
	"strconv"

	"g.hz.netease.com/horizon/core/middleware/user"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	groupManager "g.hz.netease.com/horizon/pkg/group/manager"
	"g.hz.netease.com/horizon/pkg/member"
	"g.hz.netease.com/horizon/pkg/member/models"
)

var (
	ErrMemberExist    = errors.New("MemberExist")     // "Member Exist"
	ErrNotPermitted   = errors.New("NotPermitted")    // "Not Permitted"
	ErrMemberNotExist = errors.New("MemberNotExist")  // "Member not exists"
	ErrGrantHighRole  = errors.New("GrantHigherRole") // "Grant higher role"
	ErrRemoveHighRole = errors.New("RemoveHighRole")  // "Remove higher role"
)

var (
	Ctl = NewService()
)

type Service interface {
	// GetUserMember get the service of the current user
	GetUserMember(ctx context.Context, resourceType string, resourceID uint) (*models.Member, error)
	// CreateMember post a new service
	CreateMember(ctx context.Context, postMember PostMember) (*models.Member, error)
	// UpdateMember update exist service entry
	// user can only update a role not higher than self
	UpdateMember(ctx context.Context, resourceType string, resourceID uint,
		memberInfo uint, memberType models.MemberType, role string) (*models.Member, error)
	// RemoveMember Remove the service (self leave/remove other service)
	RemoveMember(ctx context.Context, resourceType string, resourceID uint, memberInfo uint,
		memberType models.MemberType) error
	// ListMember list all the service of the resource
	ListMember(ctx context.Context, resourceType string, resourceID uint) ([]models.Member, error)
}

type service struct {
	memberManager member.Manager
	groupManager  groupManager.Manager
}

func NewService() Service {
	return &service{
		memberManager: member.Mgr,
		groupManager:  groupManager.Mgr,
	}
}

func (s *service) GetUserMember(ctx context.Context, resourceType string, resourceID uint) (*models.Member, error) {
	var currentUser userauth.User
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.getMember(ctx, resourceType, resourceID, currentUser.GetID(), models.MemberUser)
}

func (s *service) getMember(ctx context.Context, resourceType string, resourceID uint,
	memberInfo uint, memberType models.MemberType) (*models.Member, error) {
	members, err := s.ListMember(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	for _, item := range members {
		if item.MemberType == memberType &&
			item.MemberInfo == memberInfo {
			return &item, nil
		}
	}
	return nil, nil
}

func (s *service) getDirectMemberByDetail(ctx context.Context,
	resourceType string, resourceID uint,
	memberInfo uint, memberType models.MemberType) (*models.Member, error) {
	members, err := s.ListMember(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	for _, item := range members {
		if string(item.ResourceType) == resourceType &&
			item.ResourceID == resourceID &&
			item.MemberType == memberType &&
			item.MemberInfo == memberInfo {
			return &item, nil
		}
	}
	return nil, nil
}

func (s *service) CreateMember(ctx context.Context, postMember PostMember) (*models.Member, error) {
	var currentUser userauth.User
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 1. convert service
	member, err := ConvertPostMemberToMember(postMember, currentUser)
	if err != nil {
		return nil, err
	}

	// 2. If service exist return error TODO(tom): change to updateMember
	memberItem, err := s.getDirectMemberByDetail(ctx, postMember.ResourceType, postMember.ResourceID,
		postMember.MemberInfo, postMember.MemberType)
	if err != nil {
		return nil, err
	}
	if memberItem != nil {
		return nil, ErrMemberExist
	}

	// 3. check if the current user can grant the role
	var userMemberInfo *models.Member
	userMemberInfo, err = s.GetUserMember(ctx, postMember.ResourceType, postMember.ResourceID)
	if err != nil {
		return nil, err
	}
	if userMemberInfo == nil {
		return nil, ErrNotPermitted
	}

	RoleEqualOrBigger := func(role1, role2 string) bool {
		return true
	}
	if !RoleEqualOrBigger(userMemberInfo.Role, postMember.Role) {
		return nil, ErrGrantHighRole
	}

	return s.memberManager.Create(ctx, member)
}

// createMemberDirect for unit test
func (s *service) createMemberDirect(ctx context.Context, postMember PostMember) (*models.Member, error) {
	var currentUser userauth.User
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 1. convert service
	member, err := ConvertPostMemberToMember(postMember, currentUser)
	if err != nil {
		return nil, err
	}
	return s.memberManager.Create(ctx, member)
}

func (s *service) RemoveMember(ctx context.Context, resourceType string, resourceID uint, memberInfo uint,
	memberType models.MemberType) error {
	// 1. get current user and check the permission
	var currentUser userauth.User
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return err
	}

	// 2. If service not exist return error
	memberItem, err := s.getDirectMemberByDetail(ctx, resourceType, resourceID,
		memberInfo, memberType)
	if err != nil {
		return err
	}
	if memberItem == nil {
		return ErrMemberNotExist
	}

	// 3. self level the group
	if memberType == models.MemberUser &&
		memberInfo == currentUser.GetID() {
		return s.memberManager.DeleteMember(ctx, resourceID)
	}

	// 4. test if current user  have the permission to remove the
	var currentUserInfo *models.Member
	currentUserInfo, err = s.GetUserMember(ctx, resourceType, resourceID)
	if err != nil {
		return err
	}
	if currentUserInfo == nil {
		return ErrNotPermitted
	}

	RoleBigger := func(role1, role2 string) bool {
		return true
	}
	if !RoleBigger(currentUserInfo.Role, memberItem.Role) {
		return ErrRemoveHighRole
	}

	return s.memberManager.DeleteMember(ctx, memberItem.ID)
}

// UpdateMember update exist service entry
// user can only attach a role not higher than self
func (s *service) UpdateMember(ctx context.Context, resourceType string, resourceID uint,
	memberInfo uint, memberType models.MemberType, role string) (*models.Member, error) {
	// 1. get current user and check the permission
	var currentUser userauth.User
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 2. If service not exist return
	memberItem, err := s.getDirectMemberByDetail(ctx, resourceType, resourceID,
		memberInfo, memberType)
	if err != nil {
		return nil, err
	}
	if memberItem == nil {
		return nil, ErrMemberNotExist
	}

	// 3. check if the grant current user can grant the role
	var userMemberInfo *models.Member
	userMemberInfo, err = s.getDirectMemberByDetail(ctx, resourceType, resourceID,
		currentUser.GetID(), models.MemberUser)
	if err != nil {
		return nil, err
	}
	RoleEqualOrBigger := func(role1, role2 string) bool {
		return true
	}
	if !RoleEqualOrBigger(userMemberInfo.Role, role) {
		return nil, ErrGrantHighRole
	}

	return s.memberManager.UpdateByID(ctx, memberItem.ResourceID, role)
}

func (s *service) ListMember(ctx context.Context, resourceType string, resourceID uint) ([]models.Member, error) {
	// get all the members
	var allMembers []models.Member
	var err error
	switch resourceType {
	case models.TypeGroupStr:
		allMembers, err = s.listGroupMembers(ctx, resourceID)
	case models.TypeApplicationStr:
		allMembers, err = s.listApplicationMembers(ctx, resourceID)
	case models.TypeApplicationInstanceStr:
		allMembers, err = s.listApplicationInstanceMembers(ctx, resourceID)
	default:
		err = errors.New("unsupported resourceType")
	}
	if err != nil {
		return nil, err
	}
	return allMembers, nil
}

func (s *service) listGroupMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	// 1. list all the groups of the group
	groupInfo, err := s.groupManager.GetByID(ctx, uint(resourceID))
	if err != nil {
		return nil, err
	}
	groupIDs := groupManager.FormatIDsFromTraversalIDs(groupInfo.TraversalIDs)

	// 2. get all the direct service of group
	var retMembers []models.Member

	for i := len(groupIDs) - 1; i >= 0; i-- {
		members, err := s.memberManager.ListDirectMember(ctx, models.TypeGroup, groupIDs[i])
		if err != nil {
			return nil, err
		}
		retMembers = append(retMembers, members...)
	}

	return DeduplicateMember(retMembers), nil
}

func DeduplicateMember(members []models.Member) []models.Member {
	// deduplicate by memberType, memberInfo
	memberMap := make(map[string]models.Member)

	var retMembers []models.Member
	for _, item := range members {
		key := strconv.Itoa(int(item.MemberType)) + "-" + strconv.FormatUint(uint64(item.MemberInfo), 10)
		_, ok := memberMap[key]
		if !ok {
			memberMap[key] = item
			retMembers = append(retMembers, item)
		}
	}
	return retMembers
}

func (s *service) listApplicationMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	// TODO(tom)
	// 1. query the application's service
	// 2. query the group's service
	// 3. merge and return
	err := errors.New("Unimplement yet")
	return nil, err
}

func (s *service) listApplicationInstanceMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	// TODO(tom)
	// 1. query the applicationinstance's service
	// 2. query the application's service
	// 3. merge and return
	err := errors.New("Unimplement yet")
	return nil, err
}
