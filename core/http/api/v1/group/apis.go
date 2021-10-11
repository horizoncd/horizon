package group

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/group"
	"g.hz.netease.com/horizon/pkg/server/response"

	"github.com/gin-gonic/gin"
)

const (
	_paramGroupID = "groupID"
	_paramPath    = "path"
)

type API struct {
	groupCtl group.Controller
}

// NewAPI initializes a new group api
func NewAPI() *API {
	return &API{
		groupCtl: group.Ctl,
	}
}

// CreateGroup create a group
func (a *API) CreateGroup(c *gin.Context) {
	var newGroup *group.NewGroup
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

// DeleteGroup delete a group by id
func (a *API) DeleteGroup(c *gin.Context) {
	groupID := c.Param(_paramGroupID)
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

// GetGroup get a group child by id
func (a *API) GetGroup(c *gin.Context) {
	groupID := c.Param(_paramGroupID)
	intID, err := strconv.Atoi(groupID)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("%v", err))
		return
	}

	child, err := a.groupCtl.GetByID(c, uint(intID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, child)
}

// GetGroupByFullPath get a group child by fullPath
func (a *API) GetGroupByFullPath(c *gin.Context) {
	path := c.Query(_paramPath)

	child, err := a.groupCtl.GetByFullPath(c, path)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, child)
}

// TransferGroup transfer a group to another parent group
func (a *API) TransferGroup(c *gin.Context) {
	groupID := c.Param(_paramGroupID)
	parentID := c.Query(_paramGroupID)
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

// UpdateGroup update basic info of a group
func (a *API) UpdateGroup(c *gin.Context) {
	groupID := c.Param(_paramGroupID)

	intID, err := strconv.Atoi(groupID)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("%v", err))
		return
	}

	var updatedGroup *group.UpdateGroup
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

// GetChildren get children of a group, including groups and applications
func (a *API) GetChildren(c *gin.Context) {
	// todo also query application
	a.GetSubGroups(c)
}

// GetSubGroups get subGroups of a group
func (a *API) GetSubGroups(c *gin.Context) {
	groupID := c.Param(_paramGroupID)
	intID, err := strconv.Atoi(groupID)
	if err != nil || intID < -1 {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("invalid param, groupID: %s", groupID))
	}

	pageNumber, pageSize, err := checkPageParamsOnListingGroups(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
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

// SearchChildren search children of a group, including groups and applications
func (a *API) SearchChildren(c *gin.Context) {
	// TODO(wurongjun): also query application
	a.SearchGroups(c)
}

// SearchGroups search subgroups of a group
func (a *API) SearchGroups(c *gin.Context) {
	groupID := c.Query(_paramGroupID)
	intID, err := strconv.Atoi(groupID)
	if err != nil || intID < -1 {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("invalid param, groupID: %s", groupID))
	}

	pageNumber, pageSize, err := checkPageParamsOnListingGroups(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	filter := c.Query(common.Filter)

	searchGroups, count, err := a.groupCtl.SearchGroups(c, &group.SearchParams{
		GroupID:    uint(intID),
		PageSize:   pageSize,
		PageNumber: pageNumber,
		Filter:     filter,
	})
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: count,
		Items: searchGroups,
	})
}

// checkPageParamsOnListingGroups check whether the params for listing groups is valid
func checkPageParamsOnListingGroups(c *gin.Context) (int, int, error) {
	pNumber := c.Query(common.PageNumber)
	pageNumber, err := strconv.Atoi(pNumber)
	if err != nil || pageNumber <= 0 {
		return 0, 0, fmt.Errorf("invalid param, pageNumber: %d", pageNumber)
	}
	pSize := c.Query(common.PageSize)
	pageSize, err := strconv.Atoi(pSize)
	if err != nil || pageSize <= 0 || pageSize > common.MaxPageSize {
		return 0, 0, fmt.Errorf("invalid param, pageSize: %d", pageSize)
	}

	return pageNumber, pageSize, nil
}
