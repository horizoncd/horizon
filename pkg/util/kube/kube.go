package kube

import (
	"bytes"
	"context"
	"fmt"

	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
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

// GetEvents 返回一个map。key是Pod的基本信息，name-uid-namespace
func GetEvents(ctx context.Context, kubeClientset kubernetes.Interface,
	namespace string) (_ map[string][]*v1.Event, err error) {
	const op = "kube: get multi pod events from k8s "
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	eventsMapper := make(map[string][]*v1.Event)
	events, err := kubeClientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		Limit: DefaultEventsLimit,
		FieldSelector: fields.SelectorFromSet(map[string]string{
			"involvedObject.kind": "Pod",
		}).String(),
	})
	if err != nil {
		return nil, errors.E(op, err)
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
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	events, err := kubeClientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		Limit: DefaultEventsLimit,
		FieldSelector: fields.SelectorFromSet(map[string]string{
			"involvedObject.kind": "Pod",
			"involvedObject.name": pod,
		}).String(),
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	return events.Items, nil
}

func GetPods(ctx context.Context, kubeClientset kubernetes.Interface,
	namespace, labelSelector string) (_ []v1.Pod, err error) {
	const op = "kube: get pods from k8s "
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	pods, err := kubeClientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return pods.Items, nil
}

// BuildClient 根据传入的kubeconfig地址生成对应的k8sClient
// kubeconfig表示kubeconfig文件的地址。如果该地址为空，则默认使用InClusterConfig，即本Pod所在集群的config
func BuildClient(kubeconfig string) (*rest.Config, kubernetes.Interface, error) {
	var restConfig *rest.Config
	var err error
	if len(kubeconfig) > 0 {
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, nil, err
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, err
		}
	}

	groupVersion := &schema.GroupVersion{Group: "", Version: "v1"}
	restConfig.GroupVersion = groupVersion
	restConfig.APIPath = "/api"
	restConfig.ContentType = runtime.ContentTypeJSON
	restConfig.NegotiatedSerializer = scheme.Codecs

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
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
			return nil, nil, err
		}
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, nil, err
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, err
		}
	}

	groupVersion := &schema.GroupVersion{Group: "", Version: "v1"}
	restConfig.GroupVersion = groupVersion
	restConfig.APIPath = "/api"
	restConfig.ContentType = runtime.ContentTypeJSON
	restConfig.NegotiatedSerializer = scheme.Codecs

	basicClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
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
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

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
		return "", "", err
	}

	err = options.Run()
	return out.String(), errOut.String(), err
}

func GetReplicaSets(ctx context.Context, kubeClientset kubernetes.Interface,
	namespace, labelSelector string) (_ []appsv1.ReplicaSet, err error) {
	const op = "get replicaSet list from k8s "
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	replicaSetList, err := kubeClientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return replicaSetList.Items, nil
}

func GetDeploymentList(ctx context.Context, kubeClientset kubernetes.Interface,
	namespace, labelSelector string) (_ []appsv1.Deployment, err error) {
	const op = "get deployments from k8s "
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	deploymentList, err := kubeClientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return deploymentList.Items, nil
}
