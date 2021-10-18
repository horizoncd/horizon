package group

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

// SearchParams contains parameters for searching operation
type SearchParams struct {
	Filter     string
	GroupID    uint
	PageNumber int
	PageSize   int
}
