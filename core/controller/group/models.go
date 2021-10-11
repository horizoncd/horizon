package group

import "time"

// Child subResource of a group, including group & application
type Child struct {
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
	Type            string    `json:"type"`
	ChildrenCount   int       `json:"childrenCount"`
	Children        []*Child  `json:"children"`
}

// NewGroup model for creating a group
type NewGroup struct {
	Name            string `json:"name" binding:"required"`
	Path            string `json:"path" binding:"required"`
	VisibilityLevel string `json:"visibilityLevel" binding:"required"`
	Description     string `json:"description"`
	ParentID        uint   `json:"parentID"`
}

// UpdateGroup model for updating a group
type UpdateGroup struct {
	Name            string `json:"name" binding:"required"`
	Path            string `json:"path" binding:"required"`
	VisibilityLevel string `json:"visibilityLevel" binding:"required"`
	Description     string `json:"description"`
}

// Full is the fullName&fullPath for a group/application/applicationInstance
type Full struct {
	FullName string `json:"fullName"`
	FullPath string `json:"fullPath"`
}

// SearchParams contains parameters for searching operation
type SearchParams struct {
	Filter     string
	GroupID    uint
	PageNumber int
	PageSize   int
}
