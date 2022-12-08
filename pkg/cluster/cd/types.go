package cd

import (
	"context"

	"github.com/argoproj/gitops-engine/pkg/health"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetClusterStateParams struct {
	Environment  string
	Cluster      string
	RegionEntity *regionmodels.RegionEntity
}

type GetClusterStateV2Params struct {
	Environment  string
	Cluster      string
	RegionEntity *regionmodels.RegionEntity
}

type CreateClusterParams struct {
	Environment  string
	Cluster      string
	GitRepoURL   string
	ValueFiles   []string
	RegionEntity *regionmodels.RegionEntity
	Namespace    string
}

type DeployClusterParams struct {
	Environment string
	Cluster     string
	Revision    string
}

type GetPodEventsParams struct {
	RegionEntity *regionmodels.RegionEntity
	Cluster      string
	Namespace    string
	Pod          string
}

type GetResourceTreeParams struct {
	Environment string
	Cluster     string
}

type GetPodParams struct {
	RegionEntity *regionmodels.RegionEntity
	Cluster      string
	Namespace    string
	Pod          string
}

type DeletePodsParams struct {
	RegionEntity *regionmodels.RegionEntity
	Namespace    string
	Pods         []string
}

type DeleteClusterParams struct {
	Environment string
	Cluster     string
}

type ClusterNextParams struct {
	Environment string
	Cluster     string
}

type ClusterPromoteParams struct {
	RegionEntity *regionmodels.RegionEntity
	Cluster      string
	Namespace    string
	Environment  string
}

type ClusterPauseParams struct {
	RegionEntity *regionmodels.RegionEntity
	Cluster      string
	Namespace    string
	Environment  string
}

type ClusterResumeParams struct {
	RegionEntity *regionmodels.RegionEntity
	Cluster      string
	Namespace    string
	Environment  string
}

type GetContainerLogParams struct {
	Namespace   string
	Cluster     string
	Pod         string
	Container   string
	Environment string
	TailLines   int
}

type ExecParams struct {
	Environment  string
	Cluster      string
	RegionEntity *regionmodels.RegionEntity
	Namespace    string
	PodList      []string
}

type ExecResp struct {
	key    string
	Result bool
	Stdout string
	Stderr string
	Error  error
}

type OperationResult struct {
	Result bool
	Error  error
}

type ExecFunc func(ctx context.Context, params *ExecParams) (map[string]ExecResp, error)

type containerList struct {
	name       string
	containers []*corev1.Container
}

func (c *containerList) Len() int { return len(c.containers) }

func (c *containerList) Less(i, j int) bool {
	if c.containers[i].Name == c.name {
		return true
	} else if c.containers[j].Name == c.name {
		return false
	}

	a := c.containers[i].Name
	b := c.containers[j].Name

	return a <= b
}

func (c *containerList) Swap(i, j int) {
	c.containers[i], c.containers[j] = c.containers[j], c.containers[i]
}

type Revision struct {
	Pods map[string]*ClusterPod `json:"pods" yaml:"pods"`
}

type ClusterStateV2 struct {
	Workload string                  `json:"workload"`
	Status   health.HealthStatusCode `json:"-"`
	Step     *Step                   `json:"step,omitempty" yaml:"step,omitempty"`
	Revision string                  `json:"revision" yaml:"revision"`
	Versions map[string]*Revision    `json:"versions" yaml:"versions"`
}

// ClusterState 集群状态信息
type ClusterState struct {
	// Status:
	// Processing(正在部署)；Healthy(部署完成)
	// Suspended(已暂停)；Degraded(已降级)
	// Missing(Rollout或Deployment还尚未部署到业务集群)
	// Unknown(集群健康评估失败，无法获悉当前的部署状态)
	Status health.HealthStatusCode `json:"status,omitempty" yaml:"status,omitempty"`

	// Step
	Step *Step `json:"step"`

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

	// ManualPaused indicates whether the cluster is in manual pause state
	ManualPaused bool `json:"manualPaused" yaml:"manualPaused"`
}

type Step struct {
	Index        int   `json:"index"`
	Total        int   `json:"total"`
	Replicas     []int `json:"replicas"`
	ManualPaused bool  `json:"manualPaused"`
}

// ClusterVersion version information
type ClusterVersion struct {
	// Replicas the replicas of this revision
	Replicas int    `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`
	// Pods the pods detail of this revision, the key is pod name
	Pods map[string]*ClusterPod `json:"pods,omitempty" yaml:"pods,omitempty"`
}

// ClusterPod pod detail
type ClusterPod struct {
	Metadata          PodMetadata  `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec              PodSpec      `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status            PodStatus    `json:"status,omitempty" yaml:"status,omitempty"`
	DeletionTimestamp *metav1.Time `json:"deletionTimestamp,omitempty"`
}

type PodMetadata struct {
	CreationTimestamp metav1.Time       `json:"creationTimestamp"`
	Namespace         string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type PodSpec struct {
	NodeName       string       `json:"nodeName,omitempty" yaml:"nodeName,omitempty"`
	InitContainers []*Container `json:"initContainers,omitempty" yaml:"initContainers,omitempty"`
	Containers     []*Container `json:"containers,omitempty" yaml:"containers,omitempty"`
}

type Container struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
}

type PodStatus struct {
	HostIP            string             `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
	PodIP             string             `json:"podIP,omitempty" yaml:"podIP,omitempty"`
	Phase             string             `json:"phase,omitempty" yaml:"phase,omitempty"`
	Events            []Event            `json:"events,omitempty" yaml:"events,omitempty"`
	ContainerStatuses []*ContainerStatus `json:"containerStatuses,omitempty" yaml:"containerStatuses,omitempty"`
	LifeCycle         []*LifeCycleItem   `json:"lifeCycle" yaml:"lifeCycle"`
}

type ContainerStatus struct {
	Name         string         `json:"name,omitempty" yaml:"name,omitempty"`
	Ready        bool           `json:"ready" yaml:"ready"`
	RestartCount int32          `json:"restartCount"`
	State        ContainerState `json:"state" yaml:"state"`
	ImageID      string         `json:"imageID" yaml:"imageID"`
}

type Event struct {
	Type           string      `json:"type" yaml:"type"`
	Reason         string      `json:"reason,omitempty" yaml:"reason,omitempty"`
	Message        string      `json:"message,omitempty" yaml:"message,omitempty"`
	Count          int32       `json:"count,omitempty" yaml:"count,omitempty"`
	EventTimestamp metav1.Time `json:"eventTimestamp,omitempty" yaml:"eventTimestamp,omitempty"`
}

// ContainerDetail represents more information about a container
type ContainerDetail struct {
	// Name of the container.
	Name string `json:"name"`

	// Image URI of the container.
	Image string `json:"image"`

	// List of environment variables.
	Env []corev1.EnvVar `json:"env"`

	// Commands of the container
	Commands []string `json:"commands"`

	// Command arguments
	Args []string `json:"args"`

	// Information about mounted volumes
	VolumeMounts []VolumeMount `json:"volumeMounts"`

	// Security configuration that will be applied to a container.
	SecurityContext *corev1.SecurityContext `json:"securityContext"`

	// Status of a pod container
	Status *corev1.ContainerStatus `json:"status"`

	// Probes
	LivenessProbe  *corev1.Probe `json:"livenessProbe"`
	ReadinessProbe *corev1.Probe `json:"readinessProbe"`
	StartupProbe   *corev1.Probe `json:"startupProbe"`
}

type VolumeMount struct {
	// Name of the variable.
	Name string `json:"name"`

	// Is the volume read only ?
	ReadOnly bool `json:"readOnly"`

	// Path within the container at which the volume should be mounted. Must not contain ':'.
	MountPath string `json:"mountPath"`

	// Path within the volume from which the container's volume should be mounted. Defaults to "" (volume's root).
	SubPath string `json:"subPath"`

	// Information about the Volume itself
	Volume corev1.Volume `json:"volume"`
}

type ContainerState struct {
	State     string       `json:"state" yaml:"state"`
	Reason    string       `json:"reason" yaml:"reason"`
	Message   string       `json:"message" yaml:"message"`
	StartedAt *metav1.Time `json:"startedAt,omitempty"`
}

type LifeCycleItem struct {
	Type         string      `json:"type" yaml:"type"`
	Status       string      `json:"status" yaml:"status"`
	Message      string      `json:"message" yaml:"status"`
	CompleteTime metav1.Time `json:"completeTime" yaml:"completeTime"`
}
