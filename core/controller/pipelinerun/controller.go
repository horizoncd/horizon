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

package pipelinerun

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/config"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	appmanager "github.com/horizoncd/horizon/pkg/application/manager"
	"github.com/horizoncd/horizon/pkg/cluster/code"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	"github.com/horizoncd/horizon/pkg/cluster/tekton"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/collector"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/factory"
	"github.com/horizoncd/horizon/pkg/config/token"
	envmanager "github.com/horizoncd/horizon/pkg/environment/manager"
	perror "github.com/horizoncd/horizon/pkg/errors"
	membermanager "github.com/horizoncd/horizon/pkg/member"
	"github.com/horizoncd/horizon/pkg/param"
	prmanager "github.com/horizoncd/horizon/pkg/pr/manager"
	prmodels "github.com/horizoncd/horizon/pkg/pr/models"
	prservice "github.com/horizoncd/horizon/pkg/pr/service"
	regionmanager "github.com/horizoncd/horizon/pkg/region/manager"
	trmanager "github.com/horizoncd/horizon/pkg/templaterelease/manager"
	tokensvc "github.com/horizoncd/horizon/pkg/token/service"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	"github.com/horizoncd/horizon/pkg/util/errors"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

type Controller interface {
	GetPipelinerunLog(ctx context.Context, pipelinerunID uint) (*collector.Log, error)
	GetClusterLatestLog(ctx context.Context, clusterID uint) (*collector.Log, error)
	GetDiff(ctx context.Context, pipelinerunID uint) (*GetDiffResponse, error)
	GetPipelinerun(ctx context.Context, pipelinerunID uint) (*prmodels.PipelineBasic, error)
	ListPipelineruns(ctx context.Context, clusterID uint, canRollback bool,
		query q.Query) (int, []*prmodels.PipelineBasic, error)
	StopPipelinerun(ctx context.Context, pipelinerunID uint) error
	StopPipelinerunForCluster(ctx context.Context, clusterID uint) error

	CreateCheck(ctx context.Context, check *prmodels.Check) (*prmodels.Check, error)
	GetCheckRunByID(ctx context.Context, checkRunID uint) (*prmodels.CheckRun, error)
	UpdateCheckRunByID(ctx context.Context, checkRunID uint, request *CreateOrUpdateCheckRunRequest) error

	ListMessagesByPipelinerun(ctx context.Context, pipelinerunID uint, query *q.Query) (int, []*prmodels.PRMessage, error)
	// Execute runs a pipelineRun only if its state is ready.
	Execute(ctx context.Context, pipelinerunID uint, force bool) error
	// Cancel withdraws a pipelineRun only if its state is pending.
	Cancel(ctx context.Context, pipelinerunID uint) error

	ListCheckRuns(ctx context.Context, q *q.Query) ([]*prmodels.CheckRun, error)
	CreateCheckRun(ctx context.Context, pipelineRunID uint,
		request *CreateOrUpdateCheckRunRequest) (*prmodels.CheckRun, error)
	ListPRMessages(ctx context.Context, pipelineRunID uint, q *q.Query) (int, []*PrMessage, error)
	CreatePRMessage(ctx context.Context, pipelineRunID uint, request *CreatePrMessageRequest) (*prmodels.PRMessage, error)
}

type controller struct {
	prMgr              *prmanager.PRManager
	appMgr             appmanager.Manager
	clusterMgr         clustermanager.Manager
	envMgr             envmanager.Manager
	prSvc              *prservice.Service
	regionMgr          regionmanager.Manager
	tektonFty          factory.Factory
	tokenSvc           tokensvc.Service
	tokenConfig        token.Config
	memberMgr          membermanager.Manager
	templateReleaseMgr trmanager.Manager
	commitGetter       code.GitGetter
	clusterGitRepo     gitrepo.ClusterGitRepo
	userMgr            usermanager.Manager
}

var _ Controller = (*controller)(nil)

func NewController(config *config.Config, param *param.Param) Controller {
	return &controller{
		prMgr:              param.PRMgr,
		prSvc:              param.PRService,
		clusterMgr:         param.ClusterMgr,
		envMgr:             param.EnvMgr,
		tektonFty:          param.TektonFty,
		tokenSvc:           param.TokenSvc,
		tokenConfig:        config.TokenConfig,
		commitGetter:       param.GitGetter,
		appMgr:             param.ApplicationMgr,
		memberMgr:          param.MemberMgr,
		regionMgr:          param.RegionMgr,
		clusterGitRepo:     param.ClusterGitRepo,
		userMgr:            param.UserMgr,
		templateReleaseMgr: param.TemplateReleaseMgr,
	}
}

func (c *controller) GetPipelinerunLog(ctx context.Context, pipelinerunID uint) (_ *collector.Log, err error) {
	const op = "pipelinerun controller: get pipelinerun log"
	defer wlog.Start(ctx, op).StopPrint()

	pr, err := c.prMgr.PipelineRun.GetByID(ctx, pipelinerunID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	cluster, err := c.clusterMgr.GetByID(ctx, pr.ClusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// only builddeploy and deploy have logs
	if pr.Action != prmodels.ActionBuildDeploy && pr.Action != prmodels.ActionDeploy {
		return nil, errors.E(op, fmt.Errorf("%v action has no log", pr.Action))
	}

	return c.getPipelinerunLog(ctx, pr, cluster.EnvironmentName)
}

func (c *controller) GetClusterLatestLog(ctx context.Context, clusterID uint) (_ *collector.Log, err error) {
	const op = "pipelinerun controller: get cluster latest log"
	defer wlog.Start(ctx, op).StopPrint()

	pr, err := c.prMgr.PipelineRun.GetLatestByClusterIDAndActions(ctx, clusterID,
		prmodels.ActionBuildDeploy, prmodels.ActionDeploy)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if pr == nil {
		return nil, errors.E(op, fmt.Errorf("no pipelinerun with log"))
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return c.getPipelinerunLog(ctx, pr, cluster.EnvironmentName)
}

func (c *controller) getPipelinerunLog(ctx context.Context, pr *prmodels.Pipelinerun,
	environment string) (_ *collector.Log, err error) {
	const op = "pipeline controller: get pipelinerun log"
	defer wlog.Start(ctx, op).StopPrint()

	tektonCollector, err := c.tektonFty.GetTektonCollector(environment)
	if err != nil {
		return nil, perror.WithMessagef(err, "failed to get tekton collector for %s", environment)
	}

	return tektonCollector.GetPipelineRunLog(ctx, pr)
}

func (c *controller) GetDiff(ctx context.Context, pipelinerunID uint) (_ *GetDiffResponse, err error) {
	const op = "pipelinerun controller: get pipelinerun diff"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get pipeline
	pipelinerun, err := c.prMgr.PipelineRun.GetByID(ctx, pipelinerunID)
	if err != nil {
		return nil, err
	}

	// 2. get cluster and application
	cluster, err := c.clusterMgr.GetByID(ctx, pipelinerun.ClusterID)
	if err != nil {
		return nil, err
	}
	application, err := c.appMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 3. get code diff
	var codeDiff *CodeInfo
	if pipelinerun.GitURL != "" && pipelinerun.GitCommit != "" &&
		pipelinerun.GitRef != "" {
		commit, err := c.commitGetter.GetCommit(ctx, pipelinerun.GitURL,
			codemodels.GitRefTypeCommit, pipelinerun.GitCommit)
		if err != nil {
			return nil, err
		}
		historyLink, err := c.commitGetter.GetCommitHistoryLink(pipelinerun.GitURL, pipelinerun.GitCommit)
		if err != nil {
			return nil, err
		}
		codeDiff = &CodeInfo{
			CommitID:  pipelinerun.GitCommit,
			CommitMsg: commit.Message,
			Link:      historyLink,
		}
		switch pipelinerun.GitRefType {
		case codemodels.GitRefTypeTag:
			codeDiff.Tag = pipelinerun.GitRef
		case codemodels.GitRefTypeBranch:
			codeDiff.Branch = pipelinerun.GitRef
		}
	}

	// 4. get config diff
	var configDiff *ConfigDiff
	if pipelinerun.ConfigCommit != "" && pipelinerun.LastConfigCommit != "" {
		diff, err := c.clusterGitRepo.CompareConfig(ctx, application.Name, cluster.Name,
			&pipelinerun.LastConfigCommit, &pipelinerun.ConfigCommit)
		if err != nil {
			return nil, err
		}
		configDiff = &ConfigDiff{
			From: pipelinerun.LastConfigCommit,
			To:   pipelinerun.ConfigCommit,
			Diff: diff,
		}
	}

	return &GetDiffResponse{
		CodeInfo:   codeDiff,
		ConfigDiff: configDiff,
	}, nil
}

func (c *controller) GetPipelinerun(ctx context.Context, pipelineID uint) (_ *prmodels.PipelineBasic, err error) {
	const op = "pipelinerun controller: get pipelinerun basic"
	defer wlog.Start(ctx, op).StopPrint()

	pipelinerun, err := c.prMgr.PipelineRun.GetByID(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	firstCanRollbackPipelinerun, err := c.prMgr.PipelineRun.GetFirstCanRollbackPipelinerun(ctx, pipelinerun.ClusterID)
	if err != nil {
		return nil, err
	}

	return c.prSvc.OfPipelineBasic(ctx, pipelinerun, firstCanRollbackPipelinerun)
}

func (c *controller) ListPipelineruns(ctx context.Context,
	clusterID uint, canRollback bool, query q.Query) (_ int, _ []*prmodels.PipelineBasic, err error) {
	const op = "pipelinerun controller: list pipelinerun"
	defer wlog.Start(ctx, op).StopPrint()

	totalCount, pipelineruns, err := c.prMgr.PipelineRun.GetByClusterID(ctx, clusterID, canRollback, query)
	if err != nil {
		return 0, nil, err
	}

	// remove the first pipelinerun than can be rollback
	firstCanRollbackPipelinerun, err := c.prMgr.PipelineRun.GetFirstCanRollbackPipelinerun(ctx, clusterID)
	if err != nil {
		return 0, nil, err
	}

	pipelineBasics, err := c.prSvc.OfPipelineBasics(ctx, pipelineruns, firstCanRollbackPipelinerun)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, pipelineBasics, nil
}

func (c *controller) StopPipelinerun(ctx context.Context, pipelinerunID uint) (err error) {
	const op = "pipelinerun controller: stop pipelinerun"
	defer wlog.Start(ctx, op).StopPrint()

	pipelinerun, err := c.prMgr.PipelineRun.GetByID(ctx, pipelinerunID)
	if err != nil {
		return errors.E(op, err)
	}
	if pipelinerun.Status != string(prmodels.StatusCreated) &&
		pipelinerun.Status != string(prmodels.StatusRunning) {
		return errors.E(op, http.StatusBadRequest, errors.ErrorCode("BadRequest"), "pipelinerun is already completed")
	}
	cluster, err := c.clusterMgr.GetByID(ctx, pipelinerun.ClusterID)
	if err != nil {
		return errors.E(op, err)
	}

	tektonClient, err := c.tektonFty.GetTekton(cluster.EnvironmentName)
	if err != nil {
		return errors.E(op, err)
	}

	return tektonClient.StopPipelineRun(ctx, pipelinerun.CIEventID)
}

func (c *controller) StopPipelinerunForCluster(ctx context.Context, clusterID uint) (err error) {
	const op = "pipelinerun controller: stop pipelinerun"
	defer wlog.Start(ctx, op).StopPrint()

	if err != nil {
		return errors.E(op, err)
	}
	// get cluster latest builddeploy pipelinerun
	pipelinerun, err := c.prMgr.PipelineRun.GetLatestByClusterIDAndActions(ctx, clusterID, prmodels.ActionBuildDeploy)

	// if pipelinerun.Status is not created, ignore, and return success
	if pipelinerun.Status != string(prmodels.StatusCreated) &&
		pipelinerun.Status != string(prmodels.StatusRunning) {
		return nil
	}
	cluster, err := c.clusterMgr.GetByID(ctx, pipelinerun.ClusterID)
	if err != nil {
		return errors.E(op, err)
	}

	tektonClient, err := c.tektonFty.GetTekton(cluster.EnvironmentName)
	if err != nil {
		return errors.E(op, err)
	}

	return tektonClient.StopPipelineRun(ctx, pipelinerun.CIEventID)
}

func (c *controller) CreateCheck(ctx context.Context, check *prmodels.Check) (*prmodels.Check, error) {
	const op = "pipelinerun controller: create check"
	defer wlog.Start(ctx, op).StopPrint()

	return c.prMgr.Check.Create(ctx, check)
}

func (c *controller) UpdateCheckRunByID(ctx context.Context, checkRunID uint,
	request *CreateOrUpdateCheckRunRequest) error {
	const op = "pipelinerun controller: update check run"
	defer wlog.Start(ctx, op).StopPrint()

	err := c.prMgr.Check.UpdateByID(ctx, checkRunID, &prmodels.CheckRun{
		Name:      request.Name,
		Status:    prmodels.String2CheckRunStatus(request.Status),
		Message:   request.Message,
		DetailURL: request.DetailURL,
	})
	if err != nil {
		return err
	}
	return c.updatePrStatusByCheckrunID(ctx, checkRunID)
}

func (c *controller) ListMessagesByPipelinerun(ctx context.Context,
	pipelinerunID uint, query *q.Query) (int, []*prmodels.PRMessage, error) {
	const op = "pipelinerun controller: list pr message"
	defer wlog.Start(ctx, op).StopPrint()

	return c.prMgr.Message.List(ctx, pipelinerunID, query)
}

func (c *controller) Execute(ctx context.Context, pipelinerunID uint, force bool) error {
	const op = "pipelinerun controller: execute pipelinerun"
	defer wlog.Start(ctx, op).StopPrint()

	pr, err := c.prMgr.PipelineRun.GetByID(ctx, pipelinerunID)
	if err != nil {
		return err
	}

	if force {
		if pr.Status != string(prmodels.StatusReady) && pr.Status != string(prmodels.StatusPending) {
			return perror.Wrapf(herrors.ErrParamInvalid, "pipelinerun is not ready to execute")
		}
	} else {
		if pr.Status != string(prmodels.StatusReady) {
			return perror.Wrapf(herrors.ErrParamInvalid, "pipelinerun is not ready to execute")
		}
	}

	return c.execute(ctx, pr)
}

func (c *controller) execute(ctx context.Context, pr *prmodels.Pipelinerun) error {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	// 1. get cluster
	cluster, err := c.clusterMgr.GetByID(ctx, pr.ClusterID)
	if err != nil {
		return err
	}

	// 2. get application
	application, err := c.appMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return err
	}

	// 3. generate a JWT token for tekton callback
	token, err := c.tokenSvc.CreateJWTToken(strconv.Itoa(int(currentUser.GetID())),
		c.tokenConfig.CallbackTokenExpireIn, tokensvc.WithPipelinerunID(pr.ID))
	if err != nil {
		return err
	}

	// 4. create pipelinerun in k8s
	tektonClient, err := c.tektonFty.GetTekton(cluster.EnvironmentName)
	if err != nil {
		return err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return err
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return err
	}
	clusterFiles, err := c.clusterGitRepo.GetCluster(ctx,
		application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return err
	}

	prGit := tekton.PipelineRunGit{
		URL:       cluster.GitURL,
		Subfolder: cluster.GitSubfolder,
		Commit:    pr.GitCommit,
	}
	switch pr.GitRefType {
	case codemodels.GitRefTypeTag:
		prGit.Tag = pr.GitRef
	case codemodels.GitRefTypeBranch:
		prGit.Branch = pr.GitRef
	}

	ciEventID, err := tektonClient.CreatePipelineRun(ctx, &tekton.PipelineRun{
		Action:           pr.Action,
		Application:      application.Name,
		ApplicationID:    application.ID,
		Cluster:          cluster.Name,
		ClusterID:        cluster.ID,
		Environment:      cluster.EnvironmentName,
		Git:              prGit,
		ImageURL:         pr.ImageURL,
		Operator:         currentUser.GetEmail(),
		PipelinerunID:    pr.ID,
		PipelineJSONBlob: clusterFiles.PipelineJSONBlob,
		Region:           cluster.RegionName,
		RegionID:         regionEntity.ID,
		Template:         cluster.Template,
		Token:            token,
	})
	if err != nil {
		return err
	}

	// update event id returned from tekton-trigger EventListener
	log.Infof(ctx, "received event id: %s from tekton-trigger EventListener, pipelinerunID: %d", ciEventID, pr.ID)
	err = c.prMgr.PipelineRun.UpdateColumns(ctx, pr.ID, map[string]interface{}{
		"ci_event_id": ciEventID,
		"status":      prmodels.StatusRunning,
		"started_at":  time.Now(),
	})
	if err != nil {
		return err
	}
	err = c.prMgr.PipelineRun.UpdateCIEventIDByID(ctx, pr.ID, ciEventID)
	if err != nil {
		return err
	}

	return nil
}

func (c *controller) Cancel(ctx context.Context, pipelinerunID uint) error {
	const op = "pipelinerun controller: cancel pipelinerun"
	defer wlog.Start(ctx, op).StopPrint()
	pr, err := c.prMgr.PipelineRun.GetByID(ctx, pipelinerunID)
	if err != nil {
		return err
	}

	if pr.Status != string(prmodels.StatusPending) && pr.Status != string(prmodels.StatusReady) {
		return perror.Wrapf(herrors.ErrParamInvalid, "pipelinerun is not pending or ready to cancel")
	}
	return c.prMgr.PipelineRun.UpdateStatusByID(ctx, pipelinerunID, prmodels.StatusCancelled)
}

func (c *controller) ListCheckRuns(ctx context.Context, query *q.Query) ([]*prmodels.CheckRun, error) {
	const op = "pipelinerun controller: list check runs"
	defer wlog.Start(context.Background(), op).StopPrint()
	return c.prMgr.Check.ListCheckRuns(ctx, query)
}

func (c *controller) GetCheckRunByID(ctx context.Context, checkRunID uint) (*prmodels.CheckRun, error) {
	const op = "pipelinerun controller: get check run by id"
	defer wlog.Start(context.Background(), op).StopPrint()
	return c.prMgr.Check.GetCheckRunByID(ctx, checkRunID)
}

func (c *controller) CreateCheckRun(ctx context.Context, pipelineRunID uint,
	request *CreateOrUpdateCheckRunRequest) (*prmodels.CheckRun, error) {
	const op = "pipelinerun controller: create check run"
	defer wlog.Start(context.Background(), op).StopPrint()

	checkrun, err := c.prMgr.Check.CreateCheckRun(ctx, &prmodels.CheckRun{
		Name:          request.Name,
		CheckID:       request.CheckID,
		Status:        prmodels.String2CheckRunStatus(request.Status),
		Message:       request.Message,
		PipelineRunID: pipelineRunID,
		DetailURL:     request.DetailURL,
	})
	if err != nil {
		return nil, err
	}
	err = c.updatePrStatus(ctx, checkrun)
	if err != nil {
		return nil, err
	}
	return checkrun, nil
}

func (c *controller) CreatePRMessage(ctx context.Context, pipelineRunID uint,
	request *CreatePrMessageRequest) (*prmodels.PRMessage, error) {
	const op = "pipelinerun controller: create pr message"
	defer wlog.Start(context.Background(), op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return c.prMgr.Message.Create(ctx, &prmodels.PRMessage{
		PipelineRunID: pipelineRunID,
		Content:       request.Content,
		CreatedBy:     currentUser.GetID(),
		UpdatedBy:     currentUser.GetID(),
	})
}

func (c *controller) ListPRMessages(ctx context.Context,
	pipelineRunID uint, query *q.Query) (int, []*PrMessage, error) {
	const op = "pipelinerun controller: list pr messages"
	defer wlog.Start(context.Background(), op).StopPrint()

	count, messages, err := c.prMgr.Message.List(ctx, pipelineRunID, query)
	if err != nil {
		return 0, nil, err
	}
	userIDs := make([]uint, 0, len(messages))
	m := make(map[uint]struct{}, 0)
	for _, message := range messages {
		if _, ok := m[message.CreatedBy]; !ok {
			userIDs = append(userIDs, message.CreatedBy)
			m[message.CreatedBy] = struct{}{}
		}
		if _, ok := m[message.UpdatedBy]; !ok {
			userIDs = append(userIDs, message.UpdatedBy)
			m[message.UpdatedBy] = struct{}{}
		}
	}
	query = &q.Query{
		WithoutPagination: true,
		Keywords:          map[string]interface{}{common.UserQueryID: userIDs},
	}
	_, users, err := c.userMgr.List(ctx, query)
	if err != nil {
		return 0, nil, err
	}
	userMap := make(map[uint]*usermodels.User, 0)
	for _, user := range users {
		userMap[user.ID] = user
	}
	result := make([]*PrMessage, 0, len(messages))
	for _, message := range messages {
		resultMsg := &PrMessage{
			Content:   message.Content,
			CreatedAt: message.CreatedAt,
		}
		if user, ok := userMap[message.CreatedBy]; ok {
			resultMsg.CreatedBy = User{
				ID:   user.ID,
				Name: user.FullName,
			}
			if user.UserType == usermodels.UserTypeRobot {
				resultMsg.CreatedBy.UserType = "bot"
			}
		}
		if user, ok := userMap[message.UpdatedBy]; ok {
			resultMsg.UpdatedBy = User{
				ID:   user.ID,
				Name: user.FullName,
			}
			if user.UserType == usermodels.UserTypeRobot {
				resultMsg.CreatedBy.UserType = "bot"
			}
		}
		result = append(result, resultMsg)
	}
	return count, result, nil
}

func (c *controller) updatePrStatusByCheckrunID(ctx context.Context, checkrunID uint) error {
	Checkrun, err := c.prMgr.Check.GetCheckRunByID(ctx, checkrunID)
	if err != nil {
		return err
	}
	return c.updatePrStatus(ctx, Checkrun)
}

func (c *controller) updatePrStatus(ctx context.Context, checkrun *prmodels.CheckRun) error {
	pipelinerun, err := c.prMgr.PipelineRun.GetByID(ctx, checkrun.PipelineRunID)
	if err != nil {
		return err
	}
	if pipelinerun.Status != string(prmodels.StatusPending) {
		return nil
	}
	prStatus, err := func() (prmodels.PipelineStatus, error) {
		switch checkrun.Status {
		case prmodels.CheckStatusCancelled:
			return prmodels.StatusCancelled, nil
		case prmodels.CheckStatusFailure:
			return prmodels.StatusFailed, nil
		case prmodels.CheckStatusSuccess:
			return c.calculatePrSuccessStatus(ctx, pipelinerun)
		default:
			return prmodels.StatusPending, nil
		}
	}()
	if err != nil {
		return err
	}
	if prStatus == prmodels.StatusPending {
		return nil
	}
	return c.prMgr.PipelineRun.UpdateStatusByID(ctx, checkrun.PipelineRunID, prStatus)
}

func (c *controller) calculatePrSuccessStatus(ctx context.Context,
	pipelinerun *prmodels.Pipelinerun) (prmodels.PipelineStatus, error) {
	cluster, err := c.clusterMgr.GetByIDIncludeSoftDelete(ctx, pipelinerun.ClusterID)
	if err != nil {
		return prmodels.StatusPending, err
	}
	checks, err := c.prSvc.GetCheckByResource(ctx, cluster.ID, common.ResourceCluster)
	if err != nil {
		return prmodels.StatusPending, err
	}
	runs, err := c.prMgr.Check.ListCheckRuns(ctx, q.New(map[string]interface{}{
		common.CheckrunQueryByPipelinerunID: pipelinerun.ID},
	))
	if err != nil {
		return prmodels.StatusPending, err
	}
	checkSuccessMap := make(map[uint]bool, len(checks))
	for _, run := range runs {
		if run.Status != prmodels.CheckStatusSuccess {
			// if one checkrun is not success, return pending
			return prmodels.StatusPending, nil
		}
		checkSuccessMap[run.CheckID] = true
	}
	for _, check := range checks {
		if _, ok := checkSuccessMap[check.ID]; !ok {
			// if one check is not run, return pending
			return prmodels.StatusPending, nil
		}
	}
	return prmodels.StatusReady, nil
}
