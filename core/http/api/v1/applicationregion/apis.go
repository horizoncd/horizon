package applicationregion

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/controller/applicationregion"
	ccommon "g.hz.netease.com/horizon/core/controller/common"
	"g.hz.netease.com/horizon/pkg/server/response"

	"github.com/gin-gonic/gin"
)

const (
	// param
	_applicationIDParam = "applicationID"
)

type API struct {
	applicationRegionCtl applicationregion.Controller
}

func NewAPI(applicationRegionCtl applicationregion.Controller) *API {
	return &API{
		applicationRegionCtl: applicationRegionCtl,
	}
}

func (a *API) List(c *gin.Context) {
	applicationIDStr := c.Param(_applicationIDParam)
	applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
	if err != nil {
		ccommon.Response(c, ccommon.ParamError.WithErrMsg(err.Error()))
		return
	}
	var res applicationregion.ApplicationRegion
	if res, err = a.applicationRegionCtl.List(c, uint(applicationID)); err != nil {
		ccommon.Response(c, ccommon.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, res)
}

func (a *API) Update(c *gin.Context) {
	applicationIDStr := c.Param(_applicationIDParam)
	applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
	if err != nil {
		ccommon.Response(c, ccommon.ParamError.WithErrMsg(err.Error()))
		return
	}

	var request map[string]string
	if err := c.ShouldBindJSON(&request); err != nil {
		ccommon.Response(c, ccommon.ParamError.WithErrMsg(fmt.Sprintf("request body is invalid, err: %v", err)))
		return
	}

	if err := a.applicationRegionCtl.Update(c, uint(applicationID), request); err != nil {
		ccommon.Response(c, ccommon.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}
