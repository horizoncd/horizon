package group

import (
	"fmt"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group"
	"g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
)

const (
	ErrCodeNotFound = "GroupNotFound"

	ParamGroupID  = "groupID"
	ParamPath     = "path"
	ParamFilter   = "filter"
	QueryParentID = "parentId"
	ParentID      = "parent_id"
	Group         = "group"
)

type Controller struct {
	groupManager group.Manager
}

func NewController() *Controller {
	return &Controller{
		groupManager: group.Mgr,
	}
}

func (controller *Controller) CreateGroup(c *gin.Context) {
	var newGroup *NewGroup
	err := c.ShouldBindJSON(&newGroup)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("create group failed: %v", err))
		return
	}

	create, err := controller.groupManager.Create(c, convertNewGroupToGroup(newGroup))
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("create group failed: %v", err))
		return
	}

	response.SuccessWithData(c, create)
}

func (controller *Controller) DeleteGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)

	idInt, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("delete group failed: %v", err))
		return
	}
	err = controller.groupManager.Delete(c, uint(idInt))
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("delete group failed: %v", err))
		return
	}
	response.Success(c)
}

func (controller *Controller) GetGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)

	idInt, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("get group failed: %v", err))
		return
	}

	groupEntry, err := controller.groupManager.GetByID(c, uint(idInt))
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("get group failed: %v", err))
		return
	}

	if groupEntry == nil {
		response.AbortWithRequestError(c, ErrCodeNotFound, "group not found")
		return
	}

	detail := ConvertGroupToGroupDetail(groupEntry)
	fullPath, fullName, err := formatFullPathAndFullName(c, groupEntry)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("get group failed: %v", err))
		return
	}
	detail.FullPath = fullPath
	detail.FullName = fullName

	response.SuccessWithData(c, detail)
}

func (controller *Controller) GetGroupByPath(c *gin.Context) {
	path := c.Query(ParamPath)

	paths := strings.Split(c.Query(ParamPath)[1:], "/")
	groups, err := controller.groupManager.GetByPaths(c, paths)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("get group by path failed: %v", err))
		return
	}

	idToGroup := make(map[uint]*models.Group)
	fullPathToGroup := make(map[string]*models.Group)
	idToFullName := make(map[uint]string)
	for _, m := range groups {
		idToGroup[m.ID] = m
		tIDs := strings.Split(m.TraversalIDs, ",")
		paths := make([]string, len(tIDs))
		names := make([]string, len(tIDs))
		for i, d := range tIDs {
			ind, _ := strconv.Atoi(d)
			if _, ok := idToGroup[uint(ind)]; !ok {
				response.AbortWithInternalError(c, "get group by path failed")
				return
			}
			paths[i] = idToGroup[uint(ind)].Path
			names[i] = idToGroup[uint(ind)].Name
		}
		fullPath := "/" + strings.Join(paths, "/")
		fullPathToGroup[fullPath] = m
		fullName := strings.Join(names, " / ")
		idToFullName[m.ID] = fullName
	}

	// path pointing to a group
	if groupEntity, ok := fullPathToGroup[path]; ok {
		detail := ConvertGroupToGroupDetail(groupEntity)
		detail.Type = Group
		detail.FullPath = path
		detail.FullName = idToFullName[groupEntity.ID]

		response.SuccessWithData(c, detail)
	}
}

// TODO(wurongjun) support transfer group
// func (controller *Controller) TransferGroup(c *gin.Context)

// TODO(wurongjun) change to UpdateGroupBasic (also change the openapi)
func (controller *Controller) UpdateGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)

	idInt, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("upate group failed: %v", err))
		return
	}

	var updatedGroup *UpdateGroup
	err = c.ShouldBindJSON(&updatedGroup)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("upate group failed: %v", err))
		return
	}

	groupEntry := convertUpdateGroupToGroup(updatedGroup)
	groupEntry.ID = uint(idInt)
	err = controller.groupManager.UpdateBasic(c, groupEntry)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("upate group failed: %v", err))
		return
	}

	response.Success(c)
}

func (controller *Controller) GetChildren(c *gin.Context) {
	// todo also query application
	controller.GetSubGroups(c)
}

func (controller *Controller) GetSubGroups(c *gin.Context) {
	parentID := c.Param(ParamGroupID)
	atoi, err := strconv.Atoi(parentID)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("get subgroups failed: %v", err))
		return
	}
	pGroup, err := controller.groupManager.GetByID(c, uint(atoi))
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("get subgroups failed: %v", err))
		return
	}

	groups, count, err := controller.groupManager.List(c, formatQuerySubGroups(c))
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("get subgroups failed: %v", err))
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: count,
		Items: controller.formatGroupDetails(c, pGroup, groups),
	})
}

func (controller *Controller) SearchChildren(c *gin.Context) {
	// TODO(wurongjun): also query application
	controller.SearchGroups(c)
}

func (controller *Controller) SearchGroups(c *gin.Context) {
	filter := c.Query(ParamFilter)

	// filter is empty, just list the group
	if filter == "" {
		parentID := c.Query(QueryParentID)
		var pGroup *models.Group
		if parentID != "" {
			atoi, err := strconv.Atoi(parentID)
			if err != nil {
				response.AbortWithRequestError(c, common.InvalidRequestParam,
					fmt.Sprintf("get subgroups failed: %v", err))
				return
			}
			pGroup, err = controller.groupManager.GetByID(c, uint(atoi))
			if err != nil {
				response.AbortWithInternalError(c, fmt.Sprintf("get subgroups failed: %v", err))
				return
			}
		}

		groups, count, err := controller.groupManager.List(c, formatSearchGroups(c))
		if err != nil {
			response.AbortWithInternalError(c, fmt.Sprintf("search groups failed: %v", err))
			return
		}
		response.SuccessWithData(c, response.DataWithTotal{
			Total: count,
			Items: controller.formatGroupDetails(c, pGroup, groups),
		})
		return
	}

	queryGroups, err := controller.groupManager.GetByNameFuzzily(c, filter)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("search groups failed: %v", err))
		return
	}

	var ids []uint
	for _, g := range queryGroups {
		split := strings.Split(g.TraversalIDs, ",")
		for _, s := range split {
			i, _ := strconv.Atoi(s)
			ids = append(ids, uint(i))
		}
	}

	groups, err := controller.groupManager.GetByIDsOrderByIDDesc(c, ids)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("search groups failed: %v", err))
		return
	}

	// organize struct of search result
	parentIDToGroupsMap := make(map[int][]*Child)
	// group in the first level, must return in search
	firstLevelGroupsDetails := make([]*Child, 0)
	idsToFullPath := make(map[string]string)
	idsToFullName := make(map[string]string)
	for _, g := range groups {
		detail := ConvertGroupToGroupDetail(g)
		if g.ParentID == -1 {
			detail.FullPath = "/" + g.Path
			detail.FullName = g.Name
		} else {
			prefixIds := g.TraversalIDs[:len(g.TraversalIDs)-2]
			detail.FullPath = idsToFullPath[prefixIds] + "/" + g.Path
			detail.FullName = idsToFullName[prefixIds] + " / " + g.Name
		}
		idsToFullPath[g.TraversalIDs] = detail.FullPath
		idsToFullName[g.TraversalIDs] = detail.FullName

		// current only query group table
		detail.Type = Group
		// group in the first level
		if g.ParentID == -1 {
			firstLevelGroupsDetails = append(firstLevelGroupsDetails, detail)
		}

		parentID := g.ParentID
		// name match or children's names match
		if strings.Contains(g.Name, filter) || len(parentIDToGroupsMap[int(g.ID)]) > 0 {
			parentIDToGroupsMap[parentID] = append(parentIDToGroupsMap[parentID], detail)
		}

		if v, ok := parentIDToGroupsMap[int(detail.ID)]; ok {
			detail.ChildrenCount = len(v)
			detail.Children = v
		}
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(len(firstLevelGroupsDetails)),
		Items: firstLevelGroupsDetails,
	})
}

func (controller *Controller) formatGroupDetails(c *gin.Context,
	pGroup *models.Group, groups []*models.Group) []*Child {
	var parentIds []uint
	for _, m := range groups {
		parentIds = append(parentIds, m.ID)
	}
	query := q.New(q.KeyWords{
		ParentID: parentIds,
	})
	subGroups, err := controller.groupManager.ListWithoutPage(c, query)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("get subgroups failed: %v", err))
		return nil
	}
	childrenCountMap := map[int]int{}
	// update childrenCount field
	for _, subgroup := range subGroups {
		if v, ok := childrenCountMap[subgroup.ParentID]; ok {
			childrenCountMap[subgroup.ParentID] = v + 1
		} else {
			childrenCountMap[subgroup.ParentID] = 1
		}
	}

	fullPath := ""
	fullName := ""
	if pGroup != nil {
		fullPath, fullName, _ = formatFullPathAndFullName(c, pGroup)
	}
	var details = make([]*Child, len(groups))
	for idx, tmp := range groups {
		detail := ConvertGroupToGroupDetail(tmp)
		// todo currently using fixed type: group
		detail.Type = Group
		detail.ChildrenCount = childrenCountMap[int(detail.ID)]
		detail.FullPath = fullPath + "/" + detail.Path
		if fullName == "" {
			detail.FullName = detail.Name
		} else {
			detail.FullName = fullName + " / " + detail.Name
		}
		details[idx] = detail
	}

	return details
}

// url pattern: api/vi/groups/:groupId/subgroups
func formatQuerySubGroups(c *gin.Context) *q.Query {
	parentID := c.Param(ParamGroupID)
	k := q.KeyWords{
		ParentID: -1,
	}
	if parentID != "" {
		k[ParentID], _ = strconv.Atoi(parentID)
	}

	query := formatDefaultQuery(c)
	query.Keywords = k

	return query
}

// url pattern: api/vi/groups/search?parentId=?
func formatSearchGroups(c *gin.Context) *q.Query {
	parentID := c.Query(QueryParentID)
	k := q.KeyWords{
		ParentID: -1,
	}
	if parentID != "" {
		k[ParentID], _ = strconv.Atoi(parentID)
	}

	query := formatDefaultQuery(c)
	query.Keywords = k

	return query
}

func formatDefaultQuery(c *gin.Context) *q.Query {
	query := q.New(q.KeyWords{})
	query.PageNumber = common.DefaultPageNumber
	query.PageSize = common.DefaultPageSize
	pageNumber, _ := strconv.Atoi(c.Query(common.PageNumber))
	pageSize, _ := strconv.Atoi(c.Query(common.PageSize))
	if pageNumber > 0 {
		query.PageNumber = pageNumber
	}
	if pageSize > 0 {
		query.PageSize = pageSize
	}
	// sort by updated_at desc default，let newer items be in head
	s := q.NewSort("updated_at", true)
	query.Sorts = []*q.Sort{s}

	return query
}
