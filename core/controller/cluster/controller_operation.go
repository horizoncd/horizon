package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	clustercommon "g.hz.netease.com/horizon/pkg/cluster/common"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

const (
	QueryPodsMetric = "kube_pod_info{namespace=\"%s\",pod=~\"%s.*\"}"
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
	var commit string
	if diff == "" {
		if cluster.Status != clustercommon.StatusFreed {
			return nil, errors.E(op, http.StatusBadRequest, errors.ErrorCode("NoChange"), "there is no change to deploy")
		}
		// freed cluster is allowed to deploy without diff
		commitInfo, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
		if err != nil {
			return nil, errors.E(op, err)
		}
		commit = commitInfo.Master
	} else {
		// 2. merge branch
		commit, err = c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	// 3. create cluster in cd system
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, er.RegionName)
	if err != nil {
		return nil, errors.E(op, err)
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}
	repoInfo := c.clusterGitRepo.GetRepoInfo(ctx, application.Name, cluster.Name)
	if err := c.cd.CreateCluster(ctx, &cd.CreateClusterParams{
		Environment:   er.EnvironmentName,
		Cluster:       cluster.Name,
		GitRepoSSHURL: repoInfo.GitRepoSSHURL,
		ValueFiles:    repoInfo.ValueFiles,
		RegionEntity:  regionEntity,
		Namespace:     envValue.Namespace,
	}); err != nil {
		return nil, errors.E(op, err)
	}

	// 4. reset cluster status
	if cluster.Status == clustercommon.StatusFreed {
		cluster.Status = clustercommon.StatusEmpty
		cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	// 5. deploy cluster in cd system
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: er.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    commit,
	}); err != nil {
		return nil, errors.E(op, err)
	}

	timeNow := time.Now()
	// 6. add pipelinerun in db
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

	// 5. create cluster in cd system
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, er.RegionName)
	if err != nil {
		return nil, errors.E(op, err)
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}
	repoInfo := c.clusterGitRepo.GetRepoInfo(ctx, application.Name, cluster.Name)
	if err := c.cd.CreateCluster(ctx, &cd.CreateClusterParams{
		Environment:   er.EnvironmentName,
		Cluster:       cluster.Name,
		GitRepoSSHURL: repoInfo.GitRepoSSHURL,
		ValueFiles:    repoInfo.ValueFiles,
		RegionEntity:  regionEntity,
		Namespace:     envValue.Namespace,
	}); err != nil {
		return nil, errors.E(op, err)
	}

	// 6. reset cluster status
	if cluster.Status == clustercommon.StatusFreed {
		cluster.Status = clustercommon.StatusEmpty
		cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	// 7. deploy cluster in cd
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: er.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    masterRevision,
	}); err != nil {
		return nil, errors.E(op, err)
	}

	timeNow := time.Now()
	// 8. add pipelinerun in db
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

func (c *controller) SkipAllSteps(ctx context.Context, clusterID uint) (err error) {
	const op = "cluster controller: skip all steps"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return errors.E(op, err)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return errors.E(op, err)
	}

	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return errors.E(op, err)
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return errors.E(op, err)
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, er.RegionName)
	if err != nil {
		return errors.E(op, err)
	}

	param := cd.ClusterSkipAllStepsParams{
		Namespace:    envValue.Namespace,
		Cluster:      cluster.Name,
		RegionEntity: regionEntity,
		Environment:  er.EnvironmentName,
	}
	return c.cd.SkipAllSteps(ctx, &param)
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

func (c *controller) GetDashboard(ctx context.Context, clusterID uint) (*GetDashboardResponse, error) {
	const op = "cluster controller: get dashboard"
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}

	grafanaURL, ok := c.grafanaMapper[envValue.Region]
	if !ok {
		return nil, errors.E(op, fmt.Errorf("grafana does not support this region"))
	}

	getDashboardResp := &GetDashboardResponse{
		Basic: fmt.Sprintf(grafanaURL.BasicDashboard, envValue.Namespace, cluster.Name),
	}

	// TODO(tom): special dashboard about same template should be placed in the horizon template
	// get serverless dashboard
	if cluster.Template == ServerlessTemplateName {
		getDashboardResp.Serverless = fmt.Sprintf(grafanaURL.ServerlessDashboard, cluster.Name)
	}

	// get memcached dashboard
	clusterFiles, err := c.clusterGitRepo.GetCluster(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if memcached, ok := clusterFiles.ApplicationJSONBlob["memcached"]; ok {
		blob, err := json.Marshal(memcached)
		if err != nil {
			return nil, errors.E(op, err)
		}

		type MemcachedSchema struct {
			Enabled bool `json:"enabled"`
		}
		var memcachedVal MemcachedSchema
		err = json.Unmarshal(blob, &memcachedVal)
		if err != nil {
			return nil, errors.E(op, err)
		}

		if memcachedVal.Enabled {
			getDashboardResp.Memcached = fmt.Sprintf(grafanaURL.MemcachedDashboard, envValue.Namespace, cluster.Name)
		}
	}
	return getDashboardResp, nil
}

func (c *controller) GetClusterPods(ctx context.Context, clusterID uint, start, end int64) (
	*GetClusterPodsResponse, error) {
	const op = "cluster controller: get cluster pods"
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}

	grafanaURL, ok := c.grafanaMapper[envValue.Region]
	if !ok {
		return nil, errors.E(op, fmt.Errorf("grafana does not support this region"))
	}

	u := url.Values{}
	match := fmt.Sprintf(QueryPodsMetric, envValue.Namespace, cluster.Name)
	u.Set("match[]", match)
	u.Set("start", strconv.FormatInt(start, 10))
	u.Set("end", strconv.FormatInt(end, 10))

	queryURL := grafanaURL.QuerySeries + "?" + u.Encode()

	resp, err := http.Get(queryURL)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.E(op, resp.StatusCode, "grafana query series interface return fail")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.E(op, err)
	}

	var result *QueryPodsSeriesResult
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if result.Status != "success" {
		return nil, errors.E(op, "grafana query series interface return fail")
	}

	return &GetClusterPodsResponse{
		Pods: result.Data,
	}, nil
}
