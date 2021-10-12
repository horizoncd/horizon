package group

import (
	"strconv"
	"strings"

	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/group/models"
)

// generateFullFromGroups generate fullPath which looks like /a/b/c, and fullName which looks like 1 / 2
func generateFullFromGroups(groups []*models.Group) *Full {
	var fullPath, fullName string
	paths := make([]string, len(groups))
	names := make([]string, len(groups))
	for i, model := range groups {
		paths[i] = model.Path
		names[i] = model.Name
	}

	fullPath = "/" + strings.Join(paths, "/")
	fullName = strings.Join(names, " / ")

	return &Full{
		FullName: fullName,
		FullPath: fullPath,
	}
}

// convertGroupToChild format Child based on group model、fullName、fullPath and resourceType
func convertGroupToChild(group *models.Group, full *Full) *Child {
	return &Child{
		ID:              group.ID,
		Name:            group.Name,
		Path:            group.Path,
		VisibilityLevel: group.VisibilityLevel,
		Description:     group.Description,
		ParentID:        group.ParentID,
		TraversalIDs:    group.TraversalIDs,
		UpdatedAt:       group.UpdatedAt,
		FullName:        full.FullName,
		FullPath:        full.FullPath,
		Type:            ChildTypeGroup,
	}
}

func convertApplicationToChild(app *appmodels.Application, full *Full) (*Child, error) {
	return &Child{
		ID:          app.ID,
		Name:        app.Name,
		Path:        app.Name,
		Description: app.Description,
		ParentID:    app.GroupID,
		FullName:    full.FullName,
		FullPath:    full.FullPath,
		Type:        ChildTypeApplication,
	}, nil
}

// convertNewGroupToGroup convert newGroup model to group model
func convertNewGroupToGroup(newGroup *NewGroup) *models.Group {
	return &models.Group{
		Name:            newGroup.Name,
		Path:            newGroup.Path,
		VisibilityLevel: newGroup.VisibilityLevel,
		Description:     newGroup.Description,
		ParentID:        newGroup.ParentID,
	}
}

// convertUpdateGroupToGroup convert updateGroup model to group model
func convertUpdateGroupToGroup(updateGroup *UpdateGroup) *models.Group {
	return &models.Group{
		Name:            updateGroup.Name,
		Path:            updateGroup.Path,
		VisibilityLevel: updateGroup.VisibilityLevel,
		Description:     updateGroup.Description,
	}
}

/*
assuming we have 3 groups,
group one: {id: 1, name: "a", path: "w"}
group two: {id: 2, name: "b", path: "r"}
group three: {id: 3, name: "c", path: "j"}

after the function executed, we get a map:
{
  1: {
    fullName: "a",
    fullPath: "/w"
  },
  2: {
    fullName: "a / b",
    fullPath: "/w/r"
  },
  3: {
    fullName: "a / b / c",
    fullPath: "/w/r/j"
  },
}
*/
func generateIDToFull(groups []*models.Group) map[uint]*Full {
	idToFull := make(map[uint]*Full)

	for _, g := range groups {
		var fullName, fullPath string
		if g.ParentID == RootGroupID {
			fullName = g.Name
			fullPath = "/" + g.Path
		} else {
			parentFull := idToFull[g.ParentID]
			fullName = parentFull.FullName + " / " + g.Name
			fullPath = parentFull.FullPath + "/" + g.Path
		}

		idToFull[g.ID] = &Full{
			FullPath: fullPath,
			FullName: fullName,
		}
	}

	return idToFull
}

// generateIDToGroup map id to group
func generateIDToGroup(groups []*models.Group) map[uint]*models.Group {
	idToGroup := make(map[uint]*models.Group)

	for _, g := range groups {
		idToGroup[g.ID] = g
	}

	return idToGroup
}

// formatIDsFromTraversalIDs format id array from traversalIDs(1,2,3)
func formatIDsFromTraversalIDs(traversalIDs string) []uint {
	splitIds := strings.Split(traversalIDs, ",")
	var ids = make([]uint, len(splitIds))
	for i, id := range splitIds {
		ii, _ := strconv.Atoi(id)
		ids[i] = uint(ii)
	}
	return ids
}
