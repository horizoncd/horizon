package slo

import (
	"context"
	"g.hz.netease.com/horizon/pkg/pipeline/manager"
)

const (
	BuildTask            = "build"
	BuildTaskDisplayName = "构建"
	GitStep              = "git"
	CompileStep          = "compile"
	ImageStep            = "image"

	DeployTask            = "deploy"
	DeployTaskDisplayName = "发布（环境准备）"
	DeployStep            = "deploy"
)

var (
	envMapping = map[string][]string{
		"test":   {"perf", "reg", "test"},
		"online": {"pre", "online"},
	}

	// 部署RT临界值，后续放入配置项
	_deployRT uint = 30

	// 发布RT临界值，后续放入配置项
	_buildRT uint = 60
)

type Controller interface {
	PipelineSLO(ctx context.Context, environment string, start, end int64) (pipelineSLOs []*PipelineSLO, err error)
}

type controller struct {
	pipelineManager manager.Manager
}

func (c controller) PipelineSLO(ctx context.Context, environment string,
	start, end int64) (pipelineSLOs []*PipelineSLO, err error) {
	slos, err := c.pipelineManager.ListPipelineSLOsByEnvsAndTimeRange(ctx, envMapping[environment], start, end)
	if err != nil {
		return nil, err
	}

	pipelineSLOMap := map[string]*PipelineSLO{
		BuildTask: {
			Name:                BuildTask,
			DisplayName:         BuildTaskDisplayName,
			Count:               0,
			RequestAvailability: 0,
			RTAvailability:      0,
			RT:                  _deployRT,
		},
		DeployTask: {
			Name:                DeployTask,
			DisplayName:         DeployTaskDisplayName,
			Count:               0,
			RequestAvailability: 0,
			RTAvailability:      0,
			RT:                  _buildRT,
		},
	}

	buildTaskCount, buildSuccessCount, buildRTSuccessCount := 0, 0, 0
	deployTaskCount, deploySuccessCount, deployRTSuccessCount := 0, 0, 0
	for _, slo := range slos {
		if build, ok := slo.Tasks[BuildTask]; ok {
			buildTaskCount++
			if build.Result == "ok" {
				buildSuccessCount++
				// 这里注意是用整体Task的耗时减掉compile step的耗时，这样的结果更加准确，包含了Pod启动准备所需的时间
				if build.Duration-slo.Tasks[BuildTask].Steps[CompileStep].Duration < pipelineSLOMap[BuildTask].RT {
					buildRTSuccessCount++
				}
			} else {
				// 如果是compile失败了，slo维度也认为是成功的
				if compile, ok := slo.Tasks[BuildTask].Steps[CompileStep]; ok && compile.Result == "failed" {
					buildSuccessCount++
				}
			}
		}
		if deploy, ok := slo.Tasks[DeployTask]; ok {
			deployTaskCount++
			if deploy.Result == "ok" {
				deploySuccessCount++
				if deploy.Duration < pipelineSLOMap[DeployTask].RT {
					deployRTSuccessCount++
				}
			}
		}
	}

	pipelineSLOMap[BuildTask].Count = buildTaskCount
	pipelineSLOMap[BuildTask].RequestAvailability = buildSuccessCount * 100 / buildTaskCount
	pipelineSLOMap[BuildTask].RTAvailability = buildRTSuccessCount * 100 / buildTaskCount

	pipelineSLOMap[DeployTask].Count = deployTaskCount
	pipelineSLOMap[DeployTask].RequestAvailability = deploySuccessCount * 100 / deployTaskCount
	pipelineSLOMap[DeployTask].RTAvailability = deployRTSuccessCount * 100 / deployTaskCount

	for _, slo := range pipelineSLOMap {
		pipelineSLOs = append(pipelineSLOs, slo)
	}
	return pipelineSLOs, nil
}

// NewController initializes a new group controller
func NewController() Controller {
	return &controller{
		pipelineManager: manager.Mgr,
	}
}
