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

// formatFull {1,2,3} -> /a/b/c (fullPath) & w / r / j (fullName)
func formatFull(groups []*models.Group) (map[string]*Full, map[string]*models.Group) {
	traversalIDsToFull := make(map[string]*Full)
	traversalIDsToGroup := make(map[string]*models.Group)
	for _, m := range groups {
		traversalIDsToGroup[m.TraversalIDs] = m
		if m.ParentID == common.RootGroupID {
			traversalIDsToFull[m.TraversalIDs] = &Full{
				FullName: m.Name,
				FullPath: "/" + m.Path,
			}
		} else {
			split := strings.Split(m.TraversalIDs, ",")
			prefixIds := strings.Join(split[:len(split)-1], ",")
			traversalIDsToFull[m.TraversalIDs] = &Full{
				FullName: traversalIDsToFull[prefixIds].FullName + " / " + m.Name,
				FullPath: traversalIDsToFull[prefixIds].FullPath + "/" + m.Path,
			}
		}
	}

	return traversalIDsToFull, traversalIDsToGroup
}

func (controller *Controller) GetGroupByPath(c *gin.Context) {
	path := c.Query(ParamPath)

	paths := strings.Split(c.Query(ParamPath)[1:], "/")
	groups, err := controller.groupManager.GetByPaths(c, paths)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("get group by path failed: %v", err))
		return
	}

	// {1,2,3} -> /a/b/c (fullPath) & w / r / j (fullName)
	traversalIDsToFull, traversalIDsToGroup := formatFull(groups)

	for k, v := range traversalIDsToFull {
		// path pointing to a group
		if v.FullPath == path {
			m := traversalIDsToGroup[k]
			detail := ConvertGroupToGroupDetail(m)
			detail.Type = Group
			detail.FullPath = v.FullPath
			detail.FullName = v.FullName

			response.SuccessWithData(c, detail)
		}
	}
}

// TODO(wurongjun) support transfer group
func (controller *Controller) TransferGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)
	parentID := c.Query(QueryParentID)
	gInt, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("transfer group failed: %v", err))
		return
	}
	pgInt, err := strconv.ParseUint(parentID, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("transfer group failed: %v", err))
		return
	}

	err = controller.groupManager.Transfer(c, uint(gInt), uint(pgInt))
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("transfer subgroups failed: %v", err))
		return
	}

	response.Success(c)
}

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
	parentID := c.Query(QueryParentID)
	var pGroup *models.Group
	pGroupID := common.RootGroupID
	if parentID != "" {
		a, err := strconv.Atoi(parentID)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam,
				fmt.Sprintf("get subgroups failed: %v", err))
			return
		}
		pGroupID = a
		if a > 0 {
			pGroup, err = controller.groupManager.GetByID(c, uint(a))
			if err != nil {
				response.AbortWithInternalError(c, fmt.Sprintf("get subgroups failed: %v", err))
				return
			}
		}
	}

	filter := c.Query(ParamFilter)
	// filter is empty, just list the group
	if filter == "" {
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

	groups, err := controller.groupManager.GetByIDs(c, ids)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("search groups failed: %v", err))
		return
	}

	// {1,2,3} -> /a/b/c (fullPath) & w / r / j (fullName)
	traversalIDsToFull, _ := formatFull(groups)
	// organize struct of search result
	parentIDToGroupsMap := make(map[int][]*Child)
	// group in the first level, must return in search
	firstLevelGroupsDetails := make([]*Child, 0)
	for i := range groups {
		// reverse order
		g := groups[len(groups)-i-1]
		detail := ConvertGroupToGroupDetail(g)

		// groups under the parent group
		if g.ParentID == pGroupID {
			firstLevelGroupsDetails = append(firstLevelGroupsDetails, detail)
		}

		parentID := g.ParentID
		// name match or children's names match
		if strings.Contains(g.Name, filter) || len(parentIDToGroupsMap[int(g.ID)]) > 0 {
			parentIDToGroupsMap[parentID] = append(parentIDToGroupsMap[parentID], detail)
		}

		// current only query group table
		detail.Type = Group
		detail.FullPath = traversalIDsToFull[g.TraversalIDs].FullPath
		detail.FullName = traversalIDsToFull[g.TraversalIDs].FullName

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
		ParentID: common.RootGroupID,
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
		ParentID: common.RootGroupID,
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
