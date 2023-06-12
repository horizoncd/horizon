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

package types

import (
	"testing"

	"github.com/horizoncd/horizon/pkg/auth"
	"github.com/horizoncd/horizon/pkg/authentication/user"
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
			// case list case deny (GVR)
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
