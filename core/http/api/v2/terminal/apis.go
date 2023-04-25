// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terminal

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/controller/terminal"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
)

const (
	_clusterIDParam     = "clusterID"
	_podNameQuery       = "podName"
	_containerNameQuery = "containerName"
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

	// Before using sock JS to process the request, modify the URL to adapt to the sock JS protocol:
	// GET {prefix}/{server_id}/{session_id}/websocket
	// server_id：The server ID is mainly used in the sockjs protocol to connect multiple servers.
	// We can fill in a fixed value (such as: 0), and the user has no perception
	// session_id: Session ID, used for multi-person sessions or reconnection,
	// we are currently reconnecting to open a new session,
	// so the session can be directly generated and passed in the current interface, the user has no perception
	c.Request.URL.Path = fmt.Sprintf("/apis/core/v2/0/%s/websocket", sessionID)
	sockJS.ServeHTTP(c.Writer, c.Request)
}
