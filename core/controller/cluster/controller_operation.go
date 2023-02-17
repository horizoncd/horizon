package cluster

import (
	"context"
	"time"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	amodels "github.com/horizoncd/horizon/pkg/application/models"
	"github.com/horizoncd/horizon/pkg/cluster/cd"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	cmodels "github.com/horizoncd/horizon/pkg/cluster/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	prmodels "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	tmodels "github.com/horizoncd/horizon/pkg/tag/models"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

func (c *controller) Restart(ctx context.Context, clusterID uint) (_ *PipelinerunIDResponse, err error) {
	const op = "cluster controller: restart "
	defer wlog.Start(ctx, op).StopPrint()

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 1. get config commit now
	lastConfigCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}

	// 2. create pipeline record
	prCreated, err := c.pipelinerunMgr.Create(ctx, &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           prmodels.ActionRestart,
		Status:           string(prmodels.StatusCreated),
		Title:            prmodels.ActionRestart,
		LastConfigCommit: lastConfigCommit.Master,
		ConfigCommit:     lastConfigCommit.Master,
	})

	// 2. update restartTime in git repo, and return the newest commit
	commit, err := c.clusterGitRepo.UpdateRestartTime(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, err
	}
	if err := c.pipelinerunMgr.UpdateConfigCommitByID(ctx, prCreated.ID, commit); err != nil {
		log.Errorf(ctx, "UpdateConfigCommitByID error, pr = %d, commit = %s, err = %v",
			prCreated.ID, commit, err)
	}
	if err := c.updatePRStatus(ctx, prmodels.ActionRestart, prCreated.ID, prmodels.StatusMerged, commit); err != nil {
		return nil, err
	}

	// 3. deploy cluster in cd system
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: cluster.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    commit,
	}); err != nil {
		return nil, err
	}
	log.Infof(ctx, "Restart Deployed, pr = %d, commit = %s", prCreated.ID, commit)

	// 4. update status
	if err := c.updatePRStatus(ctx, prmodels.ActionRestart, prCreated.ID, prmodels.StatusOK, commit); err != nil {
		return nil, err
	}
	return &PipelinerunIDResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}

func (c *controller) Deploy(ctx context.Context, clusterID uint,
	r *DeployRequest) (_ *PipelinerunIDResponse, err error) {
	const op = "cluster controller: deploy "
	defer wlog.Start(ctx, op).StopPrint()

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}
	clusterFiles, err := c.clusterGitRepo.GetCluster(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, err
	}
	// if pipeline config exists && image is empty, should buildDeploy first
	if len(clusterFiles.PipelineJSONBlob) > 0 {
		po, err := c.clusterGitRepo.GetPipelineOutput(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil {
			if perror.Cause(err) != herrors.ErrPipelineOutputEmpty {
				return nil, err
			}
			return nil, herrors.ErrShouldBuildDeployFirst
		}
		if po == nil {
			return nil, herrors.ErrShouldBuildDeployFirst
		}
	}

	// 1. get config commit
	configCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}
	diff, err := c.clusterGitRepo.CompareConfig(ctx, application.Name, cluster.Name,
		&configCommit.Master, &configCommit.Gitops)
	if err != nil {
		return nil, err
	}
	if diff == "" && cluster.Status != common.ClusterStatusFreed {
		return nil, perror.Wrap(herrors.ErrClusterNoChange, "there is no change to deploy")
	}

	// 2. create pipeline record
	prCreated, err := c.pipelinerunMgr.Create(ctx, &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           prmodels.ActionDeploy,
		Status:           string(prmodels.StatusCreated),
		Title:            r.Title,
		Description:      r.Description,
		LastConfigCommit: configCommit.Master,
		ConfigCommit:     configCommit.Gitops,
	})
	if err != nil {
		return nil, err
	}

	// 3. merge branch & update status
	var commit string
	if diff == "" {
		// freed cluster is allowed to deploy without diff
		commit = configCommit.Master
	} else {
		commit, err = c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name,
			gitrepo.GitOpsBranch, c.clusterGitRepo.DefaultBranch(), &prCreated.ID)
		if err != nil {
			return nil, err
		}
	}
	if err := c.updatePRStatus(ctx, prmodels.ActionDeploy, prCreated.ID, prmodels.StatusMerged, commit); err != nil {
		return nil, err
	}

	// 5. create cluster in cd system
	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}
	repoInfo := c.clusterGitRepo.GetRepoInfo(ctx, application.Name, cluster.Name)
	if err := c.cd.CreateCluster(ctx, &cd.CreateClusterParams{
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
		GitRepoURL:   repoInfo.GitRepoURL,
		ValueFiles:   repoInfo.ValueFiles,
		RegionEntity: regionEntity,
		Namespace:    envValue.Namespace,
	}); err != nil {
		return nil, err
	}

	// 6. reset cluster status
	if cluster.Status == common.ClusterStatusFreed {
		cluster.Status = common.ClusterStatusEmpty
		cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
		if err != nil {
			return nil, err
		}
	}

	// 7. deploy cluster in cd system
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: cluster.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    commit,
	}); err != nil {
		return nil, err
	}
	if err := c.updatePRStatus(ctx, prmodels.ActionDeploy, prCreated.ID, prmodels.StatusOK, commit); err != nil {
		return nil, err
	}

	// 8. record event
	if _, err := c.eventMgr.CreateEvent(ctx, &eventmodels.Event{
		EventSummary: eventmodels.EventSummary{
			ResourceType: common.ResourceCluster,
			EventType:    eventmodels.ClusterDeployed,
			ResourceID:   cluster.ID,
		},
	}); err != nil {
		log.Warningf(ctx, "failed to create event, err: %s", err.Error())
	}

	return &PipelinerunIDResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}

func (c *controller) Rollback(ctx context.Context,
	clusterID uint, r *RollbackRequest) (_ *PipelinerunIDResponse, err error) {
	const op = "cluster controller: rollback "
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get pipelinerun to rollback, and do some validation
	pipelinerun, err := c.pipelinerunMgr.GetByID(ctx, r.PipelinerunID)
	if err != nil {
		return nil, err
	}

	if pipelinerun.Action == prmodels.ActionRestart || pipelinerun.Status != string(prmodels.StatusOK) ||
		pipelinerun.ConfigCommit == "" {
		return nil, perror.Wrapf(herrors.ErrParamInvalid,
			"the pipelinerun with id: %v can not be rolled back", r.PipelinerunID)
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	if pipelinerun.ClusterID != cluster.ID {
		return nil, perror.Wrapf(herrors.ErrParamInvalid,
			"the pipelinerun with id: %v is not belongs to cluster: %v", r.PipelinerunID, clusterID)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 2. get config commit now
	lastConfigCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}

	// 3. create record
	prCreated, err := c.pipelinerunMgr.Create(ctx, &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           prmodels.ActionRollback,
		Status:           string(prmodels.StatusCreated),
		Title:            prmodels.ActionRollback,
		GitURL:           pipelinerun.GitURL,
		GitRefType:       pipelinerun.GitRefType,
		GitRef:           pipelinerun.GitRef,
		GitCommit:        pipelinerun.GitCommit,
		ImageURL:         pipelinerun.ImageURL,
		LastConfigCommit: lastConfigCommit.Master,
		ConfigCommit:     lastConfigCommit.Master,
		RollbackFrom:     &r.PipelinerunID,
	})
	if err != nil {
		return nil, err
	}

	// 4. update restart time in gitops branch
	err = c.mergeDefaultBranchIntoGitOps(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}

	// 5. rollback cluster config in git repo and update status
	newConfigCommit, err := c.clusterGitRepo.Rollback(ctx, application.Name, cluster.Name, pipelinerun.ConfigCommit)
	if err != nil {
		return nil, err
	}
	if err := c.updatePRStatus(ctx, prmodels.ActionRollback, prCreated.ID, prmodels.StatusCommitted,
		newConfigCommit); err != nil {
		return nil, err
	}

	// 6. merge branch & update config commit and status
	masterRevision, err := c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name,
		gitrepo.GitOpsBranch, c.clusterGitRepo.DefaultBranch(), &prCreated.ID)
	if err != nil {
		return nil, err
	}
	if err := c.pipelinerunMgr.UpdateConfigCommitByID(ctx, prCreated.ID, masterRevision); err != nil {
		log.Errorf(ctx, "UpdateConfigCommitByID error, pr = %d, commit = %s, err = %v",
			prCreated.ID, masterRevision, err)
	}
	if err := c.updatePRStatus(ctx, prmodels.ActionRollback, prCreated.ID, prmodels.StatusMerged,
		masterRevision); err != nil {
		return nil, err
	}

	// 7. update tags and template in db
	if err := c.updateTagsAndTemplateFromFile(ctx, application, cluster); err != nil {
		return nil, err
	}

	// 8. create cluster in cd system
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, err
	}
	repoInfo := c.clusterGitRepo.GetRepoInfo(ctx, application.Name, cluster.Name)
	if err := c.cd.CreateCluster(ctx, &cd.CreateClusterParams{
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
		GitRepoURL:   repoInfo.GitRepoURL,
		ValueFiles:   repoInfo.ValueFiles,
		RegionEntity: regionEntity,
		Namespace:    envValue.Namespace,
	}); err != nil {
		return nil, err
	}

	// 9. reset cluster status
	if cluster.Status == common.ClusterStatusFreed {
		cluster.Status = common.ClusterStatusEmpty
		cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
		if err != nil {
			return nil, err
		}
	}

	// 10. deploy cluster in cd and update status
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: cluster.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    masterRevision,
	}); err != nil {
		return nil, err
	}
	if err := c.updatePRStatus(ctx, prmodels.ActionRollback, prCreated.ID, prmodels.StatusOK, masterRevision); err != nil {
		return nil, err
	}

	// 11. record event
	if _, err := c.eventMgr.CreateEvent(ctx, &eventmodels.Event{
		EventSummary: eventmodels.EventSummary{
			ResourceType: common.ResourceCluster,
			EventType:    eventmodels.ClusterRollbacked,
			ResourceID:   cluster.ID,
		},
	}); err != nil {
		log.Warningf(ctx, "failed to create event, err: %s", err.Error())
	}

	return &PipelinerunIDResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}

func (c *controller) Next(ctx context.Context, clusterID uint) (err error) {
	const op = "cluster controller: next"
	defer wlog.Start(ctx, op).StopPrint()

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return err
	}

	return c.cd.Next(ctx, &cd.ClusterNextParams{
		Environment: cluster.EnvironmentName,
		Cluster:     cluster.Name,
	})
}

func (c *controller) Promote(ctx context.Context, clusterID uint) (err error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return perror.WithMessagef(err, "failed to get cluster by id: %d", clusterID)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return perror.WithMessagef(err, "failed to get application by id: %d", cluster.ApplicationID)
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return perror.WithMessage(err, "failed to get env value")
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return perror.WithMessagef(err, "failed to get region by name: %s", cluster.RegionName)
	}

	param := cd.ClusterPromoteParams{
		Namespace:    envValue.Namespace,
		Cluster:      cluster.Name,
		RegionEntity: regionEntity,
		Environment:  cluster.EnvironmentName,
	}
	return c.cd.Promote(ctx, &param)
}

func (c *controller) Pause(ctx context.Context, clusterID uint) (err error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return perror.WithMessagef(err, "failed to get cluster by id: %d", clusterID)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return perror.WithMessagef(err, "failed to get application by id: %d", cluster.ApplicationID)
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return perror.WithMessage(err, "failed to get env value")
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return perror.WithMessagef(err, "failed to get region by name: %s", cluster.RegionName)
	}

	param := cd.ClusterPauseParams{
		Namespace:    envValue.Namespace,
		Cluster:      cluster.Name,
		RegionEntity: regionEntity,
		Environment:  cluster.EnvironmentName,
	}
	return c.cd.Pause(ctx, &param)
}

func (c *controller) Resume(ctx context.Context, clusterID uint) (err error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return perror.WithMessagef(err, "failed to get cluster by id: %d", clusterID)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return perror.WithMessagef(err, "failed to get application by id: %d", cluster.ApplicationID)
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return perror.WithMessage(err, "failed to get env value")
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return perror.WithMessagef(err, "failed to get region by name: %s", cluster.RegionName)
	}

	param := cd.ClusterResumeParams{
		Namespace:    envValue.Namespace,
		Cluster:      cluster.Name,
		RegionEntity: regionEntity,
		Environment:  cluster.EnvironmentName,
	}
	return c.cd.Resume(ctx, &param)
}

// Deprecated
func (c *controller) Online(ctx context.Context, clusterID uint, r *ExecRequest) (_ ExecResponse, err error) {
	const op = "cluster controller: online"
	defer wlog.Start(ctx, op).StopPrint()

	return c.exec(ctx, clusterID, r, c.cd.Online)
}

// Deprecated
func (c *controller) Offline(ctx context.Context, clusterID uint, r *ExecRequest) (_ ExecResponse, err error) {
	const op = "cluster controller: offline"
	defer wlog.Start(ctx, op).StopPrint()

	return c.exec(ctx, clusterID, r, c.cd.Offline)
}

func (c *controller) exec(ctx context.Context, clusterID uint,
	r *ExecRequest, execFunc cd.ExecFunc) (_ ExecResponse, err error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}

	execResp, err := execFunc(ctx, &cd.ExecParams{
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
		RegionEntity: regionEntity,
		Namespace:    envValue.Namespace,
		PodList:      r.PodList,
	})
	if err != nil {
		return nil, err
	}
	return ofExecResp(execResp), nil
}

func (c *controller) Exec(ctx context.Context, clusterID uint, r *ExecRequest) (_ ExecResponse, err error) {
	const op = "cluster controller: exec"
	defer wlog.Start(ctx, op).StopPrint()

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}

	params := &cd.ExecParams{
		Commands:     r.Commands,
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
		RegionEntity: regionEntity,
		Namespace:    envValue.Namespace,
		PodList:      r.PodList,
	}

	resp, err := c.cd.Exec(ctx, params)
	if err != nil {
		return nil, err
	}

	return ofExecResp(resp), nil
}

func (c *controller) DeleteClusterPods(ctx context.Context, clusterID uint, podName []string) (BatchResponse, error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	result, err := c.cd.DeletePods(ctx, &cd.DeletePodsParams{
		Namespace:    envValue.Namespace,
		RegionEntity: regionEntity,
		Pods:         podName,
	})
	if err != nil {
		return nil, err
	}

	return ofBatchResp(result), nil
}

func (c *controller) GetGrafanaDashBoard(ctx context.Context, clusterID uint) (*GetGrafanaDashboardsResponse, error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}

	dashboards, err := c.grafanaService.ListDashboards(ctx)
	if err != nil {
		return nil, err
	}

	return &GetGrafanaDashboardsResponse{
		Host: c.grafanaConfig.Host,
		Params: map[string]string{
			"kiosk":           "iframe",
			"theme":           "light",
			"var-datasource":  cluster.RegionName,
			"var-namespace":   envValue.Namespace,
			"var-application": application.Name,
			"var-cluster":     cluster.Name,
		},
		Dashboards: dashboards,
	}, nil
}

func (c *controller) updatePRStatus(ctx context.Context, action string, prID uint,
	pState prmodels.PipelineStatus, revision string) error {
	var err error
	if pState != prmodels.StatusOK {
		err = c.pipelinerunMgr.UpdateStatusByID(ctx, prID, pState)
	} else {
		finishedAt := time.Now()
		err = c.pipelinerunMgr.UpdateResultByID(ctx, prID, &prmodels.Result{
			Result:     string(pState),
			FinishedAt: &finishedAt,
		})
	}
	if err != nil {
		log.Errorf(ctx, "UpdateStatusByID error, pr = %d, status = %s, err = %v",
			prID, pState, err)
		return err
	}
	log.Infof(ctx, "%s status, pr = %d, status =  %s, revision = %s",
		action, prID, pState, revision)
	return nil
}

func (c *controller) updateTagsAndTemplateFromFile(ctx context.Context,
	application *amodels.Application, cluster *cmodels.Cluster) error {
	files, err := c.clusterGitRepo.GetClusterValueFiles(ctx, application.Name, cluster.Name)
	if err != nil {
		return err
	}

	templateFromFile, err := c.clusterGitRepo.GetClusterTemplate(ctx, application.Name, cluster.Name)
	if err != nil {
		return err
	}
	release, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx,
		templateFromFile.Name, templateFromFile.Release)
	if err != nil {
		return err
	}

	// update template
	_, err = c.clusterMgr.UpdateTemplateByID(ctx, cluster.ID, release.TemplateName, release.Name)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.FileName == common.GitopsFileTags {
			midMap := file.Content[release.ChartName].(map[string]interface{})
			tagsMap := midMap[common.GitopsKeyTags].(map[string]interface{})
			tags := make([]*tmodels.Tag, 0, len(tagsMap))
			for k, v := range tagsMap {
				value, ok := v.(string)
				if !ok {
					continue
				}
				tags = append(tags, &tmodels.Tag{
					ResourceID:   cluster.ID,
					ResourceType: common.ResourceCluster,
					Key:          k,
					Value:        value,
				})
			}
			return c.tagMgr.UpsertByResourceTypeID(ctx, common.ResourceCluster, cluster.ID, tags)
		}
	}
	return nil
}

func (c *controller) mergeDefaultBranchIntoGitOps(ctx context.Context, application, cluster string) error {
	gitOpsBranch := gitrepo.GitOpsBranch
	defaultBranch := c.clusterGitRepo.DefaultBranch()
	diff, err := c.clusterGitRepo.CompareConfig(ctx, application, cluster,
		&gitOpsBranch, &defaultBranch)
	if err != nil {
		return err
	}
	if diff != "" {
		_, err = c.clusterGitRepo.MergeBranch(ctx, application,
			cluster, defaultBranch, gitOpsBranch, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
