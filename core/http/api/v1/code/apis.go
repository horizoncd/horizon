package code

import (
	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/code"
	"g.hz.netease.com/horizon/pkg/server/response"
	"github.com/gin-gonic/gin"
)

const (
	_gitURL = "giturl"
)

type API struct {
	codeCtl code.Controller
}

func NewAPI(codectl code.Controller) *API {
	return &API{codeCtl: codectl}
}

func (a *API) ListBranch(c *gin.Context) {
	gitURL := c.Query(_gitURL)
	if gitURL == "" {
		response.AbortWithRequestError(c, common.InvalidRequestParam, "giturl is empty")
		return
	}
	branches, err := a.codeCtl.ListBranch(c, gitURL)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, branches)
}
