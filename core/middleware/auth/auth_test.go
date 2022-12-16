package auth

import (
	"net/http"
	"testing"

	"github.com/horizoncd/horizon/pkg/auth"
	"github.com/horizoncd/horizon/pkg/util/sets"

	"github.com/stretchr/testify/assert"
)

func TestRequestInfo(t *testing.T) {
	requestInfoFactory := auth.RequestInfoFactory{
		APIPrefixes: sets.NewString("apis", "api"),
	}

	caseTable := []struct {
		method            string
		path              string
		expectRequestInfo auth.RequestInfo
	}{
		{
			// get subresource
			"GET",
			"/apis/core/v1/groups/1/member",
			auth.RequestInfo{
				IsResourceRequest: true,
				Path:              "/apis/core/v1/groups/1/member",
				APIPrefix:         "apis",
				APIGroup:          "core",
				APIVersion:        "v1",
				Verb:              "get",
				Resource:          "groups",
				Name:              "1",
				Subresource:       "member",
				Parts:             []string{"groups", "1", "member"},
			},
		},
		{
			// non-resource request
			method: "GET",
			path:   "/ao/dsds?scope=pre",
			expectRequestInfo: auth.RequestInfo{
				IsResourceRequest: false,
				Path:              "/ao/dsds",
				Verb:              "get",
				Scope:             "pre",
			},
		},
		{
			// list group
			method: "GET",
			path:   "/apis/core/v1/groups?path=path",
			expectRequestInfo: auth.RequestInfo{
				IsResourceRequest: true,
				Path:              "/apis/core/v1/groups",
				APIPrefix:         "apis",
				APIGroup:          "core",
				APIVersion:        "v1",
				Verb:              "list",
				Resource:          "groups",
				Parts:             []string{"groups"},
			},
		},
		{
			// post resource
			method: "POST",
			path:   "/apis/core/v1/groups?scope=production",
			expectRequestInfo: auth.RequestInfo{
				IsResourceRequest: true,
				Path:              "/apis/core/v1/groups",
				APIPrefix:         "apis",
				APIGroup:          "core",
				APIVersion:        "v1",
				Verb:              "create",
				Resource:          "groups",
				Scope:             "production",
				Parts:             []string{"groups"},
			},
		},
		{
			// get resource
			method: "GET",
			path:   "/apis/core/v1/groups/1",
			expectRequestInfo: auth.RequestInfo{
				IsResourceRequest: true,
				Path:              "/apis/core/v1/groups/1",
				APIPrefix:         "apis",
				APIGroup:          "core",
				APIVersion:        "v1",
				Verb:              "get",
				Resource:          "groups",
				Name:              "1",
				Parts:             []string{"groups", "1"},
			},
		},
		{
			// get resource
			method: "PUT",
			path:   "/apis/core/v1/groups/1",
			expectRequestInfo: auth.RequestInfo{
				IsResourceRequest: true,
				Path:              "/apis/core/v1/groups/1",
				APIPrefix:         "apis",
				APIGroup:          "core",
				APIVersion:        "v1",
				Verb:              "update",
				Resource:          "groups",
				Name:              "1",
				Parts:             []string{"groups", "1"},
			},
		},
		{
			// get resource
			method: "PATCH",
			path:   "/apis/core/v1/groups/1",
			expectRequestInfo: auth.RequestInfo{
				IsResourceRequest: true,
				Path:              "/apis/core/v1/groups/1",
				APIPrefix:         "apis",
				APIGroup:          "core",
				APIVersion:        "v1",
				Verb:              "patch",
				Resource:          "groups",
				Name:              "1",
				Parts:             []string{"groups", "1"},
			},
		},
		{
			// get resource
			method: "DELETE",
			path:   "/apis/core/v1/groups/1",
			expectRequestInfo: auth.RequestInfo{
				IsResourceRequest: true,
				Path:              "/apis/core/v1/groups/1",
				APIPrefix:         "apis",
				APIGroup:          "core",
				APIVersion:        "v1",
				Verb:              "delete",
				Resource:          "groups",
				Name:              "1",
				Parts:             []string{"groups", "1"},
			},
		},
	}

	for _, v := range caseTable {
		assert.Equal(t, &v.expectRequestInfo, func(method, url string) *auth.RequestInfo {
			req, _ := http.NewRequest(method, url, nil)
			requestInfo, _ := requestInfoFactory.NewRequestInfo(req)
			return requestInfo
		}(v.method, v.path))
	}
}
