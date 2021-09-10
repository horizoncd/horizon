package group

type GroupDetail struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Path            string `json:"path"`
	VisibilityLevel string `json:"visibilityLevel"`
	Description     string `json:"description"`
	ParentId        *int   `json:"parentId"`
}

type NewGroup struct {
	Name            string `json:"name" binding:"required"`
	Path            string `json:"path" binding:"required"`
	VisibilityLevel string `json:"visibilityLevel" binding:"required"`
	Description     string `json:"description"`
	ParentId        *int   `json:"parentId"`
}
