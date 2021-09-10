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

	idInt, err := strconv.ParseInt(groupId, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, DeleteGroupError, fmt.Sprintf("delete group failed: %v", err))
		return
	}
	err = controller.manager.Delete(c, idInt)
	if err != nil {
		response.AbortWithInternalError(c, DeleteGroupError, fmt.Sprintf("delete group failed: %v", err))
		return
	}
	response.Success(c)
}

func (controller *Controller) GetGroup(c *gin.Context) {
	groupId := c.Param(ParamGroupId)

	idInt, err := strconv.ParseInt(groupId, 10, 64)
	if err != nil {
		response.AbortWithRequestError(c, GetGroupError, fmt.Sprintf("get group failed: %v", err))
		return
	}

	_group, err := controller.manager.Get(c, idInt)
	if err != nil {
		response.AbortWithInternalError(c, GetGroupError, fmt.Sprintf("get group failed: %v", err))
		return
	}

	response.SuccessWithData(c, _group)
}

func (controller *Controller) GetGroupByPath(c *gin.Context) {
	path := c.Query(ParamPath)

	_group, err := controller.manager.GetByPath(c, path)
	if err != nil {
		response.AbortWithInternalError(c, GetGroupByPathError, fmt.Sprintf("get group by path failed: %v", err))
		return
	}

	response.SuccessWithData(c, _group)
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
	groups, err := controller.manager.List(c, formatQuerySubGroups(c))
	if err != nil {
		response.AbortWithInternalError(c, GetSubGroupsError, fmt.Sprintf("get subgroups failed: %v", err))
		return
	}

	var details []*GroupDetail
	for _, tmp := range groups {
		detail := ConvertGroupToGroupDetail(tmp)
		details = append(details, detail)
	}
	response.SuccessWithData(c, details)
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
	groupId := c.Param(ParamGroupId)
	if groupId == "" {
		// 数据库判断字段空值用null
		groupId = "null"
	}
	k := q.KeyWords{
		"parent_id": groupId,
	}
	// sort by updated_at desc，let newer items be in head
	s := q.NewSort("updated_at", true)
	query := q.New(k)
	query.PageNumber, _ = strconv.Atoi(c.Param(common.PAGE_NUMBER))
	query.PageSize, _ = strconv.Atoi(c.Param(common.PAGE_SIZE))
	query.Sorts = []*q.Sort{s}

	return query
}
