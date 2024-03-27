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

package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/util/sets"
)

type RequestInfoResolver interface {
	NewRequestInfo(req *http.Request) (*RequestInfo, error)
}

type RequestInfo struct {
	// IsResourceRequest indicates whether the request is
	// for an API resource or subresource
	IsResourceRequest bool

	// Path is the URL path of the request
	Path string

	// Verb is the verb associated with the request for API requests.
	// not the http verb.  This includes things like list and watch.
	// for non-resource requests, this is the lowercase http verb
	Verb string

	APIPrefix  string
	APIGroup   string
	APIVersion string

	// Resource is the name of the resource being requested.
	// This is not the kind.  For example: pods
	Resource string
	Name     string

	// Subresource is the name of the subresource being requested.
	Scope       string
	Subresource string
	Parts       []string
}

type RequestInfoFactory struct {
	APIPrefixes sets.String
}

func (r *RequestInfoFactory) NewRequestInfo(ctx context.Context,
	req *http.Request) (*RequestInfo, error) {
	requestInfo := RequestInfo{
		IsResourceRequest: false,
		Path:              req.URL.Path,
		Verb:              strings.ToLower(req.Method),
		Scope:             common.GetScope(ctx, req),
	}

	currentParts := splitPath(req.URL.Path)
	if len(currentParts) < 3 {
		// return a non-resource request
		return &requestInfo, nil
	}

	if !r.APIPrefixes.Has(currentParts[0]) {
		// return a non-resource request
		return &requestInfo, nil
	}

	requestInfo.APIPrefix = currentParts[0]
	currentParts = currentParts[1:]

	requestInfo.APIGroup = currentParts[0]
	currentParts = currentParts[1:]

	requestInfo.APIVersion = currentParts[0]
	requestInfo.IsResourceRequest = true
	currentParts = currentParts[1:]

	switch req.Method {
	case "POST":
		requestInfo.Verb = "create"
	case "GET", "HEAD":
		// nolint
		requestInfo.Verb = "get"
	case "PUT":
		requestInfo.Verb = "update"
	case "PATCH":
		requestInfo.Verb = "patch"
	case "DELETE":
		requestInfo.Verb = "delete"
	default:
		requestInfo.Verb = ""
	}

	requestInfo.Parts = currentParts

	// parts  resource/resourceName/subresource
	switch {
	case len(requestInfo.Parts) >= 3:
		requestInfo.Subresource = requestInfo.Parts[2]
		fallthrough
	case len(requestInfo.Parts) >= 2:
		requestInfo.Name = requestInfo.Parts[1]
		fallthrough
	case len(requestInfo.Parts) >= 1:
		requestInfo.Resource = requestInfo.Parts[0]
	}

	if len(requestInfo.Name) == 0 && requestInfo.Verb == "get" {
		requestInfo.Verb = "list"
	}

	return &requestInfo, nil
}

// splitPath returns the segments for a URL path.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}
