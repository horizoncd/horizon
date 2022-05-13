package terminal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	herrors "g.hz.netease.com/horizon/core/errors"
	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	"g.hz.netease.com/horizon/pkg/cluster/kubeclient"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	envregionmanager "g.hz.netease.com/horizon/pkg/environmentregion/manager"
	perror "g.hz.netease.com/horizon/pkg/errors"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
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
	kubeClientFty  kubeclient.Factory
	clusterMgr     clustermanager.Manager
	applicationMgr applicationmanager.Manager
	envMgr         envmanager.Manager
	envRegionMgr   envregionmanager.Manager
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
		envRegionMgr:   envregionmanager.Mgr,
		regionMgr:      regionmanager.Mgr,
		clusterGitRepo: clusterGitRepo,
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

	er, err := c.envRegionMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, er.RegionName)
	if err != nil {
		return nil, err
	}

	kubeConfig, kubeClient, err := c.kubeClientFty.GetByK8SServer(regionEntity.Server, regionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, err
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
	defer wlog.Start(ctx, op).StopPrint()

	// 1. 获取各类关联资源
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return "", nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return "", nil, err
	}

	er, err := c.envRegionMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return "", nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, er.RegionName)
	if err != nil {
		return "", nil, err
	}

	kubeConfig, kubeClient, err := c.kubeClientFty.GetByK8SServer(regionEntity.Server, regionEntity.Certificate)
	if err != nil {
		return "", nil, err
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return "", nil, err
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
	handler := sockjs.NewHandler("/apis/core/v1", sockjs.DefaultOptions, handleShellSession(ctx, ref.String()))

	// 4. 启动协程，等待客户端发送绑定请求，连接容器
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
