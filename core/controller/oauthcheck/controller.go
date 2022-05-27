package oauthcheck

import (
	"fmt"
	"strconv"
	"strings"

	herrors "g.hz.netease.com/horizon/core/errors"
	rbactype "g.hz.netease.com/horizon/pkg/auth"
	"g.hz.netease.com/horizon/pkg/authentication/user"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/oauth/manager"
	"g.hz.netease.com/horizon/pkg/oauth/scope"
	"g.hz.netease.com/horizon/pkg/rbac/types"
	"g.hz.netease.com/horizon/pkg/server/middleware/auth"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	"golang.org/x/net/context"
)

type Controller interface {
	LoadAccessTokenUser(ctx context.Context, token string) (user.User, error)
	CheckScopePermission(ctx context.Context, token string, authInfo auth.RequestInfo) (bool, string, error)
}

type controller struct {
	oauthManager manager.Manager
	userManager  usermanager.Manager
	scopeService scope.Service
}

var _ Controller = &controller{}

func NewOauthChecker(oauthManager manager.Manager,
	manager2 usermanager.Manager, service scope.Service) Controller {
	return &controller{
		oauthManager: oauthManager,
		userManager:  manager2,
		scopeService: service,
	}
}

func (c *controller) LoadAccessTokenUser(ctx context.Context, accessToken string) (user.User, error) {
	token, err := c.oauthManager.LoadAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// TODO: robotID
	userID, err := func() (uint64, error) {
		userIDInstr := token.UserOrRobotIdentity
		return strconv.ParseUint(userIDInstr, 10, 0)
	}()
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrOAuthInternal,
			"userID can not convert, userID = %s", token.UserOrRobotIdentity)
	}

	usr, err := c.userManager.GetUserByID(ctx, uint(userID))
	if err != nil {
		return nil, err
	}
	return &user.DefaultInfo{
		Name:     usr.Name,
		FullName: usr.FullName,
		ID:       usr.ID,
		Email:    usr.Email,
		Admin:    usr.Admin,
	}, nil
}

func (c *controller) CheckScopePermission(ctx context.Context, accessToken string,
	requestInfo auth.RequestInfo) (bool, string, error) {
	token, err := c.oauthManager.LoadAccessToken(ctx, accessToken)
	if err != nil {
		return false, "", err
	}

	user, err := c.LoadAccessTokenUser(ctx, accessToken)
	if err != nil {
		return false, "", err
	}

	record := rbactype.AttributesRecord{
		User:            user,
		Verb:            requestInfo.Verb,
		APIGroup:        requestInfo.APIGroup,
		APIVersion:      requestInfo.APIVersion,
		Resource:        requestInfo.Resource,
		SubResource:     requestInfo.Subresource,
		Name:            requestInfo.Name,
		Scope:           requestInfo.Scope,
		ResourceRequest: requestInfo.IsResourceRequest,
		Path:            requestInfo.Path,
	}

	scopeRoles := c.scopeService.GetRulesByScope(strings.Split(token.Scope, " "))
	for _, scopeRule := range scopeRoles {
		for i, policy := range scopeRule.PolicyRules {
			if types.RuleAllow(record, &policy) {
				reason := fmt.Sprintf("user %s allowd by scope(%s) by rule[%d]",
					user.String(), scopeRule.Name, i)
				return true, reason, nil
			}
		}
	}
	return false, "", nil
}
