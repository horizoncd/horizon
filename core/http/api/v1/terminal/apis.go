package terminal

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/terminal"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
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
	const op = "terminal: create shell"
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid cluster id: %s, "+
			"err: %s", clusterIDStr, err.Error())))
		return
	}
	podName := c.Query(_podNameQuery)
	containerName := c.Query(_containerNameQuery)

	sessionID, sockJS, err := a.terminalCtl.CreateShell(c, uint(clusterID), podName, containerName)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.ClusterInDB || e.Source == herrors.PodsInK8S {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	// 在使用sockJS处理请求前，修改URL，以适配sockJS协议: GET {prefix}/{server_id}/{session_id}/websocket
	// server_id：服务端ID，主要在sockjs协议中用于连接多个服务端，我们填写固定值（如：0）即可，用户无感知
	// session_id: 会话ID，用于多人会话或者重连，我们目前重连均是开启新会话，因此session直接在当前接口生成并传入即可，用户无感知
	c.Request.URL.Path = fmt.Sprintf("/apis/core/v1/0/%s/websocket", sessionID)
	sockJS.ServeHTTP(c.Writer, c.Request)
}

func (a *API) CreateTerminal(c *gin.Context) {
	const op = "terminal: create terminal"
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
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, terminalIDResp)
}

func (a *API) ConnectTerminal(c *gin.Context) {
	const op = "terminal: connect terminal"
	sessionID := c.Param(_terminalIDParam)
	sockJS, err := a.terminalCtl.GetSockJSHandler(c, sessionID)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.ClusterInDB || e.Source == herrors.PodsInK8S {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithError(c, err)
		return
	}
	sockJS.ServeHTTP(c.Writer, c.Request)
}
