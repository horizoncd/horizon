package common

import "errors"

// const variables
const (
	PageNumber      = "pageNumber"
	PageSize        = "pageSize"
	Filter          = "filter"
	Template        = "template"
	TemplateRelease = "template_release"

	DefaultPageNumber = 1
	DefaultPageSize   = 20
	MaxPageSize       = 50
)

var (
	FilterKeywords = map[string]struct{}{
		"template":         {},
		"template_release": {},
	}
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

	// Forbidden 403 Forbidden error code
	Forbidden = "Forbidden"

	// NotFound 404 NotFound error code
	NotFound = "NotFound"
)

var (
	ErrNameConflict      = errors.New("name conflict")
	ErrParameterNotValid = errors.New("parameter not valid")
)
