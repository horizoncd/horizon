package access

import (
	"context"
	"net/http"

	"g.hz.netease.com/horizon/core/common"
	hauth "g.hz.netease.com/horizon/pkg/auth"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/rbac"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/middleware/auth"
)

type Controller interface {
	// Review return access review results for apis
	Review(ctx context.Context, apis []API) (map[string]map[string]*ReviewResult, error)
}

type controller struct {
	requestInfoFty auth.RequestInfoFactory
	authorizer     rbac.Authorizer
	skippers       []middleware.Skipper
}

var _ Controller = (*controller)(nil)

func NewController(authorizer rbac.Authorizer,
	skippers ...middleware.Skipper) Controller {
	return &controller{
		requestInfoFty: auth.RequestInfoFty,
		authorizer:     authorizer,
		skippers:       skippers,
	}
}

func (c *controller) Review(ctx context.Context, apis []API) (map[string]map[string]*ReviewResult, error) {
	reviewResponse := make(map[string]map[string]*ReviewResult)
	// get user info
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, perror.WithMessage(err, "failed to get user info")
	}
	// traverse apis to authorize
	for _, api := range apis {
		if _, ok := reviewResponse[api.URL]; !ok {
			reviewResponse[api.URL] = make(map[string]*ReviewResult)
		}
		if _, ok := reviewResponse[api.URL][api.Method]; ok {
			continue
		}
		reviewResult := &ReviewResult{}
		reviewResponse[api.URL][api.Method] = reviewResult

		// 1.new request by api
		req, err := http.NewRequest(api.Method, api.URL, nil)
		if err != nil {
			return nil, perror.Wrapf(err, "invalid api, url: %s, method: %s", api.URL, api.Method)
		}
		// 2.check skippers
		for _, skipper := range c.skippers {
			if skipper(req) {
				reviewResult.Allowed = true
				break
			}
		}
		// 3.if skipped, review next api
		if reviewResult.Allowed {
			continue
		}
		// 4.new request info by request
		requestInfo, err := c.requestInfoFty.NewRequestInfo(req)
		if err != nil {
			return nil, perror.WithMessagef(err, "invalid api, url: %s, method: %s", api.URL, api.Method)
		}
		// 4. do rbac auth
		authRecord := hauth.AttributesRecord{
			User:            currentUser,
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
		decision, reason, err := c.authorizer.Authorize(ctx, authRecord)
		if err != nil {
			return nil, perror.WithMessagef(err, "failed to authorize, url: %s, method: %s", api.URL, api.Method)
		}
		reviewResult.Allowed = decision == hauth.DecisionAllow
		reviewResult.Reason = reason
	}

	return reviewResponse, nil
}
