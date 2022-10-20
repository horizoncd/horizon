package build

import (
	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/build"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

type API struct {
	buildCtl build.Controller
}

func NewAPI(buildCtl build.Controller) *API {
	return &API{buildCtl: buildCtl}
}

func (a *API) Get(c *gin.Context) {
	const op = "build: schemaGet"
	schema, err := a.buildCtl.GetSchema(c)
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRequestError(c, common.InternalError, err.Error())
	}
	response.SuccessWithData(c, schema)
}
