package common

import triggers "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"

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
	ClusterClusterLabelKey = "cloudnative.music.netease.com/cluster"
	ClusterRestartTimeKey  = "cloudnative.music.netease.com/user-restart-time"
)

// status of cluster
const (
	ClusterStatusEmpty    = ""
	ClusterStatusFreeing  = "Freeing"
	ClusterStatusFreed    = "Freed"
	ClusterStatusDeleting = "Deleting"
	ClusterStatusCreating = "Creating"
)

const (
	TektonTriggersEventIDKey = triggers.GroupName + triggers.EventIDLabelKey
)
