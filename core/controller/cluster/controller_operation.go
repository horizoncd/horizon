package cluster

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

func (c *controller) Restart(ctx context.Context, clusterID uint) (_ *PipelinerunIDResponse, err error) {
	const op = "cluster controller: restart "
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
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

	// 1. get config commit now
	lastConfigCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 2. update restartTime in git repo, and return the newest commit
	commit, err := c.clusterGitRepo.UpdateRestartTime(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 3. deploy cluster in cd system
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: er.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    commit,
	}); err != nil {
		return nil, errors.E(op, err)
	}

	// 4. add pipelinerun in db
	timeNow := time.Now()
	pr := &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           prmodels.ActionRestart,
		Status:           prmodels.ResultOK,
		Title:            "restart",
		LastConfigCommit: lastConfigCommit.Master,
		ConfigCommit:     commit,
		StartedAt:        &timeNow,
		FinishedAt:       &timeNow,
		CreatedBy:        currentUser.GetID(),
	}
	prCreated, err := c.pipelinerunMgr.Create(ctx, pr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &PipelinerunIDResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}

func (c *controller) Deploy(ctx context.Context, clusterID uint,
	r *DeployRequest) (_ *PipelinerunIDResponse, err error) {
	const op = "cluster controller: deploy "
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
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

	// 1. get config commit
	configCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, errors.E(op, err)
	}
	diff, err := c.clusterGitRepo.CompareConfig(ctx, application.Name, cluster.Name,
		&configCommit.Master, &configCommit.Gitops)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if diff == "" {
		return nil, errors.E(op, http.StatusBadRequest, errors.ErrorCode("NoChange"), "there is no change to deploy")
	}

	// 2. merge branch
	commit, err := c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 3. deploy cluster in cd system
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: er.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    commit,
	}); err != nil {
		return nil, errors.E(op, err)
	}

	timeNow := time.Now()
	// 4. add pipelinerun in db
	pr := &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           prmodels.ActionDeploy,
		Status:           prmodels.ResultOK,
		Title:            r.Title,
		Description:      r.Description,
		LastConfigCommit: configCommit.Master,
		ConfigCommit:     configCommit.Gitops,
		StartedAt:        &timeNow,
		FinishedAt:       &timeNow,
		CreatedBy:        currentUser.GetID(),
	}
	prCreated, err := c.pipelinerunMgr.Create(ctx, pr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &PipelinerunIDResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}

func (c *controller) Rollback(ctx context.Context,
	clusterID uint, r *RollbackRequest) (_ *PipelinerunIDResponse, err error) {
	const op = "cluster controller: rollback "
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. get pipelinerun to rollback, and do some validation
	pipelinerun, err := c.pipelinerunMgr.GetByID(ctx, r.PipelinerunID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if pipelinerun.Action == prmodels.ActionRestart || pipelinerun.Status != prmodels.ResultOK ||
		pipelinerun.ConfigCommit == "" {
		return nil, errors.E(op, fmt.Errorf("the pipelinerun with id: %v can not be rollbacked", r.PipelinerunID))
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if pipelinerun.ClusterID != cluster.ID {
		return nil, errors.E(op, fmt.Errorf(
			"the pipelinerun with id: %v is not belongs to cluster: %v", r.PipelinerunID, clusterID))
	}

	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 2. get config commit now
	lastConfigCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 3. rollback cluster config in git repo
	newConfigCommit, err := c.clusterGitRepo.Rollback(ctx, application.Name, cluster.Name, pipelinerun.ConfigCommit)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 4. merge branch
	masterRevision, err := c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 5. deploy cluster in cd
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: er.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    masterRevision,
	}); err != nil {
		return nil, errors.E(op, err)
	}

	timeNow := time.Now()
	// 6. add pipelinerun in db
	pr := &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           prmodels.ActionRollback,
		Status:           prmodels.ResultOK,
		Title:            prmodels.ActionRollback,
		GitURL:           pipelinerun.GitURL,
		GitBranch:        pipelinerun.GitBranch,
		GitCommit:        pipelinerun.GitCommit,
		ImageURL:         pipelinerun.ImageURL,
		LastConfigCommit: lastConfigCommit.Master,
		ConfigCommit:     newConfigCommit,
		StartedAt:        &timeNow,
		FinishedAt:       &timeNow,
		RollbackFrom:     &r.PipelinerunID,
		CreatedBy:        currentUser.GetID(),
	}
	prCreated, err := c.pipelinerunMgr.Create(ctx, pr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &PipelinerunIDResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}

func (c *controller) Next(ctx context.Context, clusterID uint) (err error) {
	const op = "cluster controller: next"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return errors.E(op, err)
	}

	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return errors.E(op, err)
	}

	if err := c.cd.Next(ctx, &cd.ClusterNextParams{
		Environment: er.EnvironmentName,
		Cluster:     cluster.Name,
	}); err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (c *controller) Online(ctx context.Context, clusterID uint, r *ExecRequest) (_ ExecResponse, err error) {
	const op = "cluster controller: online"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	return c.exec(ctx, clusterID, r, c.cd.Online)
}

func (c *controller) Offline(ctx context.Context, clusterID uint, r *ExecRequest) (_ ExecResponse, err error) {
	const op = "cluster controller: offline"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	return c.exec(ctx, clusterID, r, c.cd.Offline)
}

func (c *controller) exec(ctx context.Context, clusterID uint,
	r *ExecRequest, execFunc cd.ExecFunc) (_ ExecResponse, err error) {
	const op = "cluster controller: exec"

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, er.RegionName)
	if err != nil {
		return nil, errors.E(op, err)
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}

	execResp, err := execFunc(ctx, &cd.ExecParams{
		Environment:  er.EnvironmentName,
		Cluster:      cluster.Name,
		RegionEntity: regionEntity,
		Namespace:    envValue.Namespace,
		PodList:      r.PodList,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return ofExecResp(execResp), nil
}
