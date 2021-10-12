package member

import (
	"context"
	"errors"
	"sort"

	"g.hz.netease.com/horizon/core/middleware/user"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	groupManager "g.hz.netease.com/horizon/pkg/group/manager"
	"g.hz.netease.com/horizon/pkg/member"
	"g.hz.netease.com/horizon/pkg/member/models"
)

var (
	ErrRoleReduced   = errors.New("RoleReduced")     // "Role Reduced"
	ErrMemberExist   = errors.New("MemberExist")     // "Member Exist"
	ErrGrantHighRole = errors.New("GrantHigherRole") // "Grant higher role"
)

var (
	Ctl = NewService()
)

type Service interface {
	// GetMember get the member of the current user ()
	GetMember(ctx context.Context) (*models.Member, error)
	// CreateMember post a new member
	CreateMember(ctx context.Context, postMember PostMember) (*models.Member, error)
	// UpdateMember update exist member entry
	// user can only attach a role not higher than self
	UpdateMember(ctx context.Context, groupID uint, memberID uint, Role string) (*models.Member, error)
	// ListMember list all the member of the resource
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

func grantRoleCheck(currentUserRole, grantRole string) bool {
	return true
}

func (s *service) GetMember(ctx context.Context) (*models.Member, error) {
	return nil, nil
}

func (s *service) CreateMember(ctx context.Context, postMember PostMember) (*models.Member, error) {
	var currentUser userauth.User
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 1. convert member
	member, err := ConvertPostMemberToMember(postMember, currentUser)
	if err != nil {
		return nil, err
	}

	// 2. If member exist in check if Role Can Create
	members, err := s.ListMember(ctx, models.TypeGroupStr, postMember.ResourceID)
	if err != nil {
		return nil, nil
	}

	index := sort.Search(len(members), func(i int) bool {
		return members[i].MemberType == postMember.MemberType &&
			members[i].MemberInfo == postMember.MemberInfo
	})

	// TODO(tom): if exist change to update
	if members[index].ResourceType == models.ResourceType(postMember.ResourceType) &&
		members[index].ResourceID == postMember.ResourceID {
		return nil, ErrMemberExist
	}

	RoleBigger := func(role1, role2 string) bool {
		return true
	}
	if index < len(members) {
		// only higher role grant is permitted
		if !RoleBigger(postMember.Role, members[index].Role) {
			return nil, ErrRoleReduced
		}
	}

	// 3. check if the grant higher
	var userMemberInfo *models.Member
	userMemberInfo, err = s.memberManager.GetByUserName(ctx, member.ResourceType, postMember.ResourceID, currentUser.GetName())
	if err != nil {
		return nil, err
	}
	RoleEqualOrBigger := func(role1, role2 string) bool {
		return true
	}
	if !RoleEqualOrBigger(userMemberInfo.Role, postMember.Role) {
		return nil, ErrGrantHighRole
	}

	return s.memberManager.Create(ctx, member)
}

// UpdateMember update exist member entry
// user can only attach a role not higher than self
func (s *service) UpdateMember(ctx context.Context, groupID uint, memberID uint, Role string) (*models.Member, error) {
	// 1. get current user and check the permission
	var currentUser userauth.User
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var userMemberInfo *models.Member
	userMemberInfo, err = s.memberManager.GetByUserName(ctx, models.TypeGroup, groupID, currentUser.GetName())
	if err != nil {
		return nil, err
	}

	if !grantRoleCheck(userMemberInfo.Role, Role) {
		return nil, nil
	}
	return nil, nil
}

func DeduplicateMemberBasedOnRole(members []models.Member) []models.Member {
	var retMember []models.Member

	RoleBigger := func(role1, role2 string) bool {
		return true
	}

	for _, member := range members {
		index := sort.Search(len(retMember), func(i int) bool {
			return retMember[i].MemberType == member.MemberType &&
				retMember[i].MemberInfo == member.MemberInfo &&
				RoleBigger(retMember[i].Role, member.Role)
		})

		if index < len(retMember) {
			retMember[index] = member
		}
	}
	return retMember
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
		allMembers, err = s.listApplicationMembers(ctx, resourceID)
	default:
		err = errors.New("unsupported resourceType")
	}
	if err != nil {
		return nil, err
	}

	// deduplicate based on the highest role
	return DeduplicateMemberBasedOnRole(allMembers), nil
}

func (s *service) listGroupMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	// 1. list all the groups of the group
	groupInfo, err := s.groupManager.GetByID(ctx, uint(resourceID))
	if err != nil {
		return nil, err
	}
	groupIDs := groupManager.FormatIDsFromTraversalIDs(groupInfo.TraversalIDs)

	// 2. get all the direct member of group
	var retMembers []models.Member
	for _, groupID := range groupIDs {
		members, err := s.memberManager.ListDirectMember(ctx, models.TypeGroup, groupID)
		if err != nil {
			return nil, err
		}
		retMembers = append(retMembers, members...)
	}
	return retMembers, nil
}

func (s *service) listApplicationMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	// TODO(tom)
	// 1. query the application's member
	// 2. query the group's member
	// 3. merge and return
	err := errors.New("Unimplement yet")
	return nil, err
}

func (s *service) listApplicationInstanceMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	// TODO(tom)
	// 1. query the applicationinstance's member
	// 2. query the application's member
	// 3. merge and return
	err := errors.New("Unimplement yet")
	return nil, err
}
