package build

import (
	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/core/common"
	"github.com/horizoncd/horizon/pkg/core/controller/build"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/util/log"
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
