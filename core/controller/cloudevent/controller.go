package cloudevent

import (
	"context"
	"net/http"
	"strconv"

	"g.hz.netease.com/horizon/pkg/cluster/common"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/collector"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
	prmanager "g.hz.netease.com/horizon/pkg/pipelinerun/manager"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	CloudEvent(ctx context.Context, wpr *WrappedPipelineRun) error
}

type controller struct {
	tektonFty      factory.Factory
	pipelinerunMgr prmanager.Manager
}

func NewController(tektonFty factory.Factory) Controller {
	return &controller{
		tektonFty:      tektonFty,
		pipelinerunMgr: prmanager.Mgr,
	}
}

func (c *controller) CloudEvent(ctx context.Context, wpr *WrappedPipelineRun) (err error) {
	const op = "cloudEvent controller: cloudEvent"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

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

	// 4. observe metrics
	// 最后指标上报，保证同一条pipelineRun，只上报一条指标
	metrics.Observe(&metrics.WrappedPipelineRun{
		PipelineRun: wpr.PipelineRun,
	})

	return nil
}
