package user

import (
	"net/http"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/user"
	usermiddle "g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/server/response"
	"github.com/gin-gonic/gin"
)

type API struct {
	userCtl user.Controller
}

func NewAPI() *API {
	return &API{
		userCtl: user.Ctl,
	}
}

func (a *API) Search(c *gin.Context) {
	var (
		filter               string
		pageNumber, pageSize int

		err error
	)
	filter = c.Query(common.Filter)

	pageNumberStr := c.Query(common.PageNumber)
	pageNumber, err = strconv.Atoi(pageNumberStr)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid pageNumber")
		return
	}

	pageSizeStr := c.Query(common.PageSize)
	pageSize, err = strconv.Atoi(pageSizeStr)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid pageSize")
		return
	}

	if pageNumber < 1 {
		pageNumber = common.DefaultPageNumber
	}

	if pageSize < 0 {
		pageSize = common.DefaultPageSize
	}

	count, res, err := a.userCtl.SearchUser(c, filter, &q.Query{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	})
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(count),
		Items: res,
	})
}

func (a *API) Status(c *gin.Context) {
	u, err := usermiddle.FromContext(c)
	if err != nil {
		response.Abort(c, http.StatusForbidden, common.Forbidden, "user not logged in")
	}
	response.SuccessWithData(c, struct {
		Name string `json:"name"`
		ID   uint   `json:"id"`
	}{
		Name: u.GetFullName(),
		ID:   u.GetID(),
	})
}
