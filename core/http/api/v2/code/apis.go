package code

import (
	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/code"
	"github.com/horizoncd/horizon/pkg/git"
	"github.com/horizoncd/horizon/pkg/server/request"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/util/log"
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

	branches, err := a.codeCtl.ListBranch(c, gitURL, &git.SearchParams{
		Filter:     filter,
		PageNumber: pageNumber,
		PageSize:   pageSize,
	})
	if err != nil {
		log.Errorf(c, "List branch error: %+v", err)
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

	tags, err := a.codeCtl.ListTag(c, gitURL, &git.SearchParams{
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
