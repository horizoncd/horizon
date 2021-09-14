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
	manager group.Manager
}

func NewController() *Controller {
	return &Controller{
		manager: group.Mgr,
	}
}

func (controller *Controller) CreateGroup(c *gin.Context) {
	var newGroup *NewGroup
	err := c.ShouldBindJSON(&newGroup)
	if err != nil {
		response.AbortWithRequestError(c, CreateGroupError, fmt.Sprintf("create group failed: %v", err))
		return
	}

	create, err := controller.manager.Create(c, convertNewGroupToGroup(newGroup))
	if err != nil {
		response.AbortWithInternalError(c, CreateGroupError, fmt.Sprintf("create group failed: %v", err))
		return
	}

	response.SuccessWithData(c, create)
}

func (controller *Controller) DeleteGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)

	idInt, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, DeleteGroupError, fmt.Sprintf("delete group failed: %v", err))
		return
	}
	err = controller.manager.Delete(c, uint(idInt))
	if err != nil {
		response.AbortWithInternalError(c, DeleteGroupError, fmt.Sprintf("delete group failed: %v", err))
		return
	}
	response.Success(c)
}

func (controller *Controller) GetGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)

	idInt, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, GetGroupError, fmt.Sprintf("get group failed: %v", err))
		return
	}

	_group, err := controller.manager.Get(c, uint(idInt))
	if err != nil {
		response.AbortWithInternalError(c, GetGroupError, fmt.Sprintf("get group failed: %v", err))
		return
	}
	detail := ConvertGroupToGroupDetail(_group)
	response.SuccessWithData(c, detail)
}

func (controller *Controller) GetGroupByPath(c *gin.Context) {
	path := c.Query(ParamPath)

	_group, err := controller.manager.GetByPath(c, path)
	if err != nil {
		response.AbortWithInternalError(c, GetGroupByPathError, fmt.Sprintf("get group by path failed: %v", err))
		return
	}
	if _group == nil {
		response.AbortWithNotFoundError(c, GetGroupByPathError, fmt.Sprintf("get group by path failed: %v", err))
		return
	}
	detail := ConvertGroupToGroupDetail(_group)
	response.SuccessWithData(c, detail)
}

func (controller *Controller) UpdateGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)

	idInt, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, UpdateGroupError, fmt.Sprintf("upate group failed: %v", err))
		return
	}

	var updatedGroup *UpdateGroup
	err = c.ShouldBindJSON(&updatedGroup)
	if err != nil {
		response.AbortWithRequestError(c, UpdateGroupError, fmt.Sprintf("upate group failed: %v", err))
		return
	}

	_group := convertUpdateGroupToGroup(updatedGroup)
	_group.ID = uint(idInt)
	err = controller.manager.Update(c, _group)
	if err != nil {
		response.AbortWithInternalError(c, UpdateGroupError, fmt.Sprintf("upate group failed: %v", err))
		return
	}

	response.Success(c)
}

func (controller *Controller) GetChildren(c *gin.Context) {
	// todo also query application
	controller.GetSubGroups(c)
}

func (controller *Controller) GetSubGroups(c *gin.Context) {
	groups, count, err := controller.manager.List(c, formatQuerySubGroups(c))
	if err != nil {
		response.AbortWithInternalError(c, GetSubGroupsError, fmt.Sprintf("get subgroups failed: %v", err))
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: count,
		Items: controller.formatPageGroupDetails(c, groups),
	})
}

func (controller *Controller) SearchChildren(c *gin.Context) {
	// todo also query application
	controller.SearchGroups(c)
}

func (controller *Controller) SearchGroups(c *gin.Context) {
	filter := c.Query(ParamFilter)
	if filter == "" {
		groups, count, err := controller.manager.List(c, formatSearchGroups(c))
		if err != nil {
			response.AbortWithInternalError(c, SearchGroupsError, fmt.Sprintf("search groups failed: %v", err))
			return
		}
		response.SuccessWithData(c, response.DataWithTotal{
			Total: count,
			Items: controller.formatPageGroupDetails(c, groups),
		})
		return
	}
	// 检索字符串过短，返回空数组回去
	if len(filter) < 3 {
		response.SuccessWithData(c, response.DataWithTotal{
			Total: 0,
			Items: []*models.Group{},
		})
		return
	}
	// 正常检索，根据name检索
	// groups, err := controller.manager.GetByNameFuzzily(c, filter)
	// if err != nil {
	// 	response.AbortWithInternalError(c, SearchGroupsError, fmt.Sprintf("search groups failed: %v", err))
	// 	return
	// }
}

func (controller *Controller) formatPageGroupDetails(c *gin.Context, groups []*models.Group) []*Child {
	var parentIds []uint
	for _, m := range groups {
		parentIds = append(parentIds, m.ID)
	}
	query := q.New(q.KeyWords{
		ParentID: parentIds,
	})
	subGroups, err := controller.manager.ListWithoutPage(c, query)
	if err != nil {
		response.AbortWithInternalError(c, GetSubGroupsError, fmt.Sprintf("get subgroups failed: %v", err))
		return nil
	}
	childrenCountMap := map[uint]int{}
	for _, subgroup := range subGroups {
		if v, ok := childrenCountMap[*subgroup.ParentID]; ok {
			childrenCountMap[*subgroup.ParentID] = v + 1
		} else {
			childrenCountMap[*subgroup.ParentID] = 1
		}
	}

	var details = make([]*Child, len(groups))
	for idx, tmp := range groups {
		detail := ConvertGroupToGroupDetail(tmp)
		// todo currently using fixed type: group
		detail.Type = Group
		detail.ChildrenCount = childrenCountMap[detail.ID]
		details[idx] = detail
	}

	return details
}

// url pattern: api/vi/groups/:groupId/subgroups
func formatQuerySubGroups(c *gin.Context) *q.Query {
	parentID := c.Param(ParamGroupID)
	k := q.KeyWords{
		ParentID: nil,
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
		ParentID: nil,
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
