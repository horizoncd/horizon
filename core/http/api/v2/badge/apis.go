package badge

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/badge"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
)

var (
	_paramBadgeIDorName = "badgeIDorName"
)

func getResourceContext(c *gin.Context) (string, uint, error) {
	resourceType := c.Param(common.ParamResourceType)
	resourceIDStr := c.Param(common.ParamResourceID)

	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)

	if err != nil {
		errMsg := fmt.Sprintf("invalid : %s, err: %s", resourceIDStr, err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(errMsg))
		return "", 0, fmt.Errorf(errMsg)
	}

	return resourceType, uint(resourceID), nil
}

type API struct {
	badgeCtl badge.Controller
}

func NewAPI(badgeCtl badge.Controller) *API {
	return &API{badgeCtl: badgeCtl}
}

func (a *API) Create(c *gin.Context) {
	// op := "badge: create"
	resourceType, resourceID, err := getResourceContext(c)
	if err != nil {
		return
	}

	badgeCreate := badge.Create{}

	if err := c.ShouldBindJSON(&badgeCreate); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	if b, err := a.badgeCtl.CreateBadge(c, resourceType, uint(resourceID), &badgeCreate); err != nil {
		if innerErr := perror.Cause(err); errors.Is(innerErr, herrors.ErrParamInvalid) {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("failed to create badge, err: %s",
			err.Error())))
	} else {
		response.SuccessWithData(c, b)
	}
}

func (a *API) UpdateClusterBadge(c *gin.Context) {
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)

	if err != nil {
		errMsg := fmt.Sprintf("invalid : %s, err: %s", clusterIDStr, err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(errMsg))
		return
	}
	a.update(c, common.ResourceCluster, uint(clusterID))
}

func (a *API) update(c *gin.Context, resourceType string, resourceID uint) {
	// op := "badge: update"
	badgeIDorName := c.Param(_paramBadgeIDorName)

	badgeID, err := strconv.ParseUint(badgeIDorName, 10, 0)
	badgeName := ""
	if err != nil {
		badgeName = badgeIDorName
	}

	badgeUpdate := badge.Update{}

	if err := c.ShouldBindJSON(&badgeUpdate); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	if badgeName == "" {
		if b, err := a.badgeCtl.UpdateBadge(c, uint(badgeID), &badgeUpdate); err != nil {
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("failed to update badge, err: %s",
				err.Error())))
		} else {
			response.SuccessWithData(c, b)
		}
	} else {
		if b, err := a.badgeCtl.UpdateBadgeByName(c, resourceType, uint(resourceID), badgeName, &badgeUpdate); err != nil {
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("failed to update badge, err: %s",
				err.Error())))
		} else {
			response.SuccessWithData(c, b)
		}
	}
}

func (a *API) Delete(c *gin.Context) {
	// op := "badge: delete"
	resourceType, resourceID, err := getResourceContext(c)
	if err != nil {
		return
	}

	badgeIDorName := c.Param(_paramBadgeIDorName)

	badgeID, err := strconv.ParseUint(badgeIDorName, 10, 0)
	badgeName := ""
	if err != nil {
		badgeName = badgeIDorName
	}

	if badgeName == "" {
		if err := a.badgeCtl.DeleteBadge(c, uint(badgeID)); err != nil {
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("failed to delete badge, err: %s",
				err.Error())))
		} else {
			response.Success(c)
		}
	} else {
		if err := a.badgeCtl.DeleteBadgeByName(c, resourceType, resourceID, badgeName); err != nil {
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("failed to delete badge, err: %s",
				err.Error())))
		} else {
			response.Success(c)
		}
	}
}

func (a *API) Get(c *gin.Context) {
	// op := "badge: get"
	resourceType, resourceID, err := getResourceContext(c)
	if err != nil {
		return
	}

	badgeIDorName := c.Param(_paramBadgeIDorName)

	badgeID, err := strconv.ParseUint(badgeIDorName, 10, 0)
	badgeName := ""
	if err != nil {
		badgeName = badgeIDorName
	}

	if badgeName == "" {
		if b, err := a.badgeCtl.GetBadge(c, uint(badgeID)); err != nil {
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("failed to get badge, err: %s",
				err.Error())))
		} else {
			response.SuccessWithData(c, b)
		}
	} else {
		if b, err := a.badgeCtl.GetBadgeByName(c, resourceType, resourceID, badgeName); err != nil {
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("failed to get badge, err: %s",
				err.Error())))
		} else {
			response.SuccessWithData(c, b)
		}
	}
}

func (a *API) List(c *gin.Context) {
	// op := "badge: list"
	resourceType, resourceID, err := getResourceContext(c)
	if err != nil {
		return
	}

	if badges, err := a.badgeCtl.ListBadges(c, resourceType, resourceID); err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("failed to list badges, err: %s",
			err.Error())))
	} else {
		response.SuccessWithData(c, badges)
	}
}
