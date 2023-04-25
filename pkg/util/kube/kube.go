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

package kube

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"

	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	kubeerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/cmd/exec"
)

const DefaultEventsLimit = 100

const (
	EnvK8sClientQPS   = "HORIZON_K8S_CLIENT_QPS"
	EnvK8sClientBurst = "HORIZON_K8S_CLIENT_BURST"
)

var (
	K8sClientConfigQPS   float32 = 50
	K8sClientConfigBurst         = 100
)

func init() {
	if envQPS := os.Getenv(EnvK8sClientQPS); envQPS != "" {
		if qps, err := strconv.ParseFloat(envQPS, 32); err == nil {
			K8sClientConfigQPS = float32(qps)
		}
	}
	if envBurst := os.Getenv(EnvK8sClientBurst); envBurst != "" {
		if burst, err := strconv.Atoi(envBurst); err == nil {
			K8sClientConfigBurst = burst
		}
	}
}

// GetEvents Returns a map. key is the basic information of Pod, name-uid-namespace
func GetEvents(ctx context.Context, kubeClientset kubernetes.Interface,
	namespace string) (_ map[string][]*v1.Event, err error) {
	const op = "kube: get multi pod events from k8s "
	defer wlog.Start(ctx, op).StopPrint()

	eventsMapper := make(map[string][]*v1.Event)
	events, err := kubeClientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		Limit: DefaultEventsLimit,
		FieldSelector: fields.SelectorFromSet(map[string]string{
			"involvedObject.kind": "Pod",
		}).String(),
	})
	if err != nil {
		return nil, herrors.NewErrListFailed(herrors.PodEventInK8S, err.Error())
	}
	for i := range events.Items {
		name := events.Items[i].InvolvedObject.Name
		uid := events.Items[i].InvolvedObject.UID
		namespace := events.Items[i].InvolvedObject.Namespace
		key := fmt.Sprintf("%v-%v-%v", name, uid, namespace)
		eventsMapper[key] = append(eventsMapper[key], &events.Items[i])
	}

	return eventsMapper, nil
}

// GetPodEvents returns event list of a pod, notice that it will return events from oldest to latest by
// DefaultEventsLimit
func GetPodEvents(ctx context.Context, kubeClientset kubernetes.Interface, namespace, pod string) (_ []v1.Event,
	err error) {
	const op = "kube: get single pod events from k8s "
	defer wlog.Start(ctx, op).StopPrint()

	events, err := kubeClientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		Limit: DefaultEventsLimit,
		FieldSelector: fields.SelectorFromSet(map[string]string{
			"involvedObject.kind": "Pod",
			"involvedObject.name": pod,
		}).String(),
	})
	if err != nil {
		return nil, herrors.NewErrListFailed(herrors.PodEventInK8S, err.Error())
	}

	return events.Items, nil
}

func GetPods(ctx context.Context, kubeClientset kubernetes.Interface,
	namespace, labelSelector string) (_ []v1.Pod, err error) {
	const op = "kube: get pods from k8s "
	defer wlog.Start(ctx, op).StopPrint()

	pods, err := kubeClientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		if kubeerror.IsNotFound(err) {
			return nil, herrors.NewErrNotFound(herrors.PodsInK8S, err.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.PodsInK8S, err.Error())
	}
	return pods.Items, nil
}

func GetPod(ctx context.Context, kubeClientset kubernetes.Interface, namespace, podName string) (_ *v1.Pod, err error) {
	pod, err := kubeClientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if kubeerror.IsNotFound(err) {
			return nil, herrors.NewErrNotFound(herrors.PodsInK8S, err.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.PodsInK8S, err.Error())
	}
	return pod, nil
}

func DeletePods(ctx context.Context, kubeClientset kubernetes.Interface, namespace string, pod string) (err error) {
	err = kubeClientset.CoreV1().Pods(namespace).Delete(ctx, pod, metav1.DeleteOptions{})
	if err != nil {
		if kubeerror.IsNotFound(err) {
			return herrors.NewErrNotFound(herrors.PodsInK8S, err.Error())
		}
		return herrors.NewErrDeleteFailed(herrors.PodsInK8S, err.Error())
	}
	return nil
}

func BuildClient(kubeconfig string) (*rest.Config, kubernetes.Interface, error) {
	var restConfig *rest.Config
	var err error
	if len(kubeconfig) > 0 {
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, nil, herrors.NewErrGetFailed(herrors.KubeConfigInK8S, err.Error())
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, herrors.NewErrGetFailed(herrors.KubeConfigInK8S, err.Error())
		}
	}

	groupVersion := &schema.GroupVersion{Group: "", Version: "v1"}
	restConfig.GroupVersion = groupVersion
	restConfig.APIPath = "/api"
	restConfig.ContentType = runtime.ContentTypeJSON
	restConfig.NegotiatedSerializer = scheme.Codecs
	restConfig.QPS = K8sClientConfigQPS
	restConfig.Burst = K8sClientConfigBurst
	log.Infof(context.Background(), "BuildClient set kube qps: %v, burst: %v", K8sClientConfigQPS,
		K8sClientConfigBurst)

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, herrors.NewErrCreateFailed(herrors.KubeConfigInK8S, err.Error())
	}
	return restConfig, k8sClient, nil
}

type Client struct {
	Basic   kubernetes.Interface
	Dynamic dynamic.Interface
}

// BuildClientFromContent build client from k8s kubeconfig content, not file path
func BuildClientFromContent(kubeconfigContent string) (*rest.Config, *Client, error) {
	var restConfig *rest.Config
	var err error
	if len(kubeconfigContent) > 0 {
		clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(kubeconfigContent))
		if err != nil {
			return nil, nil, herrors.NewErrGetFailed(herrors.KubeConfigInK8S, err.Error())
		}
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, nil, herrors.NewErrGetFailed(herrors.KubeConfigInK8S, err.Error())
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, herrors.NewErrGetFailed(herrors.KubeConfigInK8S, err.Error())
		}
	}

	groupVersion := &schema.GroupVersion{Group: "", Version: "v1"}
	restConfig.GroupVersion = groupVersion
	restConfig.APIPath = "/api"
	restConfig.ContentType = runtime.ContentTypeJSON
	restConfig.NegotiatedSerializer = scheme.Codecs
	restConfig.QPS = K8sClientConfigQPS
	restConfig.Burst = K8sClientConfigBurst
	log.Infof(context.Background(), "BuildClientFromContent set kube qps: %v, burst: %v", K8sClientConfigQPS,
		K8sClientConfigBurst)

	basicClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, herrors.NewErrCreateFailed(herrors.KubeConfigInK8S, err.Error())
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, herrors.NewErrCreateFailed(herrors.KubeConfigInK8S, err.Error())
	}

	kubeClient := &Client{
		Basic:   basicClient,
		Dynamic: dynamicClient,
	}

	return restConfig, kubeClient, nil
}

type ContainerRef struct {
	Config        *rest.Config
	KubeClientset kubernetes.Interface
	Namespace     string
	Pod           string
	Container     string
}

func Exec(ctx context.Context, c ContainerRef,
	command []string, executor exec.RemoteExecutor) (stdout string, stderr string, err error) {
	const op = "kube: execute command in pod"
	defer wlog.Start(ctx, op).StopPrint()

	out := bytes.NewBuffer([]byte{})
	errOut := bytes.NewBuffer([]byte{})
	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			IOStreams: genericclioptions.IOStreams{
				Out:    out,
				ErrOut: errOut,
			},
			Namespace:     c.Namespace,
			PodName:       c.Pod,
			ContainerName: c.Container,
		},

		Config:    c.Config,
		PodClient: c.KubeClientset.CoreV1(),

		Command:  command,
		Executor: &exec.DefaultRemoteExecutor{},
	}
	if executor != nil {
		options.Executor = executor
	}

	// use raw error
	if err := options.Validate(); err != nil {
		return "", "", perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	err = options.Run()
	if err != nil {
		return out.String(), errOut.String(), perror.Wrap(herrors.ErrKubeExecFailed, err.Error())
	}
	return out.String(), errOut.String(), nil
}

func GetReplicaSets(ctx context.Context, kubeClientset kubernetes.Interface,
	namespace, labelSelector string) (_ []appsv1.ReplicaSet, err error) {
	const op = "get replicaSet list from k8s "
	defer wlog.Start(ctx, op).StopPrint()

	replicaSetList, err := kubeClientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, herrors.NewErrListFailed(herrors.PodsInK8S, err.Error())
	}
	return replicaSetList.Items, nil
}

func GetDeploymentList(ctx context.Context, kubeClientset kubernetes.Interface,
	namespace, labelSelector string) (_ []appsv1.Deployment, err error) {
	const op = "get deployments from k8s "
	defer wlog.Start(ctx, op).StopPrint()

	deploymentList, err := kubeClientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, herrors.NewErrListFailed(herrors.DeploymentInK8S, err.Error())
	}
	return deploymentList.Items, nil
}
