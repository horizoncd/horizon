package group

import (
	"time"
)

// NewGroup model for creating a group.
type NewGroup struct {
	Name            string `json:"name" binding:"required"`
	Path            string `json:"path" binding:"required"`
	VisibilityLevel string `json:"visibilityLevel"`
	Description     string `json:"description"`
	ParentID        uint   `json:"parentID"`
}

// UpdateGroup model for updating a group.
type UpdateGroup struct {
	Name            string `json:"name" binding:"required"`
	Path            string `json:"path" binding:"required"`
	VisibilityLevel string `json:"visibilityLevel"`
	Description     string `json:"description"`
}

// SearchParams contains parameters for searching operation.
type SearchParams struct {
	Filter     string
	GroupID    uint
	PageNumber int
	PageSize   int
}

// Group basic info of group, including group & application.
type Group struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Path            string    `json:"path"`
	VisibilityLevel string    `json:"visibilityLevel"`
	Description     string    `json:"description"`
	ParentID        uint      `json:"parentID"`
	TraversalIDs    string    `json:"traversalIDs"`
	UpdatedAt       time.Time `json:"updatedAt"`
	FullName        string    `json:"fullName"`
	FullPath        string    `json:"fullPath"`
}

type RegionSelector struct {
	Key      string   `json:"key"`
	Values   []string `json:"values"`
	Operator string   `json:"operator" default:"in"`
}

type RegionSelectors []*RegionSelector

type StructuredGroup struct {
	*Group
	RegionSelectors RegionSelectors `json:"regionSelectors"`
}
