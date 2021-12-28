package terminal

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	ccommon "g.hz.netease.com/horizon/core/controller/common"
	"g.hz.netease.com/horizon/core/controller/terminal"
	perrors "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/util/log"
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
	const op = "terminal: sockjs"
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		err = perrors.Wrap(err, "failed to parse cluster id")
		log.WithFiled(c, "op", op).Errorf(err.Error())
		ccommon.Response(c, ccommon.ParamError.WithErrMsg("invalid cluster id"))
		return
	}
	podName := c.Query(_podNameQuery)
	containerName := c.Query(_containerNameQuery)

	sessionID, sockJS, err := a.terminalCtl.CreateShell(c, uint(clusterID), podName, containerName)
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		ccommon.Response(c, ccommon.InternalError.WithErrMsg(err.Error()))
		return
	}

	// 在使用sockJS处理请求前，修改URL，以适配sockJS协议: GET {prefix}/{server_id}/{session_id}/websocket
	// server_id：服务端ID，主要在sockjs协议中用于连接多个服务端，我们填写固定值（如：0）即可，用户无感知
	// session_id: 会话ID，用于多人会话或者重连，我们目前重连均是开启新会话，因此session直接在当前接口生成并传入即可，用户无感知
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
