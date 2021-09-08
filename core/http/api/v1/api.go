package v1

import (
	"net/http"

	"g.hz.netease.com/horizon/lib/log"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
)

func (controller *Controller) Test(c *gin.Context) {
	logger := log.GetLogger(c)
	for k, v := range c.Request.Header {
		logger.Infof("%s: %s", k, v)
	}

	name := c.Request.Header.Get("X-HORIZON-OIDC-USER")
	email := c.Request.Header.Get("X-HORIZON-OIDC-EMAIL")
	if len(name) == 0 {
		response.Abort(c, http.StatusUnauthorized,
			http.StatusText(http.StatusUnauthorized), http.StatusText(http.StatusUnauthorized))
		return
	}

	response.SuccessWithData(c, User{
		Name:  name,
		Email: email,
	})
}