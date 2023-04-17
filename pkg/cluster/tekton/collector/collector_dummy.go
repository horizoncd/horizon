package collector

import (
	"context"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/tekton"
	perror "github.com/horizoncd/horizon/pkg/errors"
	prmodels "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	logutil "github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

type DummyCollector struct {
	tekton tekton.Interface
}

func NewDummyCollector(tekton tekton.Interface) Interface {
	return &DummyCollector{
		tekton: tekton,
	}
}

func (c *DummyCollector) Collect(ctx context.Context, pr *v1beta1.PipelineRun,
	horizonMetaData *global.HorizonMetaData,
) (*CollectResult, error) {
	const op = "DummyCollector: collect"
	defer wlog.Start(ctx, op).StopPrint()

	metadata := resolveObjMetadata(pr, horizonMetaData)
	collectResult := &CollectResult{
		Result:         metadata.PipelineRun.Result,
		StartTime:      metadata.PipelineRun.StartTime,
		CompletionTime: metadata.PipelineRun.CompletionTime,
	}
	logutil.Infof(ctx, "collected pipelineRun log: name: %v, %+v",
		metadata.PipelineRun.Name, collectResult)
	return collectResult, nil
}

func (c *DummyCollector) GetPipelineRunLog(ctx context.Context, pr *prmodels.Pipelinerun) (*Log, error) {
	const op = "DummyCollector: getPipelineRunLog"
	defer wlog.Start(ctx, op).StopPrint()

	// get logs from k8s directly
	logCh, errCh, err := c.tekton.GetPipelineRunLogByID(ctx, pr.CIEventID)
	if err != nil {
		return nil, perror.WithMessagef(err, "failed to get pipelineRun log from k8s")
	}
	return &Log{
		LogChannel: logCh,
		ErrChannel: errCh,
	}, nil
}

func (c *DummyCollector) GetPipelineRunObject(_ context.Context,
	_ string,
) (*Object, error) {
	// no storage to collect pipelineRun object
	return nil, nil
}

func (c *DummyCollector) GetPipelineRun(ctx context.Context,
	pr *prmodels.Pipelinerun,
) (*v1beta1.PipelineRun, error) {
	const op = "DummyCollector: getPipelineRun"
	defer wlog.Start(ctx, op).StopPrint()

	// get pipelineRun from k8s directly
	tektonPipelineRun, err := c.tekton.GetPipelineRunByID(ctx, pr.CIEventID)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			return nil, nil
		}
		return nil, err
	}
	return tektonPipelineRun, nil
}
