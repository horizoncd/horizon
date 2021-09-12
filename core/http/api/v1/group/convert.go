package group

import "g.hz.netease.com/horizon/pkg/group/models"

func ConvertGroupToGroupDetail(group *models.Group) *GroupDetail {
	return &GroupDetail{
		ID:              group.ID,
		Name:            group.Name,
		FullName:        group.FullName,
		Path:            group.Path,
		VisibilityLevel: group.VisibilityLevel,
		Description:     group.Description,
		ParentId:        group.ParentId,
	}
}

func convertNewGroupToGroup(newGroup *NewGroup) *models.Group {
	return &models.Group{
		Name:            newGroup.Name,
		Path:            newGroup.Path,
		VisibilityLevel: newGroup.VisibilityLevel,
		Description:     newGroup.Description,
		ParentId:        newGroup.ParentId,
	}
}
