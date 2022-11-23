package common

const (
	ClusterQueryEnvironment   = "environment"
	ClusterQueryName          = "filter"
	ClusterQueryByUser        = "userID"
	ClusterQueryByTemplate    = "template"
	ClusterQueryByRelease     = "templateRelease"
	ClusterQueryTagSelector   = "tagSelector"
	ClusterQueryScope         = "scope"
	ClusterQueryMergePatch    = "mergePatch"
	ClusterQueryTargetBranch  = "targetBranch"
	ClusterQueryTargetCommit  = "targetCommit"
	ClusterQueryTargetTag     = "targetTag"
	ClusterQueryContainerName = "containerName"
	ClusterQueryPodName       = "podName"
	ClusterQueryTailLines     = "tailLines"
	ClusterQueryStart         = "start"
	ClusterQueryEnd           = "end"
	ClusterQueryExtraOwner    = "extraOwner"
	ClusterQueryHard          = "hard"
)

const (
	ClusterApplicationLabelKey   = "cloudnative.music.netease.com/application"
	ClusterApplicationIDLabelKey = "cloudnative.music.netease.com/application-id"
	ClusterClusterLabelKey       = "cloudnative.music.netease.com/cluster"
	ClusterClusterIDLabelKey     = "cloudnative.music.netease.com/cluster-id"
	ClusterEnvironmentLabelKey   = "cloudnative.music.netease.com/environment"
	ClusterPipelinerunIDLabelKey = "cloudnative.music.netease.com/pipelinerun-id"
	ClusterRegionLabelKey        = "cloudnative.music.netease.com/region"
	ClusterRegionIDLabelKey      = "cloudnative.music.netease.com/region-id"
	ClusterOperatorAnnotationKey = "cloudnative.music.netease.com/operator"
	ClusterTemplateKey           = "cloudnative.music.netease.com/template"
	ClusterRestartTimeKey        = "cloudnative.music.netease.com/user-restart-time"
)

// status of cluster
const (
	ClusterStatusEmpty    = ""
	ClusterStatusFreeing  = "Freeing"
	ClusterStatusFreed    = "Freed"
	ClusterStatusDeleting = "Deleting"
	ClusterStatusCreating = "Creating"
)
