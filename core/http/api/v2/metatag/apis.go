package metatag

import (
	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/controller/metatag"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
)

const (
	_tagKey = "key"
)

type API struct {
	metaController metatag.Controller
}

func NewAPI(metaController metatag.Controller) *API {
	return &API{
		metaController: metaController,
	}
}

func (a *API) GetMetatagKeys(c *gin.Context) {
	keys, err := a.metaController.GetMetatagKeys(c)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, keys)
}

func (a *API) GetMetatagsByKey(c *gin.Context) {
	key := c.Query(_tagKey)
	metatags, err := a.metaController.GetMetatagsByKey(c, key)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, metatags)
}

func (a *API) CreateMetatags(c *gin.Context) {
	var createMetatagsRequest metatag.CreateMetatagsRequest
	if err := c.ShouldBindJSON(&createMetatagsRequest); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}
	if err := a.metaController.CreateMetatags(c, &createMetatagsRequest); err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}
