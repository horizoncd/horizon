package rbac

import (
	"context"
	"fmt"

	"g.hz.netease.com/horizon/pkg/auth"
	"g.hz.netease.com/horizon/pkg/authentication/user"
)

// attention: rbac is refers to the kubernetes rbac
// we just copy core struct and logics from the kubernetes code
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

// RBACAuthorizer use the basic rbac rules to check if the user
// have the permission
type RBACAuthorizer struct {
	authorizationRuleResolver RequestToRuleResolver
}

type VisitorFunc func(fmt.Stringer, *PolicyRule, error) bool

type RequestToRuleResolver interface {
	// VisitRulesFor invokes visitor() with each rule that applies to a given user
	VisitRulesFor(user user.User, visitor VisitorFunc)
}

func (r *RBACAuthorizer) Authorize(ctx context.Context, attributes auth.Attributes) (auth.Decision,
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
