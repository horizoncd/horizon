package terminal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	"g.hz.netease.com/horizon/pkg/cluster/kubeclient"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"k8s.io/client-go/tools/remotecommand"

	"gopkg.in/igm/sockjs-go.v3/sockjs"
)

type Controller interface {
	GetTerminalID(ctx context.Context, clusterID uint, podName, containerName string) (*SessionIDResp, error)
	GetSockJSHandler(ctx context.Context, sessionID string) (http.Handler, error)
}

type controller struct {
	kubeClientFty  kubeclient.Factory
	clusterMgr     clustermanager.Manager
	applicationMgr applicationmanager.Manager
	envMgr         envmanager.EnvironmentRegionManager
	regionMgr      regionmanager.Manager
}

var _ Controller = (*controller)(nil)

func NewController() Controller {
	return &controller{
		kubeClientFty:  kubeclient.Fty,
		clusterMgr:     clustermanager.Mgr,
		applicationMgr: applicationmanager.Mgr,
		envMgr:         envmanager.Mgr,
		regionMgr:      regionmanager.Mgr,
	}
}

func (c *controller) GetTerminalID(ctx context.Context, clusterID uint, podName,
	containerName string) (*SessionIDResp, error) {
	sessionID := &SessionIDResp{}
	randomID, err := genRandomID()
	if err != nil {
		return nil, err
	}

	sessionID.ID = genSessionID(clusterID, podName, containerName, randomID)
	return sessionID, nil
}

func (c *controller) GetSockJSHandler(ctx context.Context, sessionID string) (http.Handler, error) {
	const op = "terminal: sockjs"
	clusterID, podName, containerName, randomID, err := parseSessionID(sessionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, er.RegionName)
	if err != nil {
		return nil, errors.E(op, err)
	}

	kubeConfig, kubeClient, err := c.kubeClientFty.GetByK8SServer(ctx, regionEntity.K8SCluster.Server)
	if err != nil {
		return nil, errors.E(op, err)
	}

	ref := ContainerRef{
		Environment: er.EnvironmentName,
		Cluster:     cluster.Name,
		ClusterID:   cluster.ID,
		Namespace:   fmt.Sprintf("%v-%v", er.EnvironmentName, application.GroupID),
		Pod:         podName,
		Container:   containerName,
		RandomID:    randomID,
	}

	terminalSessions.Set(ref.String(), Session{
		id:       ref.String(),
		bound:    make(chan error),
		sizeChan: make(chan remotecommand.TerminalSize),
	})

	go WaitForTerminal(kubeClient, kubeConfig, ref)

	handler := sockjs.NewHandler("/apis/front/v1", sockjs.DefaultOptions, handleTerminalSession)
	return handler, nil
}

func genRandomID() (string, error) {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
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
