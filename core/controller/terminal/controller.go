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
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/kubeclient"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/gitrepo"
	manager "github.com/horizoncd/horizon/pkg/manager"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/util/errors"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"k8s.io/client-go/tools/remotecommand"

	"gopkg.in/igm/sockjs-go.v3/sockjs"
)

type Controller interface {
	GetTerminalID(ctx context.Context, clusterID uint, podName, containerName string) (*SessionIDResp, error)
	GetSockJSHandler(ctx context.Context, sessionID string) (http.Handler, error)
	// CreateShell returns sessionID and sockJSHandler according to clusterID,podName,containerName
	CreateShell(ctx context.Context, clusterID uint, podName, containerName string) (sessionID string,
		sockJSHandler http.Handler, err error)
}

type controller struct {
	kubeClientFty      kubeclient.Factory
	clusterMgr         manager.ClusterManager
	applicationMgr     manager.ApplicationManager
	templateReleaseMgr manager.TemplateReleaseManager
	envMgr             manager.EnvironmentManager
	envRegionMgr       manager.EnvironmentRegionManager
	regionMgr          manager.RegionManager
	clusterGitRepo     gitrepo.ClusterGitRepo
}

var _ Controller = (*controller)(nil)

func NewController(param *param.Param) Controller {
	return &controller{
		kubeClientFty:      kubeclient.Fty,
		clusterMgr:         param.ClusterMgr,
		applicationMgr:     param.ApplicationManager,
		templateReleaseMgr: param.TemplateReleaseManager,
		envMgr:             param.EnvMgr,
		envRegionMgr:       param.EnvRegionMgr,
		regionMgr:          param.RegionMgr,
		clusterGitRepo:     param.ClusterGitRepo,
	}
}

func (c *controller) GetTerminalID(ctx context.Context, clusterID uint, podName,
	containerName string) (*SessionIDResp, error) {
	const op = "terminal: get terminal id"
	defer wlog.Start(ctx, op).StopPrint()
	sessionID := &SessionIDResp{}
	randomID, err := genRandomID()
	if err != nil {
		return nil, err
	}

	sessionID.ID = genSessionID(clusterID, podName, containerName, randomID)
	return sessionID, nil
}

func (c *controller) GetSockJSHandler(ctx context.Context, sessionID string) (http.Handler, error) {
	const op = "terminal: get sockjs handler"
	defer wlog.Start(ctx, op).StopPrint()

	clusterID, podName, containerName, randomID, err := parseSessionID(sessionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	kubeConfig, kubeClient, err := c.kubeClientFty.GetByK8SServer(regionEntity.Server, regionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}

	ref := ContainerRef{
		Environment: cluster.EnvironmentName,
		Cluster:     cluster.Name,
		ClusterID:   cluster.ID,
		Namespace:   envValue.Namespace,
		Pod:         podName,
		Container:   containerName,
		RandomID:    randomID,
	}

	terminalSessions.Set(ref.String(), Session{
		id:       ref.String(),
		bound:    make(chan error),
		sizeChan: make(chan remotecommand.TerminalSize),
	})

	go WaitForTerminal(kubeClient.Basic, kubeConfig, ref)

	handler := sockjs.NewHandler("/apis/front/v1", sockjs.DefaultOptions, handleTerminalSession)
	return handler, nil
}

func (c *controller) CreateShell(ctx context.Context, clusterID uint, podName,
	containerName string) (string, http.Handler, error) {
	const op = "terminal controller: create shell"
	defer wlog.Start(ctx, op).StopPrint()

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return "", nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return "", nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return "", nil, err
	}

	kubeConfig, kubeClient, err := c.kubeClientFty.GetByK8SServer(regionEntity.Server, regionEntity.Certificate)
	if err != nil {
		return "", nil, err
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return "", nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return "", nil, err
	}

	// Generate a random number as the session id
	randomID, err := genRandomID()
	if err != nil {
		return "", nil, err
	}

	ref := ContainerRef{
		Environment: cluster.EnvironmentName,
		Cluster:     cluster.Name,
		ClusterID:   cluster.ID,
		Namespace:   envValue.Namespace,
		Pod:         podName,
		Container:   containerName,
		RandomID:    randomID,
	}

	terminalSessions.Set(ref.String(), Session{
		id:       ref.String(),
		bound:    make(chan error),
		sizeChan: make(chan remotecommand.TerminalSize),
	})

	handler := sockjs.NewHandler("/apis/core/v1", sockjs.DefaultOptions, handleShellSession(ctx, ref.String()))

	go WaitForTerminal(kubeClient.Basic, kubeConfig, ref)
	return randomID, handler, nil
}

func genRandomID() (string, error) {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return "", perror.Wrap(herrors.ErrGenerateRandomID, err.Error())
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}

func genSessionID(clusterID uint, podName, containerName, randomID string) string {
	return fmt.Sprintf("%d:%s:%s:%s", clusterID, podName, containerName, randomID)
}

func parseSessionID(sessionID string) (clusterID uint, podName string, containerName string, randomID string,
	err error) {
	parts := strings.Split(sessionID, ":")
	if len(parts) != 4 {
		return clusterID, podName, containerName, randomID, fmt.Errorf("invalid sessionID")
	}

	clusterIDInt64, err := strconv.ParseUint(parts[0], 10, 0)
	if err != nil {
		return clusterID, podName, containerName, randomID, err
	}
	clusterID = uint(clusterIDInt64)
	podName = parts[1]
	containerName = parts[2]
	randomID = parts[3]
	return clusterID, podName, containerName, randomID, nil
}
