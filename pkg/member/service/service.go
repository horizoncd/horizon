package service

import (
	"context"
	"errors"
	"strconv"

	"g.hz.netease.com/horizon/core/middleware/user"
	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	"g.hz.netease.com/horizon/pkg/member"
	"g.hz.netease.com/horizon/pkg/member/models"
	roleservice "g.hz.netease.com/horizon/pkg/rbac/role"
)

var (
	ErrMemberExist    = errors.New("MemberExist")     // "Member Exist"
	ErrNotPermitted   = errors.New("NotPermitted")    // "Not Permitted"
	ErrMemberNotExist = errors.New("MemberNotExist")  // "Member not exists"
	ErrGrantHighRole  = errors.New("GrantHigherRole") // "Grant higher role"
	ErrRemoveHighRole = errors.New("RemoveHighRole")  // "Remove higher role"
)

type Service interface {
	// CreateMember post a new member
	CreateMember(ctx context.Context, postMember PostMember) (*models.Member, error)
	// GetMember return the current user member of direct or parent
	GetMember(ctx context.Context, memberID uint) (*models.Member, error)
	// UpdateMember update the member by the memberID
	UpdateMember(ctx context.Context, memberID uint, role string) (*models.Member, error)
	// RemoveMember Remove the member by the memberID
	RemoveMember(ctx context.Context, memberID uint) error
	// ListMember list all the member of the resource
	ListMember(ctx context.Context, resourceType string, resourceID uint) ([]models.Member, error)

	// GetMemberOfResource return the current user's role of the resource (member from direct or parent)
	GetMemberOfResource(ctx context.Context, resourceType string, resourceID uint) (*models.Member, error)
}

type service struct {
	memberManager      member.Manager
	groupManager       groupmanager.Manager
	applicationManager applicationmanager.Manager
	roleService        roleservice.Service
}

func NewService(roleService roleservice.Service) Service {
	return &service{
		memberManager:      member.Mgr,
		groupManager:       groupmanager.Mgr,
		applicationManager: applicationmanager.Mgr,
		roleService:        roleService,
	}
}

func (s *service) CreateMember(ctx context.Context, postMember PostMember) (*models.Member, error) {
	var currentUser userauth.User
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 1. check exist
	memberItem, err := s.memberManager.Get(ctx, models.ResourceType(postMember.ResourceType), postMember.ResourceID,
		postMember.MemberType, postMember.MemberInfo)
	if err != nil {
		return nil, err
	}
	if memberItem != nil {
		// if member exist, try to update the member
		return s.UpdateMember(ctx, memberItem.ID, postMember.Role)
	}

	// 2. check if current user can create the role
	var userMemberInfo *models.Member
	userMemberInfo, err = s.getMember(ctx, postMember.ResourceType,
		postMember.ResourceID, models.MemberUser, currentUser.GetID())
	if err != nil {
		return nil, err
	}
	if userMemberInfo == nil {
		return nil, ErrNotPermitted
	}

	comResult, err := s.roleService.RoleCompare(ctx, userMemberInfo.Role, postMember.Role)
	if err != nil {
		return nil, err
	}
	if comResult == roleservice.RoleSmaller {
		return nil, ErrGrantHighRole
	}

	// 3. do create  member
	member, err := ConvertPostMemberToMember(postMember, currentUser)
	if err != nil {
		return nil, err
	}
	return s.memberManager.Create(ctx, member)
}

func (s *service) GetMemberOfResource(ctx context.Context,
	resourceType string, resourceID uint) (*models.Member, error) {
	var currentUser userauth.User
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.getMember(ctx, resourceType, resourceID, models.MemberUser, currentUser.GetID())
}

func (s *service) GetMember(ctx context.Context, memberID uint) (*models.Member, error) {
	return s.memberManager.GetByID(ctx, memberID)
}

func (s *service) RemoveMember(ctx context.Context, memberID uint) error {
	var currentUser userauth.User
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return err
	}

	// 1. get member
	memberItem, err := s.memberManager.GetByID(ctx, memberID)
	if err != nil {
		return err
	}
	if memberItem == nil {
		return ErrMemberNotExist
	}

	// 2. check if the grant current user can remove the member
	var userMemberInfo *models.Member
	userMemberInfo, err = s.getMember(ctx, string(memberItem.ResourceType),
		memberItem.ResourceID, models.MemberUser, currentUser.GetID())
	if err != nil {
		return err
	}
	if userMemberInfo == nil {
		return ErrNotPermitted
	}

	comResult, err := s.roleService.RoleCompare(ctx, userMemberInfo.Role, memberItem.Role)
	if err != nil {
		return err
	}
	if comResult == roleservice.RoleSmaller {
		return ErrRemoveHighRole
	}

	return s.memberManager.DeleteMember(ctx, memberID)
}

func (s *service) UpdateMember(ctx context.Context, memberID uint, role string) (*models.Member, error) {
	var currentUser userauth.User
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 1. get the member
	memberItem, err := s.memberManager.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if memberItem == nil {
		return nil, ErrMemberNotExist
	}

	// 2. check if the current user have the permission to update the role
	var userMemberInfo *models.Member
	userMemberInfo, err = s.getMember(ctx, string(memberItem.ResourceType),
		memberItem.ResourceID, models.MemberUser, currentUser.GetID())
	if err != nil {
		return nil, err
	}
	if userMemberInfo == nil {
		return nil, ErrNotPermitted
	}

	comResult, err := s.roleService.RoleCompare(ctx, userMemberInfo.Role, memberItem.Role)
	if err != nil {
		return nil, err
	}
	if comResult != roleservice.RoleBigger {
		return nil, ErrNotPermitted
	}

	comResult, err = s.roleService.RoleCompare(ctx, userMemberInfo.Role, role)
	if err != nil {
		return nil, err
	}
	if comResult == roleservice.RoleSmaller {
		return nil, ErrGrantHighRole
	}

	// 3. update the role
	return s.memberManager.UpdateByID(ctx, memberItem.ID, role)
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
	case models.TypeApplicationClusterStr:
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
	groupIDs := groupmanager.FormatIDsFromTraversalIDs(groupInfo.TraversalIDs)

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
		key := strconv.Itoa(int(item.MemberType)) + "-" + strconv.FormatUint(uint64(item.MemberNameID), 10)
		_, ok := memberMap[key]
		if !ok {
			memberMap[key] = item
			retMembers = append(retMembers, item)
		}
	}
	return retMembers
}

// getMember return the direct member or member from the parent
func (s *service) getMember(ctx context.Context, resourceType string, resourceID uint,
	memberType models.MemberType, memberInfo uint) (*models.Member, error) {
	members, err := s.ListMember(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	for _, item := range members {
		if item.MemberType == memberType &&
			item.MemberNameID == memberInfo {
			return &item, nil
		}
	}
	return nil, nil
}

func (s *service) listApplicationMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	var retMembers []models.Member
	// 1. query the application's service
	applicationInfo, err := s.applicationManager.GetByID(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	members, err := s.memberManager.ListDirectMember(ctx, models.TypeApplication, resourceID)
	if err != nil {
		return nil, err
	}
	retMembers = append(retMembers, members...)
	// 2. query the group's service
	members, err = s.listGroupMembers(ctx, applicationInfo.GroupID)
	if err != nil {
		return nil, err
	}

	retMembers = append(retMembers, members...)

	return DeduplicateMember(retMembers), nil
}

func (s *service) listApplicationInstanceMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	// TODO(tom)
	// 1. query the applicationinstance's service
	// 2. query the application's service
	// 3. merge and return
	err := errors.New("Unimplement yet")
	return nil, err
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
