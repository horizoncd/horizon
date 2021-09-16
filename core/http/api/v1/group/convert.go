package group

import "g.hz.netease.com/horizon/pkg/group/models"

func ConvertGroupToGroupDetail(group *models.Group) *Child {
	return &Child{
		ID:              group.ID,
		Name:            group.Name,
		FullName:        group.FullName,
		Path:            group.Path,
		VisibilityLevel: group.VisibilityLevel,
		Description:     group.Description,
		ParentID:        group.ParentID,
	}
}

func convertNewGroupToGroup(newGroup *NewGroup) *models.Group {
	return &models.Group{
		Name:            newGroup.Name,
		Path:            newGroup.Path,
		VisibilityLevel: newGroup.VisibilityLevel,
		Description:     newGroup.Description,
		ParentID:        newGroup.ParentID,
	}
}

func convertUpdateGroupToGroup(updateGroup *UpdateGroup) *models.Group {
	return &models.Group{
		Name:            updateGroup.Name,
		VisibilityLevel: updateGroup.VisibilityLevel,
		Description:     updateGroup.Description,
	}
}
