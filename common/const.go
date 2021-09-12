package common

import "errors"

// const variables
const (
	PageNumber = "pageNumber"
	PageSize   = "pageSize"

	DefaultPageNumber = 1
	DefaultPageSize   = 20
)

var (
	NameConflictErr = errors.New("name conflict")
)
