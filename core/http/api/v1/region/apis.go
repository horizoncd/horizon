package region

import (
	"g.hz.netease.com/horizon/core/controller/region"
	"g.hz.netease.com/horizon/pkg/server/response"
	"github.com/gin-gonic/gin"
)

type API struct {
	regionCtl region.Controller
}

func NewAPI(controller region.Controller) *API {
	return &API{regionCtl: controller}
}

func (a *API) listRegions(c *gin.Context) {
	regions, err := a.regionCtl.ListRegions(c)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, regions)
}
