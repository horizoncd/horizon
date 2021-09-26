package common

import "errors"

// const variables
const (
	PageNumber = "pageNumber"
	PageSize   = "pageSize"

	DefaultPageNumber = 1
	DefaultPageSize   = 20

	RootGroupID = 0
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
)

var (
	ErrNameConflict      = errors.New("name conflict")
	ErrParameterNotValid = errors.New("parameter not valid")
)
