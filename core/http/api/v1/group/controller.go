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
	DeleteGroupError    = "DeleteGroupError"
	GetGroupError       = "GetGroupError"
	GetGroupByPathError = "GetGroupByPathError"
	UpdateGroupError    = "UpdateGroupError"
	ParamGroupId        = "groupId"
	ParamPath           = "path"
	ParamFilter         = "filter"
	ParamParentId       = "parentId"
	ParentId            = "parent_id"
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
	groupId := c.Param(ParamGroupId)

	idInt, err := strconv.ParseUint(groupId, 10, 64)
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
	groupId := c.Param(ParamGroupId)

	idInt, err := strconv.ParseUint(groupId, 10, 64)
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
	groupId := c.Param(ParamGroupId)

	idInt, err := strconv.ParseUint(groupId, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, UpdateGroupError, fmt.Sprintf("upate group failed: %v", err))
		return
	}

	var updatedGroup *models.Group
	err = c.ShouldBindJSON(&updatedGroup)
	if err != nil {
		response.AbortWithRequestError(c, UpdateGroupError, fmt.Sprintf("upate group failed: %v", err))
		return
	}
	// 以URL path中的id为准
	updatedGroup.ID = uint(idInt)

	err = controller.manager.Update(c, updatedGroup)
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

	var details = make([]*GroupDetail, len(groups))
	for idx, tmp := range groups {
		detail := ConvertGroupToGroupDetail(tmp)
		details[idx] = detail
	}
	response.SuccessWithData(c, response.DataWithTotal{
		Total: count,
		Items: details,
	})
}

func (controller *Controller) SearchGroups(c *gin.Context) {
	filter := c.Param(ParamFilter)
	// 检索字符串为空，只展示 parent_id = null 的group
	if filter == "" {
		controller.GetSubGroups(c)
		return
	}
	// 检索字符串过短，返回空数组回去
	if len(filter) < 3 {
		response.SuccessWithData(c, []*models.Group{})
		return
	}
	// 正常检索

}

func formatQuerySubGroups(c *gin.Context) *q.Query {
	paramParentId := c.Query(ParamParentId)
	k := q.KeyWords{
		ParentId: nil,
	}
	if paramParentId != "" {
		k[ParentId], _ = strconv.Atoi(paramParentId)
	}

	// sort by updated_at desc，let newer items be in head
	s := q.NewSort("updated_at", true)
	query := q.New(k)
	query.PageNumber, _ = strconv.Atoi(c.Query(common.PageNumber))
	if query.PageNumber == 0 {
		query.PageNumber = common.DefaultPageNumber
	}
	query.PageSize, _ = strconv.Atoi(c.Query(common.PageSize))
	if query.PageSize == 0 {
		query.PageSize = common.DefaultPageSize
	}
	query.Sorts = []*q.Sort{s}

	return query
}
