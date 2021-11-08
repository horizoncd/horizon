package pipelinerun

import (
	"context"
	"fmt"

	appmanager "g.hz.netease.com/horizon/pkg/application/manager"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/log"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	prmanager "g.hz.netease.com/horizon/pkg/pipelinerun/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	GetPrLog(ctx context.Context, prID uint) (*Log, error)
}

type controller struct {
	prMgr          prmanager.Manager
	clusterMgr     clustermanager.Manager
	applicationMgr appmanager.Manager
	envMgr         envmanager.Manager
	tektonFty      factory.Factory
}

var _ Controller = (*controller)(nil)

func NewController(tektonFty factory.Factory) Controller {
	return &controller{
		prMgr:          prmanager.Mgr,
		clusterMgr:     clustermanager.Mgr,
		applicationMgr: appmanager.Mgr,
		envMgr:         envmanager.Mgr,
		tektonFty:      tektonFty,
	}
}

type Log struct {
	LogChannel <-chan log.Log
	ErrChannel <-chan error

	LogBytes []byte
}

func (c *controller) GetPrLog(ctx context.Context, prID uint) (_ *Log, err error) {
	const op = "pipelinerun controller: get pipelinerun log"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	pr, err := c.prMgr.GetByID(ctx, prID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	cluster, err := c.clusterMgr.GetByID(ctx, pr.ClusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// only builddeploy have logs
	// TODO(gjq): common const variable
	if pr.Action != "builddeploy" {
		return nil, errors.E(op, fmt.Errorf("%v action has no log", pr.Action))
	}

	// if pr.PrObject is empty, get pipelinerun log in k8s
	if pr.PrObject == "" {
		tektonClient, err := c.tektonFty.GetTekton(er.EnvironmentName)
		if err != nil {
			return nil, errors.E(op, err)
		}

		// TODO(gjq): get pipelinerun log by pipelinerunID
		logCh, errCh, err := tektonClient.GetLatestPipelineRunLog(ctx, application.Name, cluster.Name)
		return &Log{
			LogChannel: logCh,
			ErrChannel: errCh,
		}, nil
	}

	// else, get log from s3
	tektonCollector, err := c.tektonFty.GetTektonCollector(er.EnvironmentName)
	if err != nil {
		return nil, errors.E(op, err)
	}
	// TODO(gjq): get pipelinerun log by pr.Object
	logBytes, err := tektonCollector.GetLatestPipelineRunLog(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return &Log{
		LogBytes: logBytes,
	}, nil
}
