package common

const (
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
