package cd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	applicationV1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/kubeclient"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"github.com/horizoncd/horizon/pkg/workload"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/cd/k8sutil_mock.go -package=mock_cd
type K8sUtil interface {
	Exec(ctx context.Context, params *ExecParams) (map[string]ExecResp, error)
	DeletePods(ctx context.Context, params *DeletePodsParams) (map[string]OperationResult, error)
	ExecuteAction(ctx context.Context, params *ExecuteActionParams) error
	GetPodContainers(ctx context.Context, params *GetPodParams) ([]ContainerDetail, error)
	GetPod(ctx context.Context, params *GetPodParams) (*corev1.Pod, error)
	GetContainerLog(ctx context.Context, params *GetContainerLogParams) (<-chan string, error)
}

type util struct {
	kubeClientFty kubeclient.Factory
}

func NewK8sUtil() K8sUtil {
	return &util{
		kubeClientFty: kubeclient.Fty,
	}
}

func (e *util) DeletePods(ctx context.Context,
	params *DeletePodsParams) (map[string]OperationResult, error) {
	result := map[string]OperationResult{}
	_, kubeClient, err := e.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return result, err
	}

	for _, pod := range params.Pods {
		err = kube.DeletePods(ctx, kubeClient.Basic, params.Namespace, pod)
		if err != nil {
			result[pod] = OperationResult{
				Result: false,
				Error:  err,
			}
			continue
		}
		result[pod] = OperationResult{
			Result: true,
		}
	}

	return result, nil
}

func (e *util) GetContainerLog(ctx context.Context, params *GetContainerLogParams) (<-chan string, error) {
	_, kubeClient, err := e.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	podLogRequest := kubeClient.Basic.CoreV1().Pods(params.Namespace).GetLogs(params.Pod, &corev1.PodLogOptions{
		Container:  params.Container,
		TailLines:  &params.TailLines,
		Timestamps: true,
	})
	stream, err := podLogRequest.Stream(context.TODO())
	if err != nil {
		return nil, herrors.NewErrGetFailed(herrors.PodLogsInK8S, err.Error())
	}

	logC := make(chan string)

	go func() {
		defer stream.Close()
		defer close(logC)
		parseLogsStream(stream, logC)
	}()

	return logC, nil
}

func parseLogsStream(stream io.ReadCloser, ch chan string) {
	bufReader := bufio.NewReader(stream)
	eof := false
	for !eof {
		line, err := bufReader.ReadString('\n')
		if err == io.EOF {
			eof = true
			if line == "" {
				break
			}
		} else if err != nil && err != io.EOF {
			ch <- fmt.Sprintf("%v\n", err)
			break
		}

		line = strings.TrimSpace(line)
		parts := strings.Split(line, " ")
		timeStampStr := parts[0]
		timeStamp, err := time.Parse(time.RFC3339Nano, timeStampStr)
		if err != nil {
			ch <- fmt.Sprintf("%v\n", err)
			break
		}

		lines := strings.Join(parts[1:], " ")
		for _, line := range strings.Split(lines, "\r") {
			ch <- fmt.Sprintf("[%s] %s\n", timeStamp.Format(time.RFC3339), line)
		}
	}
}

func (e *util) ExecuteAction(ctx context.Context,
	params *ExecuteActionParams) (err error) {
	_, kubeClient, err := e.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return herrors.NewErrGetFailed(herrors.K8SClient,
			fmt.Sprintf("failed to get kube client for %s", params.RegionEntity.Server))
	}
	un, err := kubeClient.Dynamic.Resource(params.GVR).Namespace(params.Namespace).
		Get(ctx, params.ResourceName, metav1.GetOptions{})
	if err != nil {
		return herrors.NewErrGetFailed(herrors.ResourceInK8S,
			fmt.Sprintf("failed to get %s(%s)", params.ResourceName, params.GVR.String()))
	}

	workload.LoopAbilities(func(workload workload.Workload) bool {
		if workload.MatchGK(un.GroupVersionKind().GroupKind()) {
			un, err = workload.Action(params.Action, un)
			return false
		}
		return true
	})
	if err != nil {
		return perror.Wrapf(err, "failed to execute '%s' for %s(%s)",
			params.Action, params.ResourceName, params.GVR.String())
	}

	un, err = kubeClient.Dynamic.Resource(params.GVR).Namespace(params.Namespace).
		Update(ctx, un, metav1.UpdateOptions{})
	log.Debugf(ctx, "update %s(%s) with %s: %v", params.ResourceName,
		params.GVR.String(), params.Action, un)

	return err
}

func (e *util) GetPodContainers(ctx context.Context,
	params *GetPodParams) (containers []ContainerDetail, err error) {
	_, kubeClient, err := e.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	pod, err := kube.GetPod(ctx, kubeClient.Basic, params.Namespace, params.Pod)
	if err != nil {
		return nil, err
	}

	return extractContainerDetail(pod), nil
}

func (e *util) GetPod(ctx context.Context,
	params *GetPodParams) (pod *corev1.Pod, err error) {
	_, kubeClient, err := e.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	pod, err = kube.GetPod(ctx, kubeClient.Basic, params.Namespace, params.Pod)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (e *util) Exec(ctx context.Context, params *ExecParams) (_ map[string]ExecResp, err error) {
	const op = "cd: shell exec"
	defer wlog.Start(ctx, op).StopPrint()

	config, kubeClient, err := e.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}
	containers := make([]kube.ContainerRef, 0)
	for _, pod := range params.PodList {
		containers = append(containers, kube.ContainerRef{
			Config:        config,
			KubeClientset: kubeClient.Basic,
			Namespace:     params.Namespace,
			Pod:           pod,
			Container:     params.Cluster,
		})
	}

	return executeCommandInPods(ctx, containers, params.Commands, nil), nil
}

// TraverseOperator stops if result is false
type TraverseOperator func(node *ResourceTreeNode) bool

// traverseResourceTree traverses tree by dfs
func (c *cd) traverseResourceTree(resourceTree *applicationV1alpha1.ApplicationTree,
	operators ...TraverseOperator) {
	m := make(map[string]*applicationV1alpha1.ResourceNode)
	for i, node := range resourceTree.Nodes {
		m[node.UID] = &resourceTree.Nodes[i]
	}

	visited := make(map[string]*ResourceTreeNode)
	roots := make([]*ResourceTreeNode, 0, 4)
	for i := range resourceTree.Nodes {
		tree := (*ResourceTreeNode)(nil)
		currentNode := &resourceTree.Nodes[i]
		for {
			if _, ok := visited[currentNode.UID]; ok {
				parent := visited[currentNode.UID]
				if tree != nil {
					parent.children = append(parent.children, tree)
				}
				break
			}

			t := &ResourceTreeNode{
				ResourceNode: currentNode,
			}
			if tree != nil {
				t.children = append(t.children, tree)
			}
			tree = t
			visited[currentNode.UID] = tree

			if currentNode.ParentRefs != nil {
				currentNode = m[currentNode.ParentRefs[0].UID]
			} else {
				roots = append(roots, tree)
				break
			}
		}
	}

	var dfs func(node *ResourceTreeNode, operator TraverseOperator)
	dfs = func(node *ResourceTreeNode, operator TraverseOperator) {
		if node == nil {
			return
		}
		if !operator(node) || node.children == nil {
			return
		}
		for _, child := range node.children {
			dfs(child, operator)
		}
	}

	for _, operator := range operators {
		for _, root := range roots {
			dfs(root, operator)
		}
	}
}
