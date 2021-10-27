package rbac

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/pkg/auth"
	"g.hz.netease.com/horizon/pkg/member/models"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/rbac/types"
	"g.hz.netease.com/horizon/pkg/util/log"
)

var (
	ErrMemberNotExist = errors.New("Member Not Exist")
)

// Authorizer use the basic rbac rules to check if the user
// have the permissions
type Authorizer interface {
	Authorize(ctx context.Context, attributes auth.Attributes) (auth.Decision, string, error)
}

type VisitorFunc func(fmt.Stringer, *types.PolicyRule, error) bool

func NewAuthorizer(roleservice role.Service, memberservice memberservice.Service) Authorizer {
	return &authorizer{
		roleService:   roleservice,
		memberService: memberservice,
	}
}

type authorizer struct {
	roleService   role.Service
	memberService memberservice.Service
}

const (
	NotChecked        = "not checked"
	ResourceFormatErr = "format error"
	InternalError     = "internal error"
	MemberNotExist    = "member not exist"
	RoleNotExist      = "role not exist"
)

// /apis/core/v1/applications/
// /apis/core/v1/clusters/
// /apis/core/v1/groups/
// /apis/core/v1/members/
// /apis/core/v1/pipelineruns/
func (a *authorizer) Authorize(ctx context.Context, attr auth.Attributes) (auth.Decision,
	string, error) {
	// TODO(tom): members and pipelineruns need to add to auth check
	if attr.IsResourceRequest() && (attr.GetResource() == "members" ||
		attr.GetResource() == "pipelineruns") {
		log.Warning(ctx,
			"/apis/core/v1/members/{memberid} and /apis/core/v1/pipelineruns/{pipelineruns} are not authed")
		return auth.DecisionAllow, NotChecked, nil
	}

	// 1. get the member
	resourceIDStr := attr.GetName()
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		log.Errorf(ctx, "resourceType = %s, resourceID = %s, format error\n", attr.GetResource(), attr.GetName())
		return auth.DecisionDeny, ResourceFormatErr, err
	}

	member, err := a.memberService.GetMemberOfResource(ctx, attr.GetResource(), uint(resourceID))
	if err != nil {
		log.Errorf(ctx, "GetMemberOfResource error, resourceType = %s, resourceID = %s, user = %s\n",
			attr.GetResource(), attr.GetName(), attr.GetUser().String())
		return auth.DecisionDeny, InternalError, err
	}
	if member == nil {
		log.Warningf(ctx, " user %s member not found of resourceType = %s, resourceID = %s",
			attr.GetUser().String(), attr.GetResource(), attr.GetName())
		return auth.DecisionDeny, MemberNotExist, nil
	}

	// 2. get the role
	role, err := a.roleService.GetRole(ctx, member.Role)
	if err != nil {
		log.Errorf(ctx, "get role file err = %+v", err)
		return auth.DecisionDeny, InternalError, err
	}
	if role == nil {
		return auth.DecisionDeny, RoleNotExist, nil
	}

	// 3. check the permission
	return VisitRoles(member, role, attr)
}

func VisitRoles(member *models.Member, role *types.Role,
	attr auth.Attributes) (_ auth.Decision, reason string, err error) {
	for i, rule := range role.PolicyRules {
		if types.RuleAllow(attr, &rule) {
			reason = fmt.Sprintf("user %s allowd by member(%s) by rule[%d]",
				attr.GetUser().String(), member.BaseInfo(), i)
			return auth.DecisionAllow, reason, nil
		}
	}
	reason = fmt.Sprintf("user %s denied by member(%s)", attr.GetUser().String(), member.BaseInfo())
	return auth.DecisionDeny, reason, nil
}
