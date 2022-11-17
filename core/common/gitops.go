package common

const (
	// GitopsBranchMaster is the main branch
	GitopsBranchMaster = "master"
	// GitopsBranchGitops is the gitops branch, values updated in this branch, then merge into the GitopsBranchMaster
	GitopsBranchGitops = "gitops"

	// fileName
	GitopsFileChart          = "Chart.yaml"
	GitopsFileApplication    = "application.yaml"
	GitopsFileTags           = "tags.yaml"
	GitopsFileSRE            = "sre/sre.yaml"
	GitopsFileBase           = "system/horizon.yaml"
	GitopsFileEnv            = "system/env.yaml"
	GitopsFileRestart        = "system/restart.yaml"
	GitopsFilePipeline       = "pipeline/pipeline.yaml"
	GitopsFilePipelineOutput = "pipeline/pipeline-output.yaml"
	GitopsFileManifest       = "manifest.yaml"

	// value namespace
	GitopsEnvValueNamespace  = "env"
	GitopsBaseValueNamespace = "horizon"

	GitopsMergeRequestStateOpen = "opened"

	GitopsGroupClusters          = "clusters"
	GitopsGroupRecyclingClusters = "recycling-clusters"

	GitopsKeyTags = "tags"
)
