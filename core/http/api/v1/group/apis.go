package group

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/common"
	groupctl "g.hz.netease.com/horizon/controller/group"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
)

const (
	ParamGroupID  = "groupID"
	ParamPath     = "path"
	QueryParentID = "parentID"

	RootGroupID = 0
)

type API struct {
	groupCtl groupctl.Controller
}

func NewAPI() *API {
	return &API{
		groupCtl: groupctl.Ctl,
	}
}

func (a *API) CreateGroup(c *gin.Context) {
	var newGroup *groupctl.NewGroup
	err := c.ShouldBindJSON(&newGroup)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody, fmt.Sprintf("%v", err))
		return
	}

	id, err := a.groupCtl.CreateGroup(c, newGroup)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, id)
}

func (a *API) DeleteGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)
	intID, err := strconv.Atoi(groupID)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("%v", err))
		return
	}

	err = a.groupCtl.Delete(c, uint(intID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.Success(c)
}

func (a *API) GetGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)
	intID, err := strconv.Atoi(groupID)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("%v", err))
		return
	}

	gChild, err := a.groupCtl.GetByID(c, uint(intID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, gChild)
}

func (a *API) GetGroupByPath(c *gin.Context) {
	path := c.Query(ParamPath)

	gChild, err := a.groupCtl.GetByPath(c, path)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, gChild)
}

func (a *API) TransferGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)
	parentID := c.Query(QueryParentID)
	intID, err := strconv.Atoi(groupID)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("%v", err))
		return
	}
	pIDInt, err := strconv.Atoi(parentID)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("%v", err))
		return
	}

	err = a.groupCtl.Transfer(c, uint(intID), uint(pIDInt))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.Success(c)
}

func (a *API) UpdateGroup(c *gin.Context) {
	groupID := c.Param(ParamGroupID)

	intID, err := strconv.Atoi(groupID)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("%v", err))
		return
	}

	var updatedGroup *groupctl.UpdateGroup
	err = c.ShouldBindJSON(&updatedGroup)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody, fmt.Sprintf("%v", err))
		return
	}

	err = a.groupCtl.UpdateBasic(c, uint(intID), updatedGroup)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.Success(c)
}

func (a *API) GetChildren(c *gin.Context) {
	// todo also query application
	a.GetSubGroups(c)
}

func (a *API) GetSubGroups(c *gin.Context) {
	parentID := c.Param(ParamGroupID)
	intID, err := strconv.Atoi(parentID)
	if err != nil || intID < -1 {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("invalid param, groupID: %s", parentID))
		return
	}

	pNumber := c.Query(common.PageNumber)
	pageNumber, err := strconv.Atoi(pNumber)
	if err != nil || pageNumber < 0 {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("invalid param, pageNumber: %s", pNumber))
		return
	}
	pSize := c.Query(common.PageSize)
	pageSize, err := strconv.Atoi(pSize)
	if err != nil || pageNumber <= 0 || pageSize > common.MaxPageSize {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("invalid param, pageSize: %s", pSize))
		return
	}

	subGroups, count, err := a.groupCtl.GetSubGroups(c, uint(intID), pageNumber, pageSize)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: count,
		Items: subGroups,
	})
}

func (a *API) SearchChildren(c *gin.Context) {
	// TODO(wurongjun): also query application
	a.SearchGroups(c)
}

func (a *API) SearchGroups(c *gin.Context) {
	parentID := c.Query(QueryParentID)
	intID, err := strconv.Atoi(parentID)
	if err != nil || intID < -1 {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("invalid param, parentID: %s", parentID))
		return
	}

	filter := c.Query(common.Filter)

	searchGroups, count, err := a.groupCtl.SearchGroups(c, uint(intID), filter)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: count,
		Items: searchGroups,
	})
}
