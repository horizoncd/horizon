package group

import (
	"github.com/horizoncd/horizon/pkg/group/models"
)

// convertNewGroupToGroup convert newGroup model to group model.
func convertNewGroupToGroup(newGroup *NewGroup) *models.Group {
	return &models.Group{
		Name:            newGroup.Name,
		Path:            newGroup.Path,
		VisibilityLevel: newGroup.VisibilityLevel,
		Description:     newGroup.Description,
		ParentID:        newGroup.ParentID,
	}
}

// convertUpdateGroupToGroup convert updateGroup model to group model.
func convertUpdateGroupToGroup(updateGroup *UpdateGroup) *models.Group {
	return &models.Group{
		Name:            updateGroup.Name,
		Path:            updateGroup.Path,
		VisibilityLevel: updateGroup.VisibilityLevel,
		Description:     updateGroup.Description,
	}
}
