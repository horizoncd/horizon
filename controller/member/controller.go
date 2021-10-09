package member

import (
	"context"
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/q"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/member"
	"g.hz.netease.com/horizon/pkg/member/models"
	horizonerror "g.hz.netease.com/horizon/util/errors"
	"g.hz.netease.com/horizon/util/wlog"
)

const (
	_errCodeHigherGrant  = horizonerror.ErrorCode("GrantHigherRoleNotAllow")
	_errCodeResourceType = horizonerror.ErrorCode("ResourceTypeNotSupport")
)

var (
	Ctl = NewController()
)

type Controller interface {
	// CreateMember post a new member
	CreateMember(ctx context.Context, member PostMember) (*Member, error)

	// UpdateMember update exist member entry
	// user can only attach a role not higher than self
	UpdateMember(ctx context.Context, memberID int, Role string) (*Member, error)

	// ListMember list all the member of the resource
	ListMember(ctx context.Context, resourceType, resourceID string, query *q.Query) ([]Member, error)

	// GetMemberByUserName return the rolebinding
	GetMemberByUserName(ctx context.Context, resourceType, resourceID string) ([]Member, error)
}

type controller struct {
	manager member.Manager
}

func NewController() Controller {
	return &controller{}
}

func grantRoleCheck(currentUserRole, grantRole string) bool {
	return true
}

func (c *controller) CreateMember(ctx context.Context, postMember PostMember) (_ *Member, err error) {
	const op = "member controller: createMember"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })
	// 1. get the user member info and
	// check if current user have the permission to add such role
	var currentUser userauth.User
	currentUser, err = user.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 2. convert and add the member
	member, err := ConvertPostMemberToMember(postMember, currentUser)
	if err != nil {
		return nil, horizonerror.E(op, http.StatusBadRequest, _errCodeResourceType,
			err.Error())
	}

	var userMemberInfo *models.Member
	userMemberInfo, err = c.manager.GetByUserName(ctx, member.ResourceType, postMember.ResourceID, currentUser.GetName())
	if err != nil {
		return nil, err
	}

	if !grantRoleCheck(userMemberInfo.Role, postMember.Role) {
		return nil, horizonerror.E(op, http.StatusForbidden, _errCodeHigherGrant,
			fmt.Sprintf("user %v cannot assign higher role {%v} than self role {%v}",
				currentUser.GetName(), postMember.Role, userMemberInfo.Role))
	}


	ret, err := c.manager.Create(ctx, member)
	if err != nil {
		return nil, horizonerror.E(op, http.StatusInternalServerError, err.Error())
	}

	memberResponse := ConvertMember(ret, "direct")
	return &memberResponse, nil
}

// UpdateMember update exist member entry
// user can only attach a role not higher than self
func (c *controller) UpdateMember(ctx context.Context, memberID int, Role string) (*Member, error) {
	return nil, nil
}

// ListMember list all the member of the resource
func (c *controller) ListMember(ctx context.Context, resourceType, resourceID string, query *q.Query) ([]Member, error) {
	return nil, nil
}

// GetMemberByUserName return the rolebinding
func (c *controller) GetMemberByUserName(ctx context.Context, resourceType, resourceID string) ([]Member, error) {
	return nil, nil
}
