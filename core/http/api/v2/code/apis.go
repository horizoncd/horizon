package code

import (
	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/code"
	codegetter "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/server/request"
	"github.com/horizoncd/horizon/pkg/server/response"
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
	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	filter := c.Query(common.Filter)

	branches, err := a.codeCtl.ListBranch(c, gitURL, &codegetter.SearchParams{
		Filter:     filter,
		PageNumber: pageNumber,
		PageSize:   pageSize,
	})
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, branches)
}

func (a *API) ListTag(c *gin.Context) {
	gitURL := c.Query(_gitURL)
	if gitURL == "" {
		response.AbortWithRequestError(c, common.InvalidRequestParam, "giturl is empty")
		return
	}
	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	filter := c.Query(common.Filter)

	tags, err := a.codeCtl.ListTag(c, gitURL, &codegetter.SearchParams{
		Filter:     filter,
		PageNumber: pageNumber,
		PageSize:   pageSize,
	})
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, tags)
}
