package group

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/group"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/request"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"

	"github.com/gin-gonic/gin"
)

const (
	_paramGroupID  = "groupID"
	_paramFullPath = "fullPath"
	_paramType     = "type"
)

type API struct {
	groupCtl group.Controller
}

// NewAPI initializes a new group api
func NewAPI(controller group.Controller) *API {
	return &API{
		groupCtl: controller,
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

// CreateSubGroup create a subgroup
func (a *API) CreateSubGroup(c *gin.Context) {
	groupID := c.Param(_paramGroupID)
	intID, err := strconv.ParseUint(groupID, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("%v", err))
		return
	}
	if intID <= 0 {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			"group id should be a positive integer")
		return
	}

	var newGroup *group.NewGroup
	err = c.ShouldBindJSON(&newGroup)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody, fmt.Sprintf("%v", err))
		return
	}

	newGroup.ParentID = uint(intID)
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
	intID, err := strconv.ParseUint(groupID, 10, 0)
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
	intID, err := strconv.ParseUint(groupID, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("%v", err))
		return
	}

	structuredGroup, err := a.groupCtl.GetByID(c, uint(intID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, structuredGroup)
}

func (a *API) ListAuthedGroup(c *gin.Context) {
	groups, err := a.groupCtl.ListAuthedGroup(c)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, groups)
}

// GetGroupByFullPath get a group child by fullPath
func (a *API) GetGroupByFullPath(c *gin.Context) {
	path := c.Query(_paramFullPath)
	resourceType := c.Query(_paramType)

	child, err := a.groupCtl.GetByFullPath(c, path, resourceType)
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
	intID, err := strconv.ParseUint(groupID, 10, 0)
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
	intID, err := strconv.ParseUint(groupID, 10, 0)
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
	groupID := c.Param(_paramGroupID)
	intID, err := strconv.ParseUint(groupID, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("invalid param, groupID: %s", groupID))
	}

	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	children, count, err := a.groupCtl.GetChildren(c, uint(intID), pageNumber, pageSize)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: count,
		Items: children,
	})
}

// GetSubGroups get subGroups of a group
func (a *API) GetSubGroups(c *gin.Context) {
	groupID := c.Param(_paramGroupID)
	intID, err := strconv.ParseUint(groupID, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("invalid param, groupID: %s", groupID))
	}

	pageNumber, pageSize, err := request.GetPageParam(c)
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
	groupID := c.Query(_paramGroupID)
	intID, err := strconv.ParseUint(groupID, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("invalid param, groupID: %s", groupID))
	}

	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	filter := c.Query(common.Filter)

	searchChildren, count, err := a.groupCtl.SearchChildren(c, &group.SearchParams{
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
		Items: searchChildren,
	})
}

// SearchGroups search subgroups of a group
func (a *API) SearchGroups(c *gin.Context) {
	groupID := c.Query(_paramGroupID)
	intID, err := strconv.ParseUint(groupID, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("invalid param, groupID: %s", groupID))
	}

	pageNumber, pageSize, err := request.GetPageParam(c)
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

func (a *API) UpdateRegionSelector(c *gin.Context) {
	groupID := c.Param(_paramGroupID)
	intID, err := strconv.ParseUint(groupID, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, fmt.Sprintf("invalid param, groupID: %s", groupID))
	}

	var regionSelectors group.RegionSelectors
	err = c.ShouldBindJSON(&regionSelectors)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody, fmt.Sprintf("%v", err))
		return
	}

	// todo validate format of regionSelector param
	err = a.groupCtl.UpdateRegionSelector(c, uint(intID), regionSelectors)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}
