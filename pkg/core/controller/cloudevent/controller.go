package cloudevent

import (
	"context"
	"strings"

	applicationmanager "github.com/horizoncd/horizon/pkg/application/manager"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/collector"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/factory"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/metrics"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/param"
	prmanager "github.com/horizoncd/horizon/pkg/pipelinerun/manager"
	prmodels "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	pipelinemanager "github.com/horizoncd/horizon/pkg/pipelinerun/pipeline/manager"
	"github.com/horizoncd/horizon/pkg/server/global"
	trmanager "github.com/horizoncd/horizon/pkg/templaterelease/manager"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"

	"github.com/horizoncd/horizon/pkg/core/common"
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
	applicationMgr     applicationmanager.Manager
	userMgr            usermanager.Manager
}

func NewController(tektonFty factory.Factory, parameter *param.Param) Controller {
	return &controller{
		tektonFty:          tektonFty,
		pipelinerunMgr:     parameter.PipelinerunMgr,
		pipelineMgr:        parameter.PipelineMgr,
		clusterMgr:         parameter.ClusterMgr,
		clusterGitRepo:     parameter.ClusterGitRepo,
		templateReleaseMgr: parameter.TemplateReleaseManager,
		applicationMgr:     parameter.ApplicationManager,
		userMgr:            parameter.UserManager,
	}
}

func (c *controller) CloudEvent(ctx context.Context, wpr *WrappedPipelineRun) (err error) {
	const op = "cloudEvent controller: cloudEvent"
	defer wlog.Start(ctx, op).StopPrint()

	horizonMetaData, err := c.getHorizonMetaData(ctx, wpr)
	if err != nil {
		return err
	}

	environment := horizonMetaData.Environment
	pipelinerunID := horizonMetaData.PipelinerunID

	// 1. collect log & pipelinerun object
	tektonCollector, err := c.tektonFty.GetTektonCollector(environment)
	if err != nil {
		return err
	}

	var result *collector.CollectResult
	if result, err = tektonCollector.Collect(ctx, wpr.PipelineRun, horizonMetaData); err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.Warningf(ctx, "received pipelineRun: %v is not found when collect", wpr.PipelineRun.Name)
			return nil
		}
		return err
	}

	// 2. update pipelinerun in db
	if err := c.pipelinerunMgr.UpdateResultByID(ctx, pipelinerunID, &prmodels.Result{
		S3Bucket:   result.Bucket,
		LogObject:  result.LogObject,
		PrObject:   result.PrObject,
		Result:     result.Result,
		StartedAt:  &result.StartTime.Time,
		FinishedAt: &result.CompletionTime.Time,
	}); err != nil {
		return err
	}

	// format Pipeline results
	pipelineResult := metrics.FormatPipelineResults(wpr.PipelineRun)

	// todo remove codes below in the future
	err = c.handleJibBuild(ctx, pipelineResult, horizonMetaData)
	if err != nil {
		return err
	}

	// 4. observe metrics
	metrics.Observe(pipelineResult, horizonMetaData)

	// 5. insert pipeline into db
	err = c.pipelineMgr.Create(ctx, pipelineResult, horizonMetaData)
	if err != nil {
		return err
	}

	return nil
}

// TODO remove this function in the future
// check cluster's build type, change tasks' and steps' values if needed
func (c *controller) handleJibBuild(ctx context.Context, result *metrics.PipelineResults,
	data *global.HorizonMetaData) error {
	clusterID := data.ClusterID
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return err
	}
	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return err
	}
	clusterFiles, err := c.clusterGitRepo.GetCluster(ctx, data.Application, cluster.Name, tr.ChartName)
	if err != nil {
		return err
	}

	// check if buildxml key exist in pipeline
	if buildXML, ok := clusterFiles.PipelineJSONBlob["buildxml"]; ok {
		const jibBuild = "jib-build"
		if strings.Contains(buildXML.(string), "jib-maven-plugin") {
			// change taskrun name
			for _, trResult := range result.TrResults {
				if trResult.Task == "build" {
					trResult.Task = jibBuild
				}
			}
			// change step name
			for _, stepResult := range result.StepResults {
				stepResult.Task = jibBuild
				if stepResult.Step == "compile" {
					stepResult.Step = "jib-compile"
				}
				if stepResult.Step == "image" {
					stepResult.Step = "jib-image"
				}
			}
		}
	}

	return nil
}

// getHorizonMetaData resolves info about this pipelinerun
func (c *controller) getHorizonMetaData(ctx context.Context, wpr *WrappedPipelineRun) (
	*global.HorizonMetaData, error) {
	eventID := wpr.PipelineRun.Labels[common.TektonTriggersEventIDKey]
	pipelinerun, err := c.pipelinerunMgr.GetByCIEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	cluster, err := c.clusterMgr.GetByID(ctx, pipelinerun.ClusterID)
	if err != nil {
		return nil, err
	}
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}
	user, err := c.userMgr.GetUserByID(ctx, pipelinerun.CreatedBy)
	if err != nil {
		return nil, err
	}

	return &global.HorizonMetaData{
		Application:   application.Name,
		ApplicationID: application.ID,
		Cluster:       cluster.Name,
		ClusterID:     cluster.ID,
		Environment:   cluster.EnvironmentName,
		Operator:      user.Email,
		PipelinerunID: pipelinerun.ID,
		Region:        cluster.RegionName,
		Template:      cluster.Template,
		EventID:       eventID,
	}, nil
}
