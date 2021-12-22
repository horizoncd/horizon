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
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
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
	CreateShell(ctx context.Context, clusterID uint, podName, containerName string) (string, http.Handler, error)
}

type controller struct {
	kubeClientFty  kubeclient.Factory
	clusterMgr     clustermanager.Manager
	applicationMgr applicationmanager.Manager
	envMgr         envmanager.EnvironmentRegionManager
	regionMgr      regionmanager.Manager
	clusterGitRepo gitrepo.ClusterGitRepo
}

var _ Controller = (*controller)(nil)

func NewController(clusterGitRepo gitrepo.ClusterGitRepo) Controller {
	return &controller{
		kubeClientFty:  kubeclient.Fty,
		clusterMgr:     clustermanager.Mgr,
		applicationMgr: applicationmanager.Mgr,
		envMgr:         envmanager.Mgr,
		regionMgr:      regionmanager.Mgr,
		clusterGitRepo: clusterGitRepo,
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

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}

	ref := ContainerRef{
		Environment: er.EnvironmentName,
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

	// 1. 获取各类关联资源
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return "", nil, errors.E(op, err)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return "", nil, errors.E(op, err)
	}

	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return "", nil, errors.E(op, err)
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, er.RegionName)
	if err != nil {
		return "", nil, errors.E(op, err)
	}

	kubeConfig, kubeClient, err := c.kubeClientFty.GetByK8SServer(ctx, regionEntity.K8SCluster.Server)
	if err != nil {
		return "", nil, errors.E(op, err)
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return "", nil, errors.E(op, err)
	}

	// 2. 生成随机数，作为session id
	randomID, err := genRandomID()
	if err != nil {
		return "", nil, err
	}

	ref := ContainerRef{
		Environment: er.EnvironmentName,
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

	// 3. 初始化sockJS处理函数
	handler := sockjs.NewHandler("/apis/core/v1", sockjs.DefaultOptions, handleShellSession(ref.String()))

	// 4. 启动协程，等待客户端发送绑定请求，连接容器
	go WaitForTerminal(kubeClient.Basic, kubeConfig, ref)
	return randomID, handler, nil
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
