package common

// const variables
const (
	PageNumber      = "pageNumber"
	PageSize        = "pageSize"
	Filter          = "filter"
	Template        = "template"
	TemplateRelease = "templateRelease"
	TagSelector     = "tagSelector"

	Offset    = "offset"
	Limit     = "limit"
	StartID   = "startID"
	EndID     = "endID"
	EventType = "eventType"
	CreatedAt = "createdAt"
	Enabled   = "enabled"

	DefaultPageNumber = 1
	DefaultPageSize   = 20
	MaxPageSize       = 50
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

	// ResourceOauthApps currently oauthapp do not have direct member info, will
	// use the oauthapp's groups member info
	ResourceOauthApps = "oauthapps"

	ResourceTemplate = "templates"

	ResourceTemplateRelease = "templatereleases"
	AliasTemplateRelease    = "releases"

	ResourceWebhook    = "webhooks"
	ResourceWebhookLog = "webhooklogs"
)

const (
	GroupCore  = "core"
	GroupFront = "front"
)

const (
	ParamApplicationID = "applicationID"
	ParamGroupID       = "groupID"
	ParamClusterID     = "clusterID"
	ParamClusterName   = "cluster"
	ParamTemplateID    = "templateID"
	ParamReleaseID     = "releaseID"
	ParamPipelinerunID = "pipelinerunID"
	ParamResourceType  = "resourceType"
	ParamResourceID    = "resourceID"
	ParamAccessTokenID = "accessTokenID"
)

const (
	GitlabCode     = "code"
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
