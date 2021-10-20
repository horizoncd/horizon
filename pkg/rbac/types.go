/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package rbac

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/pkg/auth"
	"g.hz.netease.com/horizon/pkg/member/models"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/util/log"
)

// attention: rbac is refers to the kubernetes rbac
// copy core struct and logics from the kubernetes code
// and do same modify
const (
	APIGroupAll    = "*"
	ResourceAll    = "*"
	VerbAll        = "*"
	ScopeAll       = "*"
	NonResourceAll = "*"
)

var (
	ErrMemberNotExist = errors.New("Member Not Exist")
)

type Role struct {
	Name        string       `yaml:"name"`
	PolicyRules []PolicyRule `yaml:"rules"`
}

type PolicyRule struct {
	Verbs           []string `yaml:"verbs"`
	APIGroups       []string `yaml:"apiGroups"`
	Resources       []string `yaml:"resources"`
	Scopes          []string `yaml:"scopes"`
	NonResourceURLs []string `yaml:"nonResourceURLs"`
}

func RuleAllow(attribute auth.Attributes, rule *PolicyRule) bool {
	if attribute.IsResourceRequest() {
		combinedResource := attribute.GetResource()
		if len(attribute.GetSubResource()) > 0 {
			combinedResource = attribute.GetResource() + "/" + attribute.GetSubResource()
		}
		return VerbMatches(rule, attribute.GetVerb()) &&
			APIGroupMatches(rule, attribute.GetAPIGroup()) &&
			ResourceMatches(rule, combinedResource, attribute.GetSubResource()) &&
			ScopeMatches(rule, attribute.GetScope())
	}
	return VerbMatches(rule, attribute.GetVerb()) &&
		NonResourceURLMatches(rule, attribute.GetPath())
}

// Authorizer use the basic rbac rules to check if the user
// have the permissions
type Authorizer interface {
	Authorize(ctx context.Context, attributes auth.Attributes) (auth.Decision, string, error)
}

type VisitorFunc func(fmt.Stringer, *PolicyRule, error) bool

type authorizer struct {
	roleService   Service
	memberService memberservice.Service
}

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
		log.Warn(ctx,
			"/apis/core/v1/members/{memberid} and /apis/core/v1/pipelineruns/{pipelineruns} are not authed")
		return auth.DecisionAllow, "not checked", nil
	}

	// 1. get the member
	resourceIDStr := attr.GetName()
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		log.Errorf(ctx, "resourceType = %s, resourceID = %s, format error\n", attr.GetResource(), attr.GetName())
		return auth.DecisionDeny, "resourceId Format Error", err
	}

	member, err := a.memberService.GetMemberOfResource(ctx, attr.GetResource(), uint(resourceID))
	if err != nil {
		log.Errorf(ctx, "GetMemberOfResource error, resourceType = %s, resourceID = %s, user = %s\n",
			attr.GetResource(), attr.GetName(), attr.GetUser().String())
		return auth.DecisionDeny, "getMember return error", err
	}
	if member == nil {
		log.Warningf(ctx, " user %s member not found of resourceType = %s, resourceID = %s",
			attr.GetUser().String(), attr.GetResource(), attr.GetName())
		return auth.DecisionDeny, "member not exist", nil
	}

	// 2. get the role
	role, err := a.roleService.GetRole(ctx, member.Role)
	if err != nil {
		log.Errorf(ctx, "get role file err = %+v", err)
		return auth.DecisionDeny, "GetRole Failed", err
	}
	if role == nil {
		return auth.DecisionDeny, "member not found", ErrMemberNotExist
	}

	// 3. check the permission
	return VisitRoles(member, role, attr)
}

func VisitRoles(member *models.Member, role *Role, attri auth.Attributes) (_ auth.Decision, reason string, err error) {
	for i, rule := range role.PolicyRules {
		if RuleAllow(attri, &rule) {
			reason = fmt.Sprintf("user %s allowd by member(%s) by rule[%d]",
				attri.GetUser().String(), member.BaseInfo(), i)
			return auth.DecisionAllow, reason, nil
		}
	}
	reason = fmt.Sprintf("user %s deny by member(%s)", attri.GetUser().String(), member.BaseInfo())
	return auth.DecisionDeny, reason, nil
}
