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

package access

import (
	"context"
	"net/http"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/middleware"
	"github.com/horizoncd/horizon/core/middleware/prehandle"
	hauth "github.com/horizoncd/horizon/pkg/auth"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/rbac"
)

type Controller interface {
	// Review return access review results for apis
	Review(ctx context.Context, apis []API) (map[string]map[string]*ReviewResult, error)
}

type controller struct {
	requestInfoFty hauth.RequestInfoFactory
	authorizer     rbac.Authorizer
	skippers       []middleware.Skipper
}

var _ Controller = (*controller)(nil)

func NewController(authorizer rbac.Authorizer,
	skippers ...middleware.Skipper) Controller {
	return &controller{
		requestInfoFty: prehandle.RequestInfoFty,
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
