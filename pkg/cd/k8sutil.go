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

package cd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	applicationV1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	eventmanager "github.com/horizoncd/horizon/pkg/event/manager"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	"github.com/horizoncd/horizon/pkg/regioninformers"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"github.com/horizoncd/horizon/pkg/workload"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
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
	informerFactories *regioninformers.RegionInformers
	eventMgr          eventmanager.Manager
}

func NewK8sUtil(factories *regioninformers.RegionInformers, mgr eventmanager.Manager) K8sUtil {
	return &util{
		informerFactories: factories,
		eventMgr:          mgr,
	}
}

func (e *util) DeletePods(ctx context.Context,
	params *DeletePodsParams) (map[string]OperationResult, error) {
	result := map[string]OperationResult{}

	_ = e.informerFactories.GetClientSet(params.RegionEntity.ID, func(clientset kubernetes.Interface) error {
		for _, pod := range params.Pods {
			err := kube.DeletePods(ctx, clientset, params.Namespace, pod)
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
		return nil
	})

	return result, nil
}

func (e *util) GetContainerLog(ctx context.Context, params *GetContainerLogParams) (<-chan string, error) {
	var logC = make(chan string)
	err := e.informerFactories.GetClientSet(params.RegionEntity.ID, func(clientset kubernetes.Interface) error {
		podLogRequest := clientset.CoreV1().Pods(params.Namespace).GetLogs(params.Pod, &corev1.PodLogOptions{
			Container:  params.Container,
			TailLines:  &params.TailLines,
			Timestamps: true,
		})
		stream, err := podLogRequest.Stream(context.TODO())
		if err != nil {
			return herrors.NewErrGetFailed(herrors.PodLogsInK8S, err.Error())
		}

		go func() {
			defer stream.Close()
			defer close(logC)
			parseLogsStream(stream, logC)
		}()
		return nil
	})

	if err != nil {
		return nil, err
	}
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
	err = e.informerFactories.GetDynamicClientSet(params.RegionEntity.ID, func(clientset dynamic.Interface) error {
		un, err := clientset.Resource(params.GVR).Namespace(params.Namespace).
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

		un, err = clientset.Resource(params.GVR).Namespace(params.Namespace).
			Update(ctx, un, metav1.UpdateOptions{})
		log.Debugf(ctx, "update %s(%s) with %s: %v", params.ResourceName,
			params.GVR.String(), params.Action, un)
		if err != nil {
			return herrors.NewErrUpdateFailed(herrors.ResourceInK8S,
				fmt.Sprintf("failed to update gvr(%s), ns(%s), name(%s)",
					params.GVR.String(), params.Namespace, un.GetName()))
		}
		if err != nil {
			bts, err := json.Marshal(map[string]interface{}{
				"action":       params.Action,
				"gvr":          params.GVR.String(),
				"resourceName": params.Namespace,
			})
			if err != nil {
				log.Warningf(ctx, "failed to marshal event extra, err: %s", err.Error())
			}
			extra := string(bts)
			if _, err = e.eventMgr.CreateEvent(ctx, &eventmodels.Event{
				EventSummary: eventmodels.EventSummary{
					ResourceType: common.ResourceApplication,
					EventType:    eventmodels.ClusterRestarted,
					ResourceID:   params.ClusterID,
					Extra:        &extra,
				},
			}); err != nil {
				log.Warningf(ctx, "failed to create event, err: %s", err.Error())
			}
		}
		return err
	})

	return err
}

func (e *util) GetPodContainers(ctx context.Context,
	params *GetPodParams) (containers []ContainerDetail, err error) {
	pod, err := e.GetPod(ctx, params)

	if err != nil {
		return nil, err
	}

	return extractContainerDetail(pod), nil
}

func (e *util) GetPod(ctx context.Context,
	params *GetPodParams) (pod *corev1.Pod, err error) {
	err = e.informerFactories.GetClientSet(params.RegionEntity.ID, func(clientset kubernetes.Interface) error {
		pod, err = kube.GetPod(ctx, clientset, params.Namespace, params.Pod)
		return err
	})

	if err != nil {
		return nil, err
	}
	return pod, nil
}

func (e *util) Exec(ctx context.Context, params *ExecParams) (resp map[string]ExecResp, err error) {
	const op = "cd: shell exec"
	defer wlog.Start(ctx, op).StopPrint()

	_ = e.informerFactories.GetClientSet(params.RegionEntity.ID, func(clientset kubernetes.Interface) error {
		containers := make([]kube.ContainerRef, 0)
		for _, pod := range params.PodList {
			containers = append(containers, kube.ContainerRef{
				KubeClientset: clientset,
				Namespace:     params.Namespace,
				Pod:           pod,
				Container:     params.Cluster,
			})
		}
		resp = executeCommandInPods(ctx, containers, params.Commands, nil)
		return nil
	})

	return resp, nil
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
