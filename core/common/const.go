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

package common

// const variables
const (
	PageNumber      = "pageNumber"
	PageSize        = "pageSize"
	Filter          = "filter"
	Template        = "template"
	TemplateRelease = "templateRelease"
	TagSelector     = "tagSelector"

	// Offset should be type int
	Offset = "offset"
	// Limit should be type int
	Limit     = "limit"
	StartID   = "startID"
	EndID     = "endID"
	EventType = "eventType"
	WebhookID = "webhookID"
	CreatedAt = "createdAt"
	Enabled   = "enabled"
	Orphaned  = "orphaned"
	OrderBy   = "orderBy"
	ReqID     = "reqID"

	DefaultPageNumber = 1
	DefaultPageSize   = 20
	MaxPageSize       = 50
	MaxItems          = 100000
)

const (
	MetaVersion2              = "0.0.2"
	ApplicationRepoDefaultEnv = "default"
)

const (
	// InternalError internal server error code
	InternalError = "InternalError"

	// InvalidRequestParam invalid request param error code
	InvalidRequestParam = "InvalidRequestParam"

	// InvalidRequestBody invalid request body error code
	InvalidRequestBody = "InvalidRequestBody"

	// NotImplemented not implemented error code
	NotImplemented = "NotImplemented"

	// RequestInfoError error to format the request
	RequestInfoError = "RequestInfoError"

	// Unauthorized  401 error code
	Unauthorized = "Unauthorized"

	// Forbidden 403 Forbidden error code
	Forbidden = "Forbidden"

	// CodeExpired 403 AccessToken and Authorization Token error code
	CodeExpired = "Expired"

	// NotFound 404 NotFound error code
	NotFound = "NotFound"
)

const (
	// ResourceApplication represent the application  member entry.
	ResourceApplication = "applications"

	// ResourceCluster represent the application instance member entry
	ResourceCluster = "clusters"

	ResourceRegion = "regions"

	// ResourceGroup represent the group member entry.
	ResourceGroup = "groups"

	// ResourcePipelinerun currently pipelineruns do not have direct member info, will
	// use the pipeline's cluster's member info
	ResourcePipelinerun = "pipelineruns"

	// ResourceCheckrun currently checkruns do not have direct member info, will
	// use the member info of the clusters that they belong to
	ResourceCheckrun = "checkruns"

	// ResourceOauthApps currently oauthapp do not have direct member info, will
	// use the oauthapp's groups member info
	ResourceOauthApps = "oauthapps"

	ResourceTemplate = "templates"

	ResourceTemplateRelease = "templatereleases"
	AliasTemplateRelease    = "releases"

	ResourceWebhook    = "webhooks"
	ResourceWebhookLog = "webhooklogs"

	ResourceMember = "members"
)

const (
	GroupCore = "core"
)

const (
	ParamApplicationID = "applicationID"
	ParamGroupID       = "groupID"
	ParamClusterID     = "clusterID"
	ParamClusterName   = "cluster"
	ParamTemplateID    = "templateID"
	ParamReleaseID     = "releaseID"
	ParamResourceType  = "resourceType"
	ParamResourceID    = "resourceID"
	ParamAccessTokenID = "accessTokenID"
)

const (
	GitlabGitops   = "gitops"
	GitlabTemplate = "template"
)

const (
	ChartVersionFormat = "%s-%s"
)

const (
	CookieKeyAuth      = "horizon|session"
	SessionKeyAuthUser = "user"
)

const (
	// IDThan query parameter, used for "id > IDThan"
	IDThan = "idThan"
)
