package group

type GroupDetail struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	FullName        string `json:"fullName"`
	Path            string `json:"path"`
	VisibilityLevel string `json:"visibilityLevel"`
	Description     string `json:"description"`
	ParentId        *uint  `json:"parentId"`
}

type NewGroup struct {
	Name            string `json:"name" binding:"required"`
	Path            string `json:"path" binding:"required"`
	VisibilityLevel string `json:"visibilityLevel" binding:"required"`
	Description     string `json:"description"`
	ParentId        *uint  `json:"parentId"`
}
