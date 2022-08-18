package cloudevent

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	prmanager "g.hz.netease.com/horizon/pkg/pipelinerun/manager"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	pipelinemanager "g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/manager"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"

	"g.hz.netease.com/horizon/pkg/cluster/common"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/collector"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	CloudEvent(ctx context.Context, wpr *WrappedPipelineRun) error
}

type controller struct {
	tektonFty          factory.Factory
	pipelinerunMgr     prmanager.Manager
	pipelineMgr        pipelinemanager.Manager
	clusterMgr         clustermanager.Manager
	clusterGitRepo     gitrepo.ClusterGitRepo
	templateReleaseMgr trmanager.Manager
}

func NewController(tektonFty factory.Factory, parameter *param.Param) Controller {
	return &controller{
		tektonFty:          tektonFty,
		pipelinerunMgr:     parameter.PipelinerunMgr,
		pipelineMgr:        parameter.PipelineMgr,
		clusterMgr:         parameter.ClusterMgr,
		clusterGitRepo:     parameter.ClusterGitRepo,
		templateReleaseMgr: parameter.TemplateReleaseManager,
	}
}

func (c *controller) CloudEvent(ctx context.Context, wpr *WrappedPipelineRun) (err error) {
	const op = "cloudEvent controller: cloudEvent"
	defer wlog.Start(ctx, op).StopPrint()

	environment := wpr.PipelineRun.Labels[common.EnvironmentLabelKey]
	pipelinerunIDStr := wpr.PipelineRun.Labels[common.PipelinerunIDLabelKey]
	pipelinerunID, err := strconv.ParseUint(pipelinerunIDStr, 10, 0)
	if err != nil {
		return errors.E(op, err)
	}

	// 1. collect log & pipelinerun object
	tektonCollector, err := c.tektonFty.GetTektonCollector(environment)
	if err != nil {
		return errors.E(op, err)
	}

	var result *collector.CollectResult
	if result, err = tektonCollector.Collect(ctx, wpr.PipelineRun); err != nil {
		if errors.Status(err) == http.StatusNotFound {
			// 如果pipelineRun已经不存在，直接忽略。
			// 这种情况一般是 tekton pipeline controller重复上报了同一个pipelineRun所致
			log.Warningf(ctx, "received pipelineRun: %v is not found when collect", wpr.PipelineRun.Name)
			return nil
		}
		return errors.E(op, err)
	}

	// 2. update pipelinerun in db
	if err := c.pipelinerunMgr.UpdateResultByID(ctx, uint(pipelinerunID), &prmodels.Result{
		S3Bucket:   result.Bucket,
		LogObject:  result.LogObject,
		PrObject:   result.PrObject,
		Result:     result.Result,
		StartedAt:  &result.StartTime.Time,
		FinishedAt: &result.CompletionTime.Time,
	}); err != nil {
		return errors.E(op, err)
	}

	// 3. delete pipelinerun in k8s
	// 提前删除，如果删除返回404，则忽略
	// 这种情况一般也是 tekton pipeline controller重复上报了同一个pipelineRun所致
	tekton, err := c.tektonFty.GetTekton(environment)
	if err != nil {
		return errors.E(op, err)
	}

	if err := tekton.DeletePipelineRun(ctx, wpr.PipelineRun); err != nil {
		if errors.Status(err) == http.StatusNotFound {
			log.Warningf(ctx, "received pipelineRun: %v is not found when delete", wpr.PipelineRun.Name)
			return nil
		}
		return errors.E(op, err)
	}

	// format Pipeline results
	pipelineResult := metrics.FormatPipelineResults(wpr.PipelineRun)

	// 判断集群是否是JIB构建，动态修改Task和Step的值
	c.handleJibBuild(ctx, pipelineResult)

	// 4. observe metrics
	// 最后指标上报，保证同一条pipelineRun，只上报一条指标
	metrics.Observe(pipelineResult)

	// 5. insert pipeline into db
	err = c.pipelineMgr.Create(ctx, pipelineResult)
	if err != nil {
		// err不往上层抛，上层也无法处理这种异常
		log.Errorf(ctx, "failed to save pipeline to db: %v, err: %v", pipelineResult, err)
	}

	return nil
}

// TODO remove this function in the future
// 判断集群是否是JIB构建，动态修改Task和Step的值
func (c *controller) handleJibBuild(ctx context.Context, result *metrics.PipelineResults) {
	clusterID, err := strconv.ParseUint(result.BusinessData.ClusterID, 10, 0)
	if err != nil {
		log.Errorf(ctx, "failed to parse clusterID to uint from string: %d", clusterID)
		return
	}
	cluster, err := c.clusterMgr.GetByID(ctx, uint(clusterID))
	if err != nil {
		log.Errorf(ctx, "failed to get cluster from db by id: %d, err: %+v", clusterID, err)
		return
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		log.Errorf(ctx, "failed to get templateRelease from db by id: %d, err: %+v", cluster.ApplicationID, err)
		return
	}
	clusterFiles, err := c.clusterGitRepo.GetCluster(ctx, result.BusinessData.Application, cluster.Name, tr.ChartName)
	if err != nil {
		log.Errorf(ctx, "failed to get files from gitlab, cluster: %s, err: %+v", cluster.Name, err)
		return
	}

	// check if buildxml key exist in pipeline
	if buildXML, ok := clusterFiles.PipelineJSONBlob["buildxml"]; ok {
		// 判断buildxml包含jib的内容，则进行替换操作
		if strings.Contains(buildXML.(string), "jib-maven-plugin") {
			// change taskrun name
			for _, trResult := range result.TrResults {
				if trResult.Name == "build" {
					trResult.Name = "jib-build"
				}
			}
			// change step name
			for _, stepResult := range result.StepResults {
				if stepResult.Step == "compile" {
					stepResult.Step = "jib-compile"
				}
				if stepResult.Step == "image" {
					stepResult.Step = "jib-image"
				}
			}
		}
	}
}
