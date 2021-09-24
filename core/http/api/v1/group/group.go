package group

type Child struct {
	ID              uint     `json:"id"`
	Name            string   `json:"name"`
	Path            string   `json:"path"`
	FullName        string   `json:"fullName"`
	FullPath        string   `json:"fullPath"`
	VisibilityLevel string   `json:"visibilityLevel"`
	Description     string   `json:"description"`
	ParentID        int      `json:"parentId"`
	Type            string   `json:"type"`
	ChildrenCount   int      `json:"childrenCount"`
	Children        []*Child `json:"children"`
}

type NewGroup struct {
	Name            string `json:"name" binding:"required"`
	Path            string `json:"path" binding:"required"`
	VisibilityLevel string `json:"visibilityLevel" binding:"required"`
	Description     string `json:"description"`
	ParentID        int    `json:"parentId"`
}

type UpdateGroup struct {
	Name            string `json:"name" binding:"required"`
	VisibilityLevel string `json:"visibilityLevel" binding:"required"`
	Description     string `json:"description"`
}
