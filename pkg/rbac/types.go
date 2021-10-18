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
	"fmt"

	"g.hz.netease.com/horizon/pkg/auth"
	"g.hz.netease.com/horizon/pkg/authentication/user"
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

type Role struct {
	Name        string
	PolicyRules []PolicyRule
}

type PolicyRule struct {
	Verbs           []string
	APIGroups       []string
	Resources       []string
	Scopes          []string
	NonResourceURLs []string
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

// authorizingVisitor short-circuits once allowed, and collects any resolution errors encountered
type authorizingVisitor struct {
	requestAttributes auth.Attributes

	allowed bool
	reason  string
	errors  []error
}

func (v *authorizingVisitor) visit(source fmt.Stringer, rule *PolicyRule, err error) bool {
	if rule != nil && RuleAllow(v.requestAttributes, rule) {
		v.allowed = true
		if source != nil {
			v.reason = source.String()
		}
	}
	if err != nil {
		v.errors = append(v.errors, err)
	}
	return true
}

// Authorizer use the basic rbac rules to check if the user
// have the permission
type Authorizer struct {
	authorizationRuleResolver RequestToRuleResolver
}

type VisitorFunc func(fmt.Stringer, *PolicyRule, error) bool

type RequestToRuleResolver interface {
	// VisitRulesFor invokes visitor() with each rule that applies to a given user
	VisitRulesFor(user user.User, visitor VisitorFunc)
}

func (r *Authorizer) Authorize(ctx context.Context, attributes auth.Attributes) (auth.Decision,
	string, error) {
	ruleCheckingVisitor := &authorizingVisitor{requestAttributes: attributes}

	r.authorizationRuleResolver.VisitRulesFor(attributes.GetUser(), ruleCheckingVisitor.visit)

	if ruleCheckingVisitor.allowed {
		return auth.DecisionAllow, ruleCheckingVisitor.reason, nil
	}

	// Build a detailed log of the denial.
	reason := ruleCheckingVisitor.reason
	return auth.DecisionDeny, reason, nil
}
