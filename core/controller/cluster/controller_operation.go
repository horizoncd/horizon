// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	amodels "github.com/horizoncd/horizon/pkg/application/models"
	"github.com/horizoncd/horizon/pkg/cd"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	cmodels "github.com/horizoncd/horizon/pkg/cluster/models"
	"github.com/horizoncd/horizon/pkg/cluster/tekton"
	perror "github.com/horizoncd/horizon/pkg/errors"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	prmodels "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	tmodels "github.com/horizoncd/horizon/pkg/tag/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
	tokensvc "github.com/horizoncd/horizon/pkg/token/service"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	if err := c.updatePipelineRunStatus(ctx,
		prmodels.ActionRestart, prCreated.ID, prmodels.StatusMerged, commit); err != nil {
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
	if err := c.updatePipelineRunStatus(ctx, prmodels.ActionRestart, prCreated.ID, prmodels.StatusOK, commit); err != nil {
		return nil, err
	}

	// 5. record event
	if _, err := c.eventMgr.CreateEvent(ctx, &eventmodels.Event{
		EventSummary: eventmodels.EventSummary{
			ResourceType: common.ResourceCluster,
			EventType:    eventmodels.ClusterRestarted,
			ResourceID:   cluster.ID,
		},
	}); err != nil {
		log.Warningf(ctx, "failed to create event, err: %s", err.Error())
	}

	return &PipelinerunIDResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}

func (c *controller) Deploy(ctx context.Context, clusterID uint,
	r *DeployRequest) (_ *PipelinerunIDResponse, err error) {
	const op = "cluster controller: deploy"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get models and do some validation
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
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
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}
	configCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}
	codeCommitID := cluster.GitRef
	imageURL := cluster.Image

	if cluster.GitURL != "" {
		err = c.checkAllowDeploy(ctx, application, cluster, clusterFiles, configCommit)
		if err != nil {
			return nil, err
		}
		commit, err := c.commitGetter.GetCommit(ctx, cluster.GitURL, cluster.GitRefType, cluster.GitRef)
		if err == nil {
			codeCommitID = commit.ID
		}
	} else if cluster.Image != "" {
		imageURL, err = getDeployImage(cluster.Image, r.ImageTag)
	}

	// 2. create pipeline record
	prCreated, err := c.pipelinerunMgr.Create(ctx, &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           prmodels.ActionDeploy,
		Status:           string(prmodels.StatusCreated),
		Title:            r.Title,
		Description:      r.Description,
		GitURL:           cluster.GitURL,
		GitRefType:       cluster.GitRefType,
		GitRef:           cluster.GitRef,
		GitCommit:        codeCommitID,
		ImageURL:         imageURL,
		LastConfigCommit: configCommit.Master,
		ConfigCommit:     configCommit.Gitops,
	})
	if err != nil {
		return nil, err
	}

	// 3. generate a JWT token for tekton callback
	token, err := c.tokenSvc.CreateJWTToken(strconv.Itoa(int(currentUser.GetID())),
		c.tokenConfig.CallbackTokenExpireIn, tokensvc.WithPipelinerunID(prCreated.ID))
	if err != nil {
		return nil, err
	}

	// 4. create pipelinerun in k8s
	prGit := tekton.PipelineRunGit{
		URL:       cluster.GitURL,
		Subfolder: cluster.GitSubfolder,
		Commit:    codeCommitID,
	}
	switch prCreated.GitRefType {
	case codemodels.GitRefTypeTag:
		prGit.Tag = prCreated.GitRef
	case codemodels.GitRefTypeBranch:
		prGit.Branch = prCreated.GitRef
	}
	pipelineJSONBlob := make(map[string]interface{})
	if clusterFiles.PipelineJSONBlob != nil {
		pipelineJSONBlob = clusterFiles.PipelineJSONBlob
	}
	tektonClient, err := c.tektonFty.GetTekton(cluster.EnvironmentName)
	if err != nil {
		return nil, err
	}

	ciEventID, err := tektonClient.CreatePipelineRun(ctx, &tekton.PipelineRun{
		Action:           prmodels.ActionDeploy,
		Application:      application.Name,
		ApplicationID:    application.ID,
		Cluster:          cluster.Name,
		ClusterID:        cluster.ID,
		Environment:      cluster.EnvironmentName,
		Git:              prGit,
		ImageURL:         imageURL,
		Operator:         currentUser.GetEmail(),
		PipelinerunID:    prCreated.ID,
		PipelineJSONBlob: pipelineJSONBlob,
		Region:           cluster.RegionName,
		RegionID:         regionEntity.ID,
		Template:         cluster.Template,
		Token:            token,
	})
	if err != nil {
		return nil, err
	}

	// update event id returned from tekton-trigger EventListener
	log.Infof(ctx, "received event id: %s from tekton-trigger EventListener, pipelinerunID: %d",
		ciEventID, prCreated.ID)
	err = c.pipelinerunMgr.UpdateCIEventIDByID(ctx, prCreated.ID, ciEventID)
	if err != nil {
		return nil, err
	}

	return &PipelinerunIDResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}

func (c *controller) checkAllowDeploy(ctx context.Context,
	application *amodels.Application, cluster *cmodels.Cluster,
	clusterFiles *gitrepo.ClusterFiles, configCommit *gitrepo.ClusterCommit) error {
	// check pipeline output
	if len(clusterFiles.PipelineJSONBlob) > 0 {
		po, err := c.clusterGitRepo.GetPipelineOutput(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil {
			if perror.Cause(err) != herrors.ErrPipelineOutputEmpty {
				return err
			}
			return herrors.ErrShouldBuildDeployFirst
		}
		if po == nil {
			return herrors.ErrShouldBuildDeployFirst
		}
	}

	// check config diffs
	diff, err := c.clusterGitRepo.CompareConfig(ctx, application.Name, cluster.Name,
		&configCommit.Master, &configCommit.Gitops)
	if err != nil {
		return err
	}
	if diff == "" && cluster.Status != common.ClusterStatusFreed {
		return perror.Wrap(herrors.ErrClusterNoChange, "there is no change to deploy")
	}
	return nil
}

func getDeployImage(imageURL, deployTag string) (string, error) {
	imageRef, err := name.ParseReference(imageURL)
	if err != nil {
		return "", perror.Wrapf(herrors.ErrParamInvalid, "invalid image url: %s", imageURL)
	}
	if deployTag != "" {
		return fmt.Sprintf("%s:%s", imageRef.Context().Name(), deployTag), nil
	}
	return imageRef.Name(), nil
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

	// Deprecated: for internal usage
	err = c.checkAndSyncGitOpsBranch(ctx, application.Name, cluster.Name, pipelinerun.ConfigCommit)
	if err != nil {
		return nil, err
	}

	// 4. rollback cluster config in git repo and update status
	newConfigCommit, err := c.clusterGitRepo.Rollback(ctx, application.Name, cluster.Name, pipelinerun.ConfigCommit)
	if err != nil {
		return nil, err
	}
	if err := c.updatePipelineRunStatus(ctx, prmodels.ActionRollback, prCreated.ID, prmodels.StatusCommitted,
		newConfigCommit); err != nil {
		return nil, err
	}

	// 5. merge branch & update config commit and status
	masterRevision, err := c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name,
		gitrepo.GitOpsBranch, c.clusterGitRepo.DefaultBranch(), &prCreated.ID)
	if err != nil {
		return nil, err
	}
	if err := c.pipelinerunMgr.UpdateConfigCommitByID(ctx, prCreated.ID, masterRevision); err != nil {
		log.Errorf(ctx, "UpdateConfigCommitByID error, pr = %d, commit = %s, err = %v",
			prCreated.ID, masterRevision, err)
	}
	if err := c.updatePipelineRunStatus(ctx, prmodels.ActionRollback, prCreated.ID, prmodels.StatusMerged,
		masterRevision); err != nil {
		return nil, err
	}

	// 6. update template and tags in db
	// TODO(zhuxu): remove strong dependencies on db updates, just print an err log when updates fail
	cluster, err = c.updateTemplateAndTagsFromFile(ctx, application, cluster)
	if err != nil {
		return nil, err
	}

	// 7. create cluster in cd system
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

	// 8. reset cluster status
	if cluster.Status == common.ClusterStatusFreed {
		cluster.Status = common.ClusterStatusEmpty
		cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
		if err != nil {
			return nil, err
		}
	}

	// 9. deploy cluster in cd and update status
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: cluster.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    masterRevision,
	}); err != nil {
		return nil, err
	}
	if err := c.updatePipelineRunStatus(ctx,
		prmodels.ActionRollback, prCreated.ID, prmodels.StatusOK, masterRevision); err != nil {
		return nil, err
	}

	// 10. record event
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

func (c *controller) retrieveClusterCtx(ctx context.Context, clusterID uint) (*cmodels.Cluster,
	*amodels.Application, *trmodels.TemplateRelease, *regionmodels.RegionEntity, *gitrepo.EnvValue, error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, nil, nil, nil, nil,
			herrors.NewErrGetFailed(herrors.ClusterInDB, fmt.Sprintf("cluster id: %d", clusterID))
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, nil, nil, nil, nil,
			herrors.NewErrGetFailed(herrors.ApplicationInDB, fmt.Sprintf("application id: %d", cluster.ApplicationID))
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, nil, nil, nil, nil,
			herrors.NewErrGetFailed(herrors.EnvValueInGit,
				fmt.Sprintf("application id: %d, cluster id: %d", cluster.ApplicationID, cluster.ID))
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, nil, nil, nil, nil,
			herrors.NewErrGetFailed(herrors.RegionInDB, fmt.Sprintf("region name: %s", cluster.RegionName))
	}
	return cluster, application, tr, regionEntity, envValue, nil
}

func (c *controller) ExecuteAction(ctx context.Context, clusterID uint,
	action string, gvr schema.GroupVersionResource) error {
	cluster, _, _, regionEntity, envValue, err := c.retrieveClusterCtx(ctx, clusterID)
	if err != nil {
		return err
	}

	return c.k8sutil.ExecuteAction(ctx, &cd.ExecuteActionParams{
		RegionEntity: regionEntity,
		Namespace:    envValue.Namespace,
		Action:       action,
		GVR:          gvr,
		ResourceName: cluster.Name,
	})
}

// onlineCommand the location of online.sh in pod is /home/appops/.probe/online-once.sh
var onlineCommands = []string{"bash", "-c", `
export ONLINE_SHELL="/home/appops/.probe/online-once.sh"
[[ -f "$ONLINE_SHELL" ]] || {
	echo "there is no online config for this cluster." >&2; exit 1
}

bash "$ONLINE_SHELL"
`}

// offlineCommand the location of offline.sh in pod is /home/appops/.probe/offline-once.sh
var offlineCommands = []string{"bash", "-c", `
export OFFLINE_SHELL="/home/appops/.probe/offline-once.sh"
[[ -f "$OFFLINE_SHELL" ]] || {
	echo "there is no offline config for this cluster." >&2; exit 1
}

bash "$OFFLINE_SHELL"
`}

// Deprecated
func (c *controller) Online(ctx context.Context, clusterID uint, r *ExecRequest) (_ ExecResponse, err error) {
	const op = "cluster controller: online"
	defer wlog.Start(ctx, op).StopPrint()

	r.Commands = onlineCommands
	return c.Exec(ctx, clusterID, r)
}

// Deprecated
func (c *controller) Offline(ctx context.Context, clusterID uint, r *ExecRequest) (_ ExecResponse, err error) {
	const op = "cluster controller: offline"
	defer wlog.Start(ctx, op).StopPrint()

	r.Commands = offlineCommands
	return c.Exec(ctx, clusterID, r)
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

	resp, err := c.k8sutil.Exec(ctx, params)
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

	result, err := c.k8sutil.DeletePods(ctx, &cd.DeletePodsParams{
		Namespace:    envValue.Namespace,
		RegionEntity: regionEntity,
		Pods:         podName,
	})
	if err != nil {
		return nil, err
	}

	// 5. record event
	podNameEncodedBts, err := json.Marshal(podName)
	if err != nil {
		log.Warningf(ctx, "failed to marshal podNames: %v", err.Error())
	}
	podNameEncoded := string(podNameEncodedBts)
	if _, err := c.eventMgr.CreateEvent(ctx, &eventmodels.Event{
		EventSummary: eventmodels.EventSummary{
			ResourceType: common.ResourceCluster,
			EventType:    eventmodels.ClusterPodsRescheduled,
			ResourceID:   cluster.ID,
			Extra:        &podNameEncoded,
		},
	}); err != nil {
		log.Warningf(ctx, "failed to create event, err: %s", err.Error())
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

// Deprecated: Upgrade v1 to v2
func (c *controller) Upgrade(ctx context.Context, clusterID uint) error {
	const op = "cluster controller: upgrade to v2"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. validate infos
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return err
	}
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return err
	}
	templateFromFile, err := c.clusterGitRepo.GetClusterTemplate(ctx, application.Name, cluster.Name)
	if err != nil {
		return err
	}

	// 2. match target template
	targetTemplate, ok := c.templateUpgradeMapper[templateFromFile.Name]
	if !ok {
		return perror.Wrapf(herrors.ErrParamInvalid,
			"cluster template %s does not support upgrade", templateFromFile.Name)
	}
	targetRelease, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx,
		targetTemplate.Name, targetTemplate.Release)
	if err != nil {
		return err
	}

	// 3. sync gitops branch if restarts occur
	err = c.syncGitOpsBranch(ctx, application.Name, cluster.Name)
	if err != nil {
		return err
	}

	// 4. upgrade git repo files to v2
	_, err = c.clusterGitRepo.UpgradeCluster(ctx, &gitrepo.UpgradeValuesParam{
		Application:   application.Name,
		Cluster:       cluster.Name,
		Template:      templateFromFile,
		TargetRelease: targetRelease,
		BuildConfig:   &targetTemplate.BuildConfig,
	})
	if err != nil {
		return err
	}

	// 5. update template in db
	// TODO(zhuxu): remove strong dependencies on db updates, just print an err log when updates fail
	cluster.Template = targetRelease.TemplateName
	cluster.TemplateRelease = targetRelease.Name
	_, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
	if err != nil {
		return err
	}
	return nil
}

func (c *controller) updatePipelineRunStatus(ctx context.Context,
	action string, prID uint, pState prmodels.PipelineStatus, revision string) error {
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

// updateTemplateAndTagsFromFile syncs template and tags in db when git repo files are updated
func (c *controller) updateTemplateAndTagsFromFile(ctx context.Context,
	application *amodels.Application, cluster *cmodels.Cluster) (*cmodels.Cluster, error) {
	templateFromFile, err := c.clusterGitRepo.GetClusterTemplate(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}
	cluster.Template = templateFromFile.Name
	cluster.TemplateRelease = templateFromFile.Release
	cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
	if err != nil {
		return nil, err
	}

	files, err := c.clusterGitRepo.GetClusterValueFiles(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.FileName == common.GitopsFileTags {
			release, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
			if err != nil {
				return nil, err
			}
			midMap := file.Content[release.ChartName].(map[string]interface{})
			tagsMap := midMap[common.GitopsKeyTags].(map[string]interface{})
			tags := make([]*tmodels.TagBasic, 0, len(tagsMap))
			for k, v := range tagsMap {
				value, ok := v.(string)
				if !ok {
					continue
				}
				tags = append(tags, &tmodels.TagBasic{
					Key:   k,
					Value: value,
				})
			}
			return cluster, c.tagMgr.UpsertByResourceTypeID(ctx,
				common.ResourceCluster, cluster.ID, tags)
		}
	}
	return cluster, nil
}

func (c *controller) checkAndSyncGitOpsBranch(ctx context.Context, application,
	cluster string, commit string) error {
	changed, err := c.manifestVersionChanged(ctx, application, cluster, commit)
	if err != nil {
		return err
	}
	if changed {
		err = c.syncGitOpsBranch(ctx, application, cluster)
		if err != nil {
			return err
		}
	}
	return nil
}

// Deprecated: for internal usage
// manifestVersionChanged determines whether manifest version is changed
func (c *controller) manifestVersionChanged(ctx context.Context, application,
	cluster string, commit string) (bool, error) {
	currentManifest, err1 := c.clusterGitRepo.GetManifest(ctx, application, cluster, nil)
	if err1 != nil {
		if _, ok := perror.Cause(err1).(*herrors.HorizonErrNotFound); !ok {
			log.Errorf(ctx, "get cluster manifest error, err = %s", err1.Error())
			return false, err1
		}
	}
	targetManifest, err2 := c.clusterGitRepo.GetManifest(ctx, application, cluster, &commit)
	if err2 != nil {
		if _, ok := perror.Cause(err2).(*herrors.HorizonErrNotFound); !ok {
			log.Errorf(ctx, "get cluster manifest error, err = %s", err2.Error())
			return false, err2
		}
	}
	if err1 != nil && err2 != nil {
		// manifest does not exist in both revisions
		return false, nil
	}
	if err1 != nil || err2 != nil {
		// One exists and the other does not exist in two revisions
		return true, nil
	}
	return currentManifest.Version != targetManifest.Version, nil
}

// Deprecated: for internal usage
// syncGitOpsBranch syncs gitOps branch with default branch to avoid merge conflicts.
// Restart updates time in restart.yaml in default branch. When other actions update
// template prefix in gitOps branch, there are merge conflicts in restart.yaml because
// usual context lines of 'git diff' are three. Ref: https://git-scm.com/docs/git-diff
// For example:
//
//	<<<<<<< HEAD
//	javaapp:
//	  restartTime: "2025-02-19 10:24:52"
//	=======
//	rollout:
//	  restartTime: "2025-02-14 12:12:07"
//	>>>>>>> gitops
func (c *controller) syncGitOpsBranch(ctx context.Context, application, cluster string) error {
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
