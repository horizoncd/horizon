package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	clustercommon "g.hz.netease.com/horizon/pkg/cluster/common"
	perror "g.hz.netease.com/horizon/pkg/errors"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

const (
	QueryPodsMetric = "kube_pod_container_info{namespace=\"%s\",pod=~\"%s.*\"}"
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
	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}

	// 3. update restartTime in git repo, and return the newest commit
	commit, err := c.clusterGitRepo.UpdateRestartTime(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}
	updatePRStatus := func(pState prmodels.PipelineStatus, revision string) error {
		if err = c.pipelinerunMgr.UpdateStatusByID(ctx, prCreated.ID, pState); err != nil {
			log.Errorf(ctx, "UpdateStatusByID error, pr = %d, status = %s, err = %v",
				prCreated.ID, pState, err)
			return err
		}
		log.Infof(ctx, "Restart status, pr = %d, status = %s, revision = %s",
			prCreated.ID, pState, revision)
		return nil
	}
	if err := c.pipelinerunMgr.UpdateConfigCommitByID(ctx, prCreated.ID, commit); err != nil {
		log.Errorf(ctx, "UpdateConfigCommitByID error, pr = %d, commit = %s, err = %v",
			prCreated.ID, commit, err)
	}
	if err := updatePRStatus(prmodels.StatusMerged, commit); err != nil {
		return nil, err
	}

	// 5. deploy cluster in cd system
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: cluster.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    commit,
	}); err != nil {
		return nil, err
	}
	log.Infof(ctx, "Restart Deployed, pr = %d, commit = %s", prCreated.ID, commit)

	// 6. update status
	if err := updatePRStatus(prmodels.StatusOK, commit); err != nil {
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
		if po == nil || po.Image == nil || *po.Image == "" {
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
	if diff == "" && cluster.Status != clustercommon.StatusFreed {
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
		commit, err = c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name, prCreated.ID)
		if err != nil {
			return nil, err
		}
	}
	updatePRStatus := func(pState prmodels.PipelineStatus, revision string) error {
		if err = c.pipelinerunMgr.UpdateStatusByID(ctx, prCreated.ID, pState); err != nil {
			log.Errorf(ctx, "UpdateStatusByID error, pr = %d, status = %s, err = %v",
				prCreated.ID, pState, err)
			return err
		}
		log.Infof(ctx, "Deploy status, pr = %d, status =  %s, revision = %s",
			prCreated.ID, pState, revision)
		return nil
	}
	if err := updatePRStatus(prmodels.StatusMerged, commit); err != nil {
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
		Environment:   cluster.EnvironmentName,
		Cluster:       cluster.Name,
		GitRepoSSHURL: repoInfo.GitRepoSSHURL,
		ValueFiles:    repoInfo.ValueFiles,
		RegionEntity:  regionEntity,
		Namespace:     envValue.Namespace,
	}); err != nil {
		return nil, err
	}

	// 6. reset cluster status
	if cluster.Status == clustercommon.StatusFreed {
		cluster.Status = clustercommon.StatusEmpty
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
	if err := updatePRStatus(prmodels.StatusOK, commit); err != nil {
		return nil, err
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
			"the pipelinerun with id: %v can not be rollbacked", r.PipelinerunID)
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

	// 4. rollback cluster config in git repo and update status
	newConfigCommit, err := c.clusterGitRepo.Rollback(ctx, application.Name, cluster.Name, pipelinerun.ConfigCommit)
	if err != nil {
		return nil, err
	}
	updatePRStatus := func(pState prmodels.PipelineStatus, revision string) error {
		if err = c.pipelinerunMgr.UpdateStatusByID(ctx, prCreated.ID, pState); err != nil {
			log.Errorf(ctx, "UpdateStatusByID error, pr = %d, status = %s, err = %v",
				prCreated.ID, pState, err)
			return err
		}
		log.Infof(ctx, "Rollback status, pr = %d, status =  %s, revision = %s",
			prCreated.ID, pState, revision)
		return nil
	}
	if err := updatePRStatus(prmodels.StatusCommitted, newConfigCommit); err != nil {
		return nil, err
	}

	// 5. merge branch & update config commit and status
	masterRevision, err := c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name, prCreated.ID)
	if err != nil {
		return nil, err
	}
	if err := c.pipelinerunMgr.UpdateConfigCommitByID(ctx, prCreated.ID, masterRevision); err != nil {
		log.Errorf(ctx, "UpdateConfigCommitByID error, pr = %d, commit = %s, err = %v",
			prCreated.ID, masterRevision, err)
	}
	if err := updatePRStatus(prmodels.StatusMerged, masterRevision); err != nil {
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
		Environment:   cluster.EnvironmentName,
		Cluster:       cluster.Name,
		GitRepoSSHURL: repoInfo.GitRepoSSHURL,
		ValueFiles:    repoInfo.ValueFiles,
		RegionEntity:  regionEntity,
		Namespace:     envValue.Namespace,
	}); err != nil {
		return nil, err
	}

	// 9. reset cluster status
	if cluster.Status == clustercommon.StatusFreed {
		cluster.Status = clustercommon.StatusEmpty
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
	if err := updatePRStatus(prmodels.StatusOK, masterRevision); err != nil {
		return nil, err
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

func (c *controller) Online(ctx context.Context, clusterID uint, r *ExecRequest) (_ ExecResponse, err error) {
	const op = "cluster controller: online"
	defer wlog.Start(ctx, op).StopPrint()

	return c.exec(ctx, clusterID, r, c.cd.Online)
}

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

func (c *controller) GetDashboard(ctx context.Context, clusterID uint) (*GetDashboardResponse, error) {
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

	grafanaURL, ok := c.grafanaMapper[envValue.Region]
	if !ok {
		return nil, perror.Wrap(herrors.ErrGrafanaNotSupport,
			"grafana does not support this region")
	}

	getDashboardResp := &GetDashboardResponse{
		Basic:     fmt.Sprintf(grafanaURL.BasicDashboard, envValue.Namespace, cluster.Name),
		Container: fmt.Sprintf(grafanaURL.ContainerDashboard, envValue.Namespace, cluster.Name),
	}

	// TODO(tom): special dashboard about same template should be placed in the horizon template
	// get serverless dashboard
	if cluster.Template == ServerlessTemplateName {
		getDashboardResp.Serverless = fmt.Sprintf(grafanaURL.ServerlessDashboard, cluster.Name)
	}

	// get memcached dashboard
	clusterFiles, err := c.clusterGitRepo.GetCluster(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}
	if memcached, ok := clusterFiles.ApplicationJSONBlob["memcached"]; ok {
		blob, err := json.Marshal(memcached)
		if err != nil {
			return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}

		type MemcachedSchema struct {
			Enabled bool `json:"enabled"`
		}
		var memcachedVal MemcachedSchema
		err = json.Unmarshal(blob, &memcachedVal)
		if err != nil {
			return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}

		if memcachedVal.Enabled {
			getDashboardResp.Memcached = fmt.Sprintf(grafanaURL.MemcachedDashboard, envValue.Namespace, cluster.Name)
		}
	}
	return getDashboardResp, nil
}

func (c *controller) GetClusterPods(ctx context.Context, clusterID uint, start, end int64) (
	*GetClusterPodsResponse, error) {
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

	grafanaURL, ok := c.grafanaMapper[envValue.Region]
	if !ok {
		return nil, perror.Wrap(herrors.ErrGrafanaNotSupport,
			"grafana does not support this region")
	}

	u := url.Values{}
	match := fmt.Sprintf(QueryPodsMetric, envValue.Namespace, cluster.Name)
	u.Set("match[]", match)
	u.Set("start", strconv.FormatInt(start, 10))
	u.Set("end", strconv.FormatInt(end, 10))

	queryURL := grafanaURL.QuerySeries + "?" + u.Encode()

	resp, err := http.Get(queryURL)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed, err.Error())
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, perror.Wrap(herrors.ErrHTTPRespNotAsExpected,
			"grafana query series interface return fail")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrReadFailed,
			"failed to read http response body")
	}

	var result *QueryPodsSeriesResult
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	if result.Status != "success" {
		return nil, perror.Wrap(herrors.ErrHTTPRespNotAsExpected,
			"grafana query series interface return fail")
	}

	return &GetClusterPodsResponse{
		Pods: removeDuplicatePods(result.Data),
	}, nil
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

func removeDuplicatePods(pods []KubePodInfo) []KubePodInfo {
	set := make(map[string]struct{}, len(pods))
	j := 0
	for _, v := range pods {
		key := v.Pod + v.Container
		_, ok := set[key]
		if ok {
			continue
		}
		set[key] = struct{}{}
		pods[j] = v
		j++
	}

	return pods[:j]
}
