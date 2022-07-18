package common

// const variables
const (
	PageNumber      = "pageNumber"
	PageSize        = "pageSize"
	Filter          = "filter"
	Template        = "template"
	TemplateRelease = "templateRelease"
	TagSelector     = "tagSelector"

	DefaultPageNumber = 1
	DefaultPageSize   = 20
	MaxPageSize       = 50
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

	// ResourceOauthApps urrently oauthapp do not have direct member info, will
	// use the oauthapp's groups member info
	ResourceOauthApps = "oauthapps"
)

const (
	GroupCore  = "core"
	GroupFront = "front"
)
