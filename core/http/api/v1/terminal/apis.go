package terminal

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/terminal"
	"g.hz.netease.com/horizon/pkg/server/response"

	"github.com/gin-gonic/gin"
)

const (
	_clusterIDParam     = "clusterID"
	_podNameQuery       = "podName"
	_containerNameQuery = "containerName"
	_terminalIDParam    = "terminalID"
)

type API struct {
	terminalCtl terminal.Controller
}

func NewAPI(t terminal.Controller) *API {
	return &API{
		terminalCtl: t,
	}
}

func (a *API) CreateShell(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	podName := c.Query(_podNameQuery)
	containerName := c.Query(_containerNameQuery)

	sessionID, sockJS, err := a.terminalCtl.CreateShell(c, uint(clusterID), podName, containerName)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	// 修改url，以适配sockJS协议: GET {prefix}/{server_id}/{session_id}/websocket
	c.Request.URL.Path = fmt.Sprintf("/apis/core/v1/0/%s/websocket", sessionID)
	sockJS.ServeHTTP(c.Writer, c.Request)
}

func (a *API) CreateTerminal(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	podName := c.Query(_podNameQuery)
	containerName := c.Query(_containerNameQuery)

	terminalIDResp, err := a.terminalCtl.GetTerminalID(c, uint(clusterID), podName, containerName)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, terminalIDResp)
}

func (a *API) ConnectTerminal(c *gin.Context) {
	// todo(sph): add authorization and move session to db
	sessionID := c.Param(_terminalIDParam)
	sockJS, err := a.terminalCtl.GetSockJSHandler(c, sessionID)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	sockJS.ServeHTTP(c.Writer, c.Request)
}
