package group

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group"
	"g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
)

const (
	CreateGroupError    = "CreateGroupError"
	GetSubGroupsError   = "GetSubGroupsError"
	SearchGroupsError   = "SearchGroupsError"
	NotImplemented      = "NotImplemented"
	DeleteGroupError    = "DeleteGroupError"
	GetGroupError       = "GetGroupError"
	GetGroupByPathError = "GetGroupByPathError"
	UpdateGroupError    = "UpdateGroupError"
	ParamGroupID        = "groupID"
	ParamPath           = "path"
	ParamFilter         = "filter"
	QueryParentID       = "parentId"
	ParentID            = "parent_id"
	Group               = "group"
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
		response.AbortWithRequestError(c, CreateGroupError,
			fmt.Sprintf("create group failed: %v", err))
		return
	}

	create, err := controller.groupManager.Create(c, convertNewGroupToGroup(newGroup))
	if err != nil {
		response.AbortWithInternalError(c, CreateGroupError,
			fmt.Sprintf("create group failed: %v", err))
		return
	}

	response.SuccessWithData(c, create)
}

func (controller *Controller) DeleteGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)

	idInt, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, DeleteGroupError,
			fmt.Sprintf("delete group failed: %v", err))
		return
	}
	err = controller.groupManager.Delete(c, uint(idInt))
	if err != nil {
		response.AbortWithInternalError(c, DeleteGroupError,
			fmt.Sprintf("delete group failed: %v", err))
		return
	}
	response.Success(c)
}

func (controller *Controller) GetGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)

	idInt, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, GetGroupError,
			fmt.Sprintf("get group failed: %v", err))
		return
	}

	groupEntry, err := controller.groupManager.Get(c, uint(idInt))
	if err != nil {
		response.AbortWithInternalError(c, GetGroupError,
			fmt.Sprintf("get group failed: %v", err))
		return
	}
	detail := ConvertGroupToGroupDetail(groupEntry)
	response.SuccessWithData(c, detail)
}

func (controller *Controller) GetGroupByPath(c *gin.Context) {
	path := c.Query(ParamPath)

	groupEntry, err := controller.groupManager.GetByPath(c, path)
	if err != nil {
		response.AbortWithInternalError(c, GetGroupByPathError,
			fmt.Sprintf("get group by path failed: %v", err))
		return
	}
	if groupEntry == nil {
		response.AbortWithNotFoundError(c, GetGroupByPathError,
			fmt.Sprintf("get group by path failed: %v", err))
		return
	}
	detail := ConvertGroupToGroupDetail(groupEntry)
	response.SuccessWithData(c, detail)
}

// TODO(wurongjun) support transfer group
// func (controller *Controller) TransferGroup(c *gin.Context)

// TODO(wurongjun) change to UpdateGroupBasic (also change the openapi)
func (controller *Controller) UpdateGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)

	idInt, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, UpdateGroupError,
			fmt.Sprintf("upate group failed: %v", err))
		return
	}

	var updatedGroup *UpdateGroup
	err = c.ShouldBindJSON(&updatedGroup)
	if err != nil {
		response.AbortWithRequestError(c, UpdateGroupError,
			fmt.Sprintf("upate group failed: %v", err))
		return
	}

	groupEntry := convertUpdateGroupToGroup(updatedGroup)
	groupEntry.ID = uint(idInt)
	err = controller.groupManager.Update(c, groupEntry)
	if err != nil {
		response.AbortWithInternalError(c, UpdateGroupError,
			fmt.Sprintf("upate group failed: %v", err))
		return
	}

	response.Success(c)
}

func (controller *Controller) GetChildren(c *gin.Context) {
	// todo also query application
	controller.GetSubGroups(c)
}

func (controller *Controller) GetSubGroups(c *gin.Context) {
	groups, count, err := controller.groupManager.List(c, formatQuerySubGroups(c))
	if err != nil {
		response.AbortWithInternalError(c, GetSubGroupsError,
			fmt.Sprintf("get subgroups failed: %v", err))
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: count,
		Items: controller.formatPageGroupDetails(c, groups),
	})
}

func (controller *Controller) SearchChildren(c *gin.Context) {
	//TODO(wurongjun): also query application
	controller.SearchGroups(c)
}

func (controller *Controller) SearchGroups(c *gin.Context) {
	filter := c.Query(ParamFilter)

	// filter is empty, just list the group
	if filter == "" {
		groups, count, err := controller.groupManager.List(c, formatSearchGroups(c))
		if err != nil {
			response.AbortWithInternalError(c, SearchGroupsError,
				fmt.Sprintf("search groups failed: %v", err))
			return
		}
		response.SuccessWithData(c, response.DataWithTotal{
			Total: count,
			Items: controller.formatPageGroupDetails(c, groups),
		})
		return
	}

	// filter is too short will be ignore
	if len(filter) < 3 {
		response.SuccessWithData(c, response.DataWithTotal{
			Total: 0,
			Items: []*models.Group{},
		})
		return
	}

	//TODO(wurongjun): not implemented yet.
	response.AbortWithRequestError(c, NotImplemented,
		"search groups by name have not implemented")
}

func (controller *Controller) formatPageGroupDetails(c *gin.Context, groups []*models.Group) []*Child {
	var parentIds []uint
	for _, m := range groups {
		parentIds = append(parentIds, m.ID)
	}
	query := q.New(q.KeyWords{
		ParentID: parentIds,
	})
	subGroups, err := controller.groupManager.ListWithoutPage(c, query)
	if err != nil {
		response.AbortWithInternalError(c, GetSubGroupsError,
			fmt.Sprintf("get subgroups failed: %v", err))
		return nil
	}
	childrenCountMap := map[int]int{}
	for _, subgroup := range subGroups {
		if v, ok := childrenCountMap[subgroup.ParentID]; ok {
			childrenCountMap[subgroup.ParentID] = v + 1
		} else {
			childrenCountMap[subgroup.ParentID] = 1
		}
	}

	var details = make([]*Child, len(groups))
	for idx, tmp := range groups {
		detail := ConvertGroupToGroupDetail(tmp)
		// todo currently using fixed type: group
		detail.Type = Group
		detail.ChildrenCount = childrenCountMap[int(detail.ID)]
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

	query := formatDefaultQuery()
	query.Keywords = k
	pageNumber, _ := strconv.Atoi(c.Query(common.PageNumber))
	if pageNumber > 0 {
		query.PageNumber = pageNumber
	}

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

	query := formatDefaultQuery()
	query.Keywords = k
	pageNumber, _ := strconv.Atoi(c.Query(common.PageNumber))
	if pageNumber > 0 {
		query.PageNumber = pageNumber
	}

	return query
}

func formatDefaultQuery() *q.Query {
	// sort by updated_at desc，let newer items be in head
	s := q.NewSort("updated_at", true)
	query := q.New(q.KeyWords{})
	query.PageNumber = common.DefaultPageNumber
	query.PageSize = common.DefaultPageSize
	query.Sorts = []*q.Sort{s}

	return query
}
