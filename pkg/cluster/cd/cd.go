package cd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"g.hz.netease.com/horizon/pkg/cluster/cd/argocd"
	"g.hz.netease.com/horizon/pkg/cluster/common"
	"g.hz.netease.com/horizon/pkg/cluster/kubeclient"
	argocdconf "g.hz.netease.com/horizon/pkg/config/argocd"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/kube"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"

	"github.com/argoproj/gitops-engine/pkg/health"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

const (
	_deploymentRevision       = "deployment.kubernetes.io/revision"
	DeploymentPodTemplateHash = "pod-template-hash"
	_rolloutRevision          = "rollout.argoproj.io/revision"
	RolloutPodTemplateHash    = "rollouts-pod-template-hash"
)

type GetClusterStateParams struct {
	Environment  string
	Cluster      string
	Namespace    string
	RegionEntity *regionmodels.RegionEntity
}

type DeployClusterParams struct {
	Environment   string
	Cluster       string
	GitRepoSSHURL string
	ValueFiles    []string
	RegionEntity  *regionmodels.RegionEntity
	Namespace     string
	Revision      string
}

type CD interface {
	DeployCluster(ctx context.Context, params *DeployClusterParams) error
	// GetClusterState get cluster state in cd system
	GetClusterState(ctx context.Context, params *GetClusterStateParams) (*ClusterState, error)
}

type cd struct {
	kubeClientFty kubeclient.Factory
	factory       argocd.Factory
}

func NewCD(argoCDMapper argocdconf.Mapper) CD {
	return &cd{
		kubeClientFty: kubeclient.Fty,
		factory:       argocd.NewFactory(argoCDMapper),
	}
}

func (c *cd) DeployCluster(ctx context.Context, params *DeployClusterParams) (err error) {
	const op = "cd: deploy cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return errors.E(op, err)
	}

	// if argo application is not exists, create it first
	if _, err = argo.GetApplication(ctx, params.Cluster); err != nil {
		if errors.Status(err) != http.StatusNotFound {
			return errors.E(op, err)
		}
		var argoApplication = argocd.AssembleArgoApplication(params.Cluster, params.Namespace,
			params.GitRepoSSHURL, params.RegionEntity.K8SCluster.Server, params.ValueFiles)

		manifest, err := json.Marshal(argoApplication)
		if err != nil {
			return errors.E(op, err)
		}

		if err := argo.CreateApplication(ctx, manifest); err != nil {
			return errors.E(op, err)
		}
	}

	return argo.DeployApplication(ctx, params.Cluster, params.Revision)
}

// GetClusterState TODO(gjq) restructure
func (c *cd) GetClusterState(ctx context.Context,
	params *GetClusterStateParams) (clusterState *ClusterState, err error) {
	const op = "cd: get cluster status"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return nil, errors.E(op, err)
	}

	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(ctx, params.RegionEntity.K8SCluster.Server)
	if err != nil {
		return nil, errors.E(op, err)
	}

	clusterState = &ClusterState{Versions: map[string]*ClusterVersion{}}

	// get application status
	argoApp, err := argo.GetApplication(ctx, params.Cluster)
	if err != nil {
		return nil, errors.E(op, err)
	}
	// namespace = argoApp.Spec.Destination.Namespace
	clusterState.Status = argoApp.Status.Health.Status
	if clusterState.Status == "" {
		return nil, errors.E(op, http.StatusNotFound, "clusterState.Status == ''")
	}
	if clusterState.Status == health.HealthStatusUnknown {
		clusterState.Status = health.HealthStatusDegraded
	} else if clusterState.Status == health.HealthStatusMissing {
		clusterState.Status = health.HealthStatusProgressing
	}

	var latestReplicaSet *appsv1.ReplicaSet
	labelSelector := fields.ParseSelectorOrDie(fmt.Sprintf("%v=%v",
		common.ClusterLabelKey, params.Cluster))
	rss, err := kube.GetReplicaSets(ctx, kubeClient, params.Namespace, labelSelector.String())
	if err != nil {
		return nil, err
	} else if len(rss) == 0 {
		return nil, errors.E(op, http.StatusNotFound, "ReplicaSet instance not found")
	}

	for i := range rss {
		rs := &rss[i]

		if latestReplicaSet == nil || CompareRevision(ctx, rs, latestReplicaSet) {
			latestReplicaSet = rs
		}
		_, hash := getPodTemplateHash(rs)
		clusterState.Versions[hash] = &ClusterVersion{
			Pods:     map[string]*ClusterPod{},
			Revision: getRevision(rs),
		}
	}

	// set revision, podTemplateHash
	clusterState.PodTemplateHashKey, clusterState.PodTemplateHash = getPodTemplateHash(latestReplicaSet)
	clusterState.Revision = getRevision(latestReplicaSet)

	if clusterState.PodTemplateHash == "" {
		return nil, errors.E(op, http.StatusNotFound, "clusterState.PodTemplateHash == ''")
	}

	// 从目前的配置来看，该 if 分支表示负载类型是 serverless 应用
	if clusterState.PodTemplateHashKey == DeploymentPodTemplateHash {
		labelSelector := fields.ParseSelectorOrDie(
			fmt.Sprintf("%v=%v", common.ClusterLabelKey, params.Cluster))
		// serverless 应用会有多个 Deployment 对象
		deploymentList, err := kube.GetDeploymentList(ctx, kubeClient, params.Namespace, labelSelector.String())
		if err != nil {
			return nil, errors.E(op, err)
		}
		for i := range deploymentList {
			deployment := &deploymentList[i]
			// Borrowed at kubernetes/kubectl/rollout_status.go
			if deployment.Generation <= deployment.Status.ObservedGeneration {
				cond := getDeploymentCondition(deployment.Status, appsv1.DeploymentProgressing)
				if cond != nil && cond.Reason == "ProgressDeadlineExceeded" {
					// 默认情况下，如果 Deployment 在十分钟之内，未能完成滚动更新，
					// 则 Deployment 的健康状态是 HealthStatusDegraded.
					clusterState.Status = health.HealthStatusDegraded
				}
			} else {
				// 如果 Deployment 有更新，而 Deployment Controller 尚未处理，
				// 则 Deployment 的健康状态是 HealthStatusProgressing
				clusterState.Status = health.HealthStatusProgressing
			}
		}
	}

	return clusterState, nil
}

type (

	// ClusterState 集群状态信息
	ClusterState struct {
		// Status:
		// Processing(正在部署)；Healthy(部署完成)
		// Suspended(已暂停)；Degraded(已降级)
		// Missing(Rollout或Deployment还尚未部署到业务集群)
		// Unknown(集群健康评估失败，无法获悉当前的部署状态)
		Status health.HealthStatusCode `json:"status,omitempty" yaml:"status,omitempty"`

		// Replicas the actual number of replicas running in k8s
		Replicas int `json:"replicas,omitempty" yaml:"replicas,omitempty"`

		// DesiredReplicas desired replicas
		DesiredReplicas *int `json:"desiredReplicas,omitempty" yaml:"desiredReplicas,omitempty"`

		// PodTemplateHash
		PodTemplateHash string `json:"podTemplateHash,omitempty" yaml:"podTemplateHash,omitempty"`

		// PodTemplateHashKey 在 Deployment 或 Rollout 对象中的 label 的 key
		PodTemplateHashKey string `json:"podTemplateHashKey,omitempty" yaml:"podTemplateHashKey,omitempty"`

		// Revision the desired revision
		Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`

		// Versions versions detail
		// key is pod-template-hash, if equal to PodTemplateHash, the version is the desired version
		Versions map[string]*ClusterVersion `json:"versions,omitempty" yaml:"versions,omitempty"`
	}

	// ClusterVersion version information
	ClusterVersion struct {
		// Replicas the replicas of this revision
		Replicas int    `json:"replicas,omitempty" yaml:"replicas,omitempty"`
		Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`
		// Pods the pods detail of this revision, the key is pod name
		Pods map[string]*ClusterPod `json:"pods,omitempty" yaml:"pods,omitempty"`
	}

	// ClusterPod pod detail
	ClusterPod struct {
		Metadata PodMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
		Spec     PodSpec     `json:"spec,omitempty" yaml:"spec,omitempty"`
		Status   PodStatus   `json:"status,omitempty" yaml:"status,omitempty"`
	}

	PodMetadata struct {
		Namespace   string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
		Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	}

	PodSpec struct {
		NodeName       string       `json:"nodeName,omitempty" yaml:"nodeName,omitempty"`
		InitContainers []*Container `json:"initContainers,omitempty" yaml:"initContainers,omitempty"`
		Containers     []*Container `json:"containers,omitempty" yaml:"containers,omitempty"`
	}

	Container struct {
		Name  string `json:"name,omitempty" yaml:"name,omitempty"`
		Image string `json:"image,omitempty" yaml:"image,omitempty"`
	}

	PodStatus struct {
		HostIP            string             `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
		PodIP             string             `json:"podIP,omitempty" yaml:"podIP,omitempty"`
		Phase             string             `json:"phase,omitempty" yaml:"phase,omitempty"`
		Events            []Event            `json:"events,omitempty" yaml:"events,omitempty"`
		ContainerStatuses []*ContainerStatus `json:"containerStatuses,omitempty" yaml:"containerStatuses,omitempty"`
	}

	ContainerStatus struct {
		Name  string         `json:"name,omitempty" yaml:"name,omitempty"`
		Ready bool           `json:"ready" yaml:"ready"`
		State ContainerState `json:"state" yaml:"state"`
	}

	Event struct {
		Reason         string      `json:"reason,omitempty" yaml:"reason,omitempty"`
		Message        string      `json:"message,omitempty" yaml:"message,omitempty"`
		Count          int32       `json:"count,omitempty" yaml:"count,omitempty"`
		EventTimestamp metav1.Time `json:"eventTimestamp,omitempty" yaml:"eventTimestamp,omitempty"`
	}

	ContainerState struct {
		State   string `json:"state" yaml:"state"`
		Reason  string `json:"reason" yaml:"reason"`
		Message string `json:"message" yaml:"message"`
	}
)

func getPodTemplateHash(rs *appsv1.ReplicaSet) (string, string) {
	podTemplateHash := rs.Labels[DeploymentPodTemplateHash]
	podTemplateHashKey := DeploymentPodTemplateHash
	if podTemplateHash == "" {
		podTemplateHash = rs.Labels[RolloutPodTemplateHash]
		podTemplateHashKey = RolloutPodTemplateHash
	}
	return podTemplateHashKey, podTemplateHash
}

func getRevision(rs *appsv1.ReplicaSet) string {
	revision := rs.Annotations[_deploymentRevision]
	if revision == "" {
		revision = rs.Annotations[_rolloutRevision]
	}
	return revision
}

// CompareRevision
// rs1 版本更加新的话，则返回true，否则返回false
// 注意：仅仅通过 CreationTimestamp 是无法判断最新版本的，尤其是当 集群进行回滚操作。
func CompareRevision(ctx context.Context, rs1, rs2 *appsv1.ReplicaSet) bool {
	revision1 := getRevision(rs1)
	revision2 := getRevision(rs2)

	hasSameOwnerRef := false
	for _, rs1Owner := range rs1.OwnerReferences {
		for _, rs2Owner := range rs2.OwnerReferences {
			if rs1Owner.UID == rs2Owner.UID {
				hasSameOwnerRef = true
				break
			}
		}
	}

	if revision1 == "" || revision2 == "" || !hasSameOwnerRef {
		// 如果它们属于不同的Deployment或Rollout，或
		// 如果某个revision不存在，
		// 则使用 CreationTimestamp 进行比较
		return rs2.CreationTimestamp.Before(&rs1.CreationTimestamp)
	}

	num1, err := strconv.Atoi(revision1)
	if err != nil {
		log.Error(ctx, err)
		return false
	}
	num2, err := strconv.Atoi(revision2)
	if err != nil {
		log.Error(ctx, err)
		return true
	}

	return num1 > num2
}

func getDeploymentCondition(status appsv1.DeploymentStatus,
	condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}
