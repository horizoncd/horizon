package service

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

// Full is the fullName&fullPath for a group/application/applicationInstance
type Full struct {
	FullName string `json:"fullName"`
	FullPath string `json:"fullPath"`
}
