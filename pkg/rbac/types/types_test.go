package types

import (
	"testing"

	"g.hz.netease.com/horizon/pkg/auth"
	"g.hz.netease.com/horizon/pkg/authentication/user"
	"github.com/stretchr/testify/assert"
)

func TestRuleAllow(t *testing.T) {
	testUser := user.DefaultInfo{
		Name:     "tom",
		FullName: "hzsunjianliang",
		ID:       12345,
	}

	caseTable := []struct {
		attr    auth.Attributes
		policy  PolicyRule
		allowed bool
	}{
		{
			// wizard case pass
			attr: auth.AttributesRecord{
				User:            &testUser,
				Verb:            "get",
				APIGroup:        "core",
				APIVersion:      "v1",
				Resource:        "*",
				SubResource:     "",
				Scope:           "online/hz",
				ResourceRequest: true,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:     []string{"*"},
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Scopes:    []string{"*"},
			},
			allowed: true,
		}, {
			// case list case pass
			attr: auth.AttributesRecord{
				User:            &testUser,
				Verb:            "get",
				APIGroup:        "core",
				APIVersion:      "v1",
				Resource:        "group",
				SubResource:     "",
				Scope:           "online/hz",
				ResourceRequest: true,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:     []string{"get", "create", "update", "patch", "delete"},
				APIGroups: []string{"core", "rest"},
				Resources: []string{"group", "application"},
				Scopes:    []string{"online/*", "test/*"},
			},
			allowed: true,
		}, {
			// case list case deny (Verb)
			attr: auth.AttributesRecord{
				User:            &testUser,
				Verb:            "delete",
				APIGroup:        "core",
				APIVersion:      "v1",
				Resource:        "group",
				SubResource:     "",
				Scope:           "online/hz",
				ResourceRequest: true,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:     []string{"get", "create", "update", "patch"},
				APIGroups: []string{"core", "rest"},
				Resources: []string{"group", "application"},
				Scopes:    []string{"online/*", "test/*"},
			},
			allowed: false,
		}, {
			// case list case deny (API group)
			attr: auth.AttributesRecord{
				User:            &testUser,
				Verb:            "delete",
				APIGroup:        "core",
				APIVersion:      "v1",
				Resource:        "group",
				SubResource:     "",
				Scope:           "online/hz",
				ResourceRequest: true,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:     []string{"get", "create", "update", "delete"},
				APIGroups: []string{"rest"},
				Resources: []string{"group", "application"},
				Scopes:    []string{"online/*", "test/*"},
			},
			allowed: false,
		}, {
			// case list case deny (Resource)
			attr: auth.AttributesRecord{
				User:            &testUser,
				Verb:            "delete",
				APIGroup:        "core",
				APIVersion:      "v1",
				Resource:        "group",
				SubResource:     "",
				Scope:           "online/hz",
				ResourceRequest: true,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:     []string{"get", "create", "update", "patch", "delete"},
				APIGroups: []string{"core"},
				Resources: []string{"application"},
				Scopes:    []string{"online/*", "test/*"},
			},
			allowed: false,
		}, {
			// wizard case allow (subResource)
			attr: auth.AttributesRecord{
				User:            &testUser,
				Verb:            "delete",
				APIGroup:        "core",
				APIVersion:      "v1",
				Resource:        "application",
				SubResource:     "member",
				Scope:           "online/hz",
				ResourceRequest: true,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:     []string{"get", "create", "update", "patch", "delete"},
				APIGroups: []string{"core"},
				Resources: []string{"application/*"},
				Scopes:    []string{"online/*", "test/*"},
			},
			allowed: false,
		}, {
			// case  case allow (subResource)
			attr: auth.AttributesRecord{
				User:            &testUser,
				Verb:            "delete",
				APIGroup:        "core",
				APIVersion:      "v1",
				Resource:        "application",
				SubResource:     "member",
				Scope:           "online/hz",
				ResourceRequest: true,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:     []string{"get", "create", "update", "patch", "delete"},
				APIGroups: []string{"core"},
				Resources: []string{"application/member"},
				Scopes:    []string{"*"},
			},
			allowed: true,
		}, {
			// case  case deny (subResource)
			attr: auth.AttributesRecord{
				User:            &testUser,
				Verb:            "delete",
				APIGroup:        "core",
				APIVersion:      "v1",
				Resource:        "application",
				SubResource:     "member",
				Scope:           "online/hz",
				ResourceRequest: true,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:     []string{"get", "create", "update", "patch", "delete"},
				APIGroups: []string{"core"},
				Resources: []string{"group/*"},
				Scopes:    []string{"*"},
			},
			allowed: false,
		}, {
			// wizard case deny (scope)
			attr: auth.AttributesRecord{
				User:            &testUser,
				Verb:            "delete",
				APIGroup:        "core",
				APIVersion:      "v1",
				Resource:        "application",
				SubResource:     "",
				Scope:           "online/hz",
				ResourceRequest: true,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:     []string{"get", "create", "update", "patch", "delete"},
				APIGroups: []string{"core"},
				Resources: []string{"application"},
				Scopes:    []string{"test/*"},
			},
			allowed: false,
		}, {
			// wizard case allow (scope)
			attr: auth.AttributesRecord{
				User:            &testUser,
				Verb:            "delete",
				APIGroup:        "core",
				APIVersion:      "v1",
				Resource:        "application",
				SubResource:     "",
				Scope:           "test/hz",
				ResourceRequest: true,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:     []string{"get", "create", "update", "patch", "delete"},
				APIGroups: []string{"core"},
				Resources: []string{"application"},
				Scopes:    []string{"test/*"},
			},
			allowed: true,
		}, {
			// wizard case allow (nonResourceURLs)
			attr: auth.AttributesRecord{
				Verb:            "get",
				ResourceRequest: false,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:           []string{"get"},
				APIGroups:       []string{"core"},
				Resources:       []string{"application"},
				NonResourceURLs: []string{"/*"},
			},
			allowed: true,
		}, {
			// wizard case allow (nonResourceURLs)
			attr: auth.AttributesRecord{
				Verb:            "get",
				ResourceRequest: false,
				Path:            "/apis/core/v1/group/1",
			},
			policy: PolicyRule{
				Verbs:           []string{"get"},
				APIGroups:       []string{"core"},
				Resources:       []string{"application"},
				NonResourceURLs: []string{"/apis/front/*"},
			},
			allowed: false,
		},
	}

	for _, v := range caseTable {
		assert.Equal(t, RuleAllow(v.attr, &v.policy), v.allowed)
	}
}
