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
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	cdmock "github.com/horizoncd/horizon/mock/pkg/cd"
	applicationservice "github.com/horizoncd/horizon/pkg/application/service"
	clusterservice "github.com/horizoncd/horizon/pkg/cluster/service"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	eventservice "github.com/horizoncd/horizon/pkg/event/service"
	groupservice "github.com/horizoncd/horizon/pkg/group/service"
	"github.com/stretchr/testify/assert"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	applicationmockmanager "github.com/horizoncd/horizon/mock/pkg/application/manager"
	commitmock "github.com/horizoncd/horizon/mock/pkg/cluster/code"
	clustergitrepomock "github.com/horizoncd/horizon/mock/pkg/cluster/gitrepo"
	clustermockmananger "github.com/horizoncd/horizon/mock/pkg/cluster/manager"
	tektonmock "github.com/horizoncd/horizon/mock/pkg/cluster/tekton"
	tektoncollectormock "github.com/horizoncd/horizon/mock/pkg/cluster/tekton/collector"
	tektonftymock "github.com/horizoncd/horizon/mock/pkg/cluster/tekton/factory"
	pipelinemockmanager "github.com/horizoncd/horizon/mock/pkg/pipelinerun/manager"
	usermock "github.com/horizoncd/horizon/mock/pkg/user/manager"
	applicationmodel "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	clustergitrepo "github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	clustermodel "github.com/horizoncd/horizon/pkg/cluster/models"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/collector"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/log"
	"github.com/horizoncd/horizon/pkg/config/token"
	envmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	"github.com/horizoncd/horizon/pkg/git"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	prmanager "github.com/horizoncd/horizon/pkg/pr/manager"
	"github.com/horizoncd/horizon/pkg/pr/models"
	prmodels "github.com/horizoncd/horizon/pkg/pr/models"
	prservice "github.com/horizoncd/horizon/pkg/pr/service"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	registrymodels "github.com/horizoncd/horizon/pkg/registry/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
	tokenservice "github.com/horizoncd/horizon/pkg/token/service"

	pipelinemodel "github.com/horizoncd/horizon/pkg/pr/models"
	usermodel "github.com/horizoncd/horizon/pkg/user/models"
)

var (
	ctx     context.Context
	manager *managerparam.Manager
)

func TestGetAndListPipelinerun(t *testing.T) {
	mockCtl := gomock.NewController(t)
	ctx := context.TODO()

	mockCommitGetter := commitmock.NewMockGitGetter(mockCtl)
	mockClusterManager := clustermockmananger.NewMockManager(mockCtl)
	mockApplicationMananger := applicationmockmanager.NewMockManager(mockCtl)
	mockPipelineManager := pipelinemockmanager.NewMockPipelineRunManager(mockCtl)
	mockClusterGitRepo := clustergitrepomock.NewMockClusterGitRepo(mockCtl)
	mockUserManager := usermock.NewMockManager(mockCtl)
	var ctl Controller = &controller{
		prMgr:      &prmanager.PRManager{PipelineRun: mockPipelineManager},
		clusterMgr: mockClusterManager,
		appMgr:     mockApplicationMananger,
		prSvc: prservice.NewService(
			&managerparam.Manager{
				ApplicationMgr: mockApplicationMananger,
				ClusterMgr:     mockClusterManager,
				UserMgr:        mockUserManager,
				PRMgr:          &prmanager.PRManager{PipelineRun: mockPipelineManager},
			}),
		envMgr:         nil,
		tektonFty:      nil,
		commitGetter:   mockCommitGetter,
		clusterGitRepo: mockClusterGitRepo,
		userMgr:        mockUserManager,
	}

	// 1. test Get PipelineBasic
	var pipelineID uint = 1932
	var createUser uint = 32
	mockPipelineManager.EXPECT().GetFirstCanRollbackPipelinerun(ctx, gomock.Any()).Return(nil, nil).AnyTimes()
	mockPipelineManager.EXPECT().GetByID(ctx, pipelineID).Return(&models.Pipelinerun{
		ID:        pipelineID,
		CreatedBy: createUser,
	}, nil).Times(1)
	var UserName = "tom"
	mockUserManager.EXPECT().GetUserByID(ctx, createUser).Return(&usermodel.User{
		Name: UserName,
	}, nil).Times(1)

	pipelineBasic, err := ctl.GetPipelinerun(ctx, pipelineID)
	assert.Nil(t, err)
	assert.Equal(t, pipelineBasic.ID, pipelineID)
	assert.Equal(t, pipelineBasic.CreatedBy, models.UserInfo{
		UserID:   createUser,
		UserName: UserName,
	})
	body, _ := json.MarshalIndent(pipelineBasic, "", " ")
	t.Logf("%s", string(body))

	// 2. test ListPipeline
	var clusterID uint = 56
	var totalCount = 100
	// var pipelineruns []*pipelinemodel.Pipelinerun
	pipelineruns := make([]*pipelinemodel.Pipelinerun, 0)
	pipelineruns = append(pipelineruns, &pipelinemodel.Pipelinerun{
		ID:        2,
		ClusterID: clusterID,
		CreatedBy: 1,
	})
	pipelineruns = append(pipelineruns, &pipelinemodel.Pipelinerun{
		ID:        3,
		ClusterID: clusterID,
		CreatedBy: 0,
	})

	mockPipelineManager.EXPECT().GetByClusterID(ctx,
		clusterID, gomock.Any(), gomock.Any()).Return(totalCount, pipelineruns, nil).Times(1)
	mockUserManager.EXPECT().GetUserByID(ctx, gomock.Any()).Return(&usermodel.User{
		Name: UserName,
	}, nil).AnyTimes()

	query := q.Query{
		PageNumber: 1,
		PageSize:   10,
	}
	count, pipelineBasics, err := ctl.ListPipelineruns(ctx, clusterID, false, query)
	assert.Nil(t, err)
	assert.Equal(t, count, totalCount)
	assert.Equal(t, len(pipelineBasics), 2)

	body, _ = json.MarshalIndent(pipelineBasics, "", " ")
	t.Logf("%s", string(body))
}

func TestGetDiff(t *testing.T) {
	mockCtl := gomock.NewController(t)
	ctx := context.TODO()

	mockCommitGetter := commitmock.NewMockGitGetter(mockCtl)
	mockClusterManager := clustermockmananger.NewMockManager(mockCtl)
	mockApplicationMananger := applicationmockmanager.NewMockManager(mockCtl)
	mockPipelineManager := pipelinemockmanager.NewMockPipelineRunManager(mockCtl)
	mockClusterGitRepo := clustergitrepomock.NewMockClusterGitRepo(mockCtl)

	var ctl Controller = &controller{
		prMgr:          &prmanager.PRManager{PipelineRun: mockPipelineManager},
		clusterMgr:     mockClusterManager,
		appMgr:         mockApplicationMananger,
		envMgr:         nil,
		tektonFty:      nil,
		commitGetter:   mockCommitGetter,
		clusterGitRepo: mockClusterGitRepo,
	}

	var pipelineID uint = 99854
	var clusterID uint = 3213
	gitURL := "ssh://git@cloudnative.com:22222/demo/springboot-demo.git"
	gitBranch := "master"
	gitCommit := "12388508504390230802832"
	configCommit := "123123"
	lastConfigCommit := "23232"
	mockCommitGetter.EXPECT().GetCommitHistoryLink(gomock.Any(), gomock.Any()).
		Return("https://cloudnative.com:22222/demo/springboot-demo/-/commits/"+gitCommit, nil).AnyTimes()
	mockPipelineManager.EXPECT().GetByID(ctx, pipelineID).Return(&models.Pipelinerun{
		ID:               0,
		ClusterID:        clusterID,
		GitURL:           gitURL,
		GitRefType:       codemodels.GitRefTypeBranch,
		GitRef:           gitBranch,
		GitCommit:        gitCommit,
		LastConfigCommit: lastConfigCommit,
		ConfigCommit:     configCommit,
	}, nil).Times(1)

	clusterName := "mycluster"
	var applicationID uint = 1234988
	mockClusterManager.EXPECT().GetByID(ctx, clusterID).Return(&clustermodel.Cluster{
		ApplicationID: uint(applicationID),
		Name:          clusterName,
	}, nil).Times(1)

	applicationName := "myapplication"
	mockApplicationMananger.EXPECT().GetByID(ctx, applicationID).Return(&applicationmodel.Application{
		Name: applicationName,
	}, nil).Times(1)

	commitMsg := "hello world"
	mockCommitGetter.EXPECT().GetCommit(ctx, gitURL, codemodels.GitRefTypeCommit, gitCommit).
		Return(&git.Commit{
			ID:      gitCommit,
			Message: commitMsg,
		}, nil)

	diff := "this is mydiff"
	mockClusterGitRepo.EXPECT().CompareConfig(ctx, applicationName, clusterName,
		&lastConfigCommit, &configCommit).Return(diff, nil).Times(1)

	resp, err := ctl.GetDiff(ctx, pipelineID)
	assert.Nil(t, err)
	body, _ := json.Marshal(resp)
	t.Logf("response = %s", string(body))

	link, _ := mockCommitGetter.GetCommitHistoryLink(gitURL, gitCommit)
	expectResp := &GetDiffResponse{
		CodeInfo: &CodeInfo{
			Branch:    gitBranch,
			CommitID:  gitCommit,
			CommitMsg: commitMsg,
			Link:      link,
		},
		ConfigDiff: &ConfigDiff{
			From: lastConfigCommit,
			To:   configCommit,
			Diff: diff,
		},
	}
	assert.Equal(t, *expectResp, *resp)
}

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)

	if err := db.AutoMigrate(&clustermodel.Cluster{}, &membermodels.Member{},
		&envmodels.EnvironmentRegion{}, &prmodels.Pipelinerun{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&groupmodels.Group{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   uint(1),
	})

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tektonFty := tektonftymock.NewMockFactory(mockCtl)
	tekton := tektonmock.NewMockInterface(mockCtl)
	tektonCollector := tektoncollectormock.NewMockInterface(mockCtl)
	tektonFty.EXPECT().GetTekton(gomock.Any()).Return(tekton, nil).AnyTimes()
	tektonFty.EXPECT().GetTektonCollector(gomock.Any()).Return(tektonCollector, nil).AnyTimes()

	envMgr := manager.EnvMgr

	clusterMgr := manager.ClusterMgr
	cluster, err := clusterMgr.Create(ctx, &clustermodel.Cluster{
		Name:            "cluster",
		EnvironmentName: "test",
		RegionName:      "hz",
		CreatedBy:       0,
		UpdatedBy:       0,
	}, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, cluster)

	pipelinerunMgr := manager.PRMgr.PipelineRun
	pipelinerun, err := pipelinerunMgr.Create(ctx, &prmodels.Pipelinerun{
		ClusterID: cluster.ID,
		Action:    "builddeploy",
		Status:    "ok",
		S3Bucket:  "bucket",
		LogObject: "logObject",
		PrObject:  "prObject",
		CreatedBy: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, pipelinerun)

	c := &controller{
		prMgr:      manager.PRMgr,
		clusterMgr: clusterMgr,
		envMgr:     envMgr,
		tektonFty:  tektonFty,
	}

	logBytes := []byte("this is a log")
	tektonCollector.EXPECT().GetPipelineRunLog(ctx, gomock.Any()).
		Return(&collector.Log{LogBytes: logBytes}, nil).Times(1)

	l, err := c.GetPipelinerunLog(ctx, pipelinerun.ID)
	assert.Nil(t, err)
	assert.Nil(t, l.LogChannel)
	assert.Nil(t, l.ErrChannel)
	assert.Equal(t, l.LogBytes, logBytes)
	t.Logf("logBytes: %v", string(l.LogBytes))

	pipelinerun, err = pipelinerunMgr.Create(ctx, &prmodels.Pipelinerun{
		ClusterID: cluster.ID,
		Action:    "builddeploy",
		Status:    "ok",
		S3Bucket:  "",
		LogObject: "",
		PrObject:  "",
		CreatedBy: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, pipelinerun)

	logCh := make(chan log.Log)
	errCh := make(chan error)
	tektonCollector.EXPECT().GetPipelineRunLog(ctx, gomock.Any()).
		Return(&collector.Log{
			LogChannel: logCh,
			ErrChannel: errCh,
		}, nil).Times(1)

	go func() {
		defer close(logCh)
		defer close(errCh)
		for i := 0; i < 10; i++ {
			logCh <- log.Log{
				Pipeline: "default",
				Task:     "task",
				Step:     "step",
				Log:      strconv.Itoa(i),
			}
		}
	}()

	l, err = c.GetClusterLatestLog(ctx, cluster.ID)
	assert.Nil(t, err)
	assert.Nil(t, l.LogBytes)
	logC := l.LogChannel
	errC := l.ErrChannel
	for logC != nil || errC != nil {
		select {
		case l, ok := <-logC:
			if !ok {
				logC = nil
				continue
			}
			if l.Log == "EOFLOG" {
				t.Logf("\n")
				continue
			}
			t.Logf("[%s : %s] %s\n", l.Task, l.Step, l.Log)
		case e, ok := <-errC:
			if !ok {
				errC = nil
				continue
			}
			t.Logf("%s\n", e)
		}
	}

	// test for stop
	err = c.StopPipelinerun(ctx, pipelinerun.ID)
	assert.NotNil(t, err)
	t.Logf("err: %v", err)

	pipelinerun, err = pipelinerunMgr.Create(ctx, &prmodels.Pipelinerun{
		ClusterID: cluster.ID,
		Action:    "builddeploy",
		Status:    string(prmodels.StatusRunning),
		S3Bucket:  "",
		LogObject: "",
		PrObject:  "",
		CreatedBy: 1,
	})
	assert.Nil(t, err)
	tekton.EXPECT().StopPipelineRun(ctx, gomock.Any()).Return(nil).AnyTimes()
	err = c.StopPipelinerun(ctx, pipelinerun.ID)
	assert.Nil(t, err)

	err = c.StopPipelinerunForCluster(ctx, pipelinerun.ClusterID)
	assert.Nil(t, err)
}

func TestExecutePipelineRun(t *testing.T) {
	db, _ := orm.NewSqliteDB("")
	if err := db.AutoMigrate(&applicationmodel.Application{}, &clustermodel.Cluster{},
		&regionmodels.Region{}, &membermodels.Member{}, &registrymodels.Registry{},
		&prmodels.Pipelinerun{}, &groupmodels.Group{}, &prmodels.Check{},
		&usermodel.User{}, &trmodels.TemplateRelease{}, &eventmodels.Event{}); err != nil {
		panic(err)
	}
	mgr := managerparam.InitManager(db)
	ctx := context.Background()
	// nolint
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   uint(1),
	})
	mockCtl := gomock.NewController(t)

	mockTektonInterface := tektonmock.NewMockInterface(mockCtl)

	mockFactory := tektonftymock.NewMockFactory(mockCtl)
	mockFactory.EXPECT().GetTekton(gomock.Any()).Return(mockTektonInterface, nil).AnyTimes()
	tokenConfig := token.Config{
		JwtSigningKey:         "hello",
		CallbackTokenExpireIn: 24 * time.Hour,
	}
	mockClusterGitRepo := clustergitrepomock.NewMockClusterGitRepo(mockCtl)
	mockCD := cdmock.NewMockCD(mockCtl)

	groupSvc := groupservice.NewService(mgr)
	eventSvc := eventservice.New(mgr)
	applicationSvc := applicationservice.NewService(groupSvc, mgr)
	clusterSvc := clusterservice.NewService(applicationSvc, mockClusterGitRepo, mgr)

	ctrl := controller{
		prMgr:              mgr.PRMgr,
		appMgr:             mgr.ApplicationMgr,
		clusterMgr:         mgr.ClusterMgr,
		envMgr:             mgr.EnvMgr,
		regionMgr:          mgr.RegionMgr,
		tektonFty:          mockFactory,
		tokenSvc:           tokenservice.NewService(mgr, tokenConfig),
		tokenConfig:        tokenConfig,
		clusterGitRepo:     mockClusterGitRepo,
		templateReleaseMgr: mgr.TemplateReleaseMgr,
		cd:                 mockCD,
		clusterSvc:         clusterSvc,
		eventSvc:           eventSvc,
	}

	_, err1 := mgr.EventMgr.CreateEvent(ctx, &eventmodels.Event{
		EventSummary: eventmodels.EventSummary{
			ResourceID: 1,
		},
	})
	assert.NoError(t, err1)

	_, err := mgr.UserMgr.Create(ctx, &usermodel.User{
		Name: "Tony",
	})
	assert.NoError(t, err)

	group, err := mgr.GroupMgr.Create(ctx, &groupmodels.Group{
		Name: "test",
	})
	assert.NoError(t, err)

	app, err := mgr.ApplicationMgr.Create(ctx, &applicationmodel.Application{
		Name:    "test",
		GroupID: group.ID,
	}, nil)
	assert.NoError(t, err)

	registryID, err := mgr.RegistryMgr.Create(ctx, &registrymodels.Registry{
		Name: "test",
	})
	assert.NoError(t, err)

	region, err := mgr.RegionMgr.Create(ctx, &regionmodels.Region{
		Name:       "test",
		RegistryID: registryID,
	})
	assert.NoError(t, err)

	cluster, err := mgr.ClusterMgr.Create(ctx, &clustermodel.Cluster{
		Name:            "clusterGit",
		ApplicationID:   app.ID,
		GitURL:          "hello",
		RegionName:      region.Name,
		Template:        "javaapp",
		TemplateRelease: "v1.0.0",
	}, nil, nil)
	assert.NoError(t, err)

	_, err = mgr.TemplateReleaseMgr.Create(ctx, &trmodels.TemplateRelease{
		TemplateName: "javaapp",
		Name:         "v1.0.0",
	})
	assert.NoError(t, err)

	mockClusterGitRepo.EXPECT().GetCluster(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&clustergitrepo.ClusterFiles{
			PipelineJSONBlob:    map[string]interface{}{},
			ApplicationJSONBlob: map[string]interface{}{},
		}, nil).AnyTimes()
	mockClusterGitRepo.EXPECT().GetConfigCommit(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&clustergitrepo.ClusterCommit{
			Master: "master",
			Gitops: "gitops",
		}, nil).AnyTimes()
	mockClusterGitRepo.EXPECT().UpdateRestartTime(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("restart_commit", nil).AnyTimes()
	mockClusterGitRepo.EXPECT().CheckAndSyncGitOpsBranch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	mockClusterGitRepo.EXPECT().Rollback(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("rollback_commit", nil).AnyTimes()
	mockClusterGitRepo.EXPECT().MergeBranch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).Return("rollback_master_commit", nil).AnyTimes()
	mockClusterGitRepo.EXPECT().GetClusterTemplate(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&clustergitrepo.ClusterTemplate{
			Name:    "javaapp",
			Release: "v1.0.0",
		}, nil).AnyTimes()
	mockClusterGitRepo.EXPECT().GetClusterValueFiles(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]clustergitrepo.ClusterValueFile{}, nil).AnyTimes()
	mockClusterGitRepo.EXPECT().GetEnvValue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&clustergitrepo.EnvValue{
			Environment:   "test",
			Region:        region.Name,
			Namespace:     "default",
			BaseRegistry:  "registry.cn-hangzhou.aliyuncs.com",
			IngressDomain: region.IngressDomain,
		}, nil).AnyTimes()
	mockClusterGitRepo.EXPECT().GetRepoInfo(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&clustergitrepo.RepoInfo{}).AnyTimes()
	mockClusterGitRepo.EXPECT().DefaultBranch().Return("master").AnyTimes()

	mockTektonInterface.EXPECT().CreatePipelineRun(ctx, gomock.Any()).Return("hello", nil).AnyTimes()

	mockCD.EXPECT().CreateCluster(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockCD.EXPECT().DeployCluster(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	prPending, err := mgr.PRMgr.PipelineRun.Create(ctx, &prmodels.Pipelinerun{
		ClusterID: cluster.ID,
		Action:    prmodels.ActionBuildDeploy,
		Status:    string(pipelinemodel.StatusPending),
	})
	assert.NoError(t, err)
	assert.Equal(t, string(pipelinemodel.StatusPending), prPending.Status)

	prPendingToForceReady, err := mgr.PRMgr.PipelineRun.Create(ctx, &prmodels.Pipelinerun{
		ClusterID: cluster.ID,
		Action:    prmodels.ActionBuildDeploy,
		Status:    string(pipelinemodel.StatusPending),
	})
	assert.NoError(t, err)
	assert.Equal(t, string(pipelinemodel.StatusPending), prPendingToForceReady.Status)

	prDeployReady, err := mgr.PRMgr.PipelineRun.Create(ctx, &prmodels.Pipelinerun{
		ClusterID: cluster.ID,
		Action:    prmodels.ActionDeploy,
		Status:    string(pipelinemodel.StatusReady),
	})
	assert.NoError(t, err)
	assert.Equal(t, string(pipelinemodel.StatusReady), prDeployReady.Status)

	prRestartReady, err := mgr.PRMgr.PipelineRun.Create(ctx, &prmodels.Pipelinerun{
		ClusterID: cluster.ID,
		Action:    prmodels.ActionRestart,
		Status:    string(pipelinemodel.StatusReady),
	})
	assert.NoError(t, err)
	assert.Equal(t, string(pipelinemodel.StatusReady), prRestartReady.Status)

	prOK, err := mgr.PRMgr.PipelineRun.Create(ctx, &prmodels.Pipelinerun{
		ClusterID:        cluster.ID,
		Action:           prmodels.ActionBuildDeploy,
		Status:           string(pipelinemodel.StatusOK),
		ConfigCommit:     "ok_commit",
		LastConfigCommit: "last_ok_commit",
	})
	assert.NoError(t, err)
	assert.Equal(t, string(pipelinemodel.StatusOK), prOK.Status)

	prRollbackReady, err := mgr.PRMgr.PipelineRun.Create(ctx, &prmodels.Pipelinerun{
		ClusterID:    cluster.ID,
		Action:       prmodels.ActionRollback,
		Status:       string(pipelinemodel.StatusReady),
		RollbackFrom: &prOK.ID,
	})
	assert.NoError(t, err)
	assert.Equal(t, string(pipelinemodel.StatusReady), prRollbackReady.Status)

	err = ctrl.Execute(ctx, prDeployReady.ID)
	assert.NoError(t, err)

	err = ctrl.Execute(ctx, prRestartReady.ID)
	assert.NoError(t, err)

	err = ctrl.Execute(ctx, prRollbackReady.ID)
	assert.NoError(t, err)

	err = ctrl.Execute(ctx, prPending.ID)
	assert.NotNil(t, err)

	err = ctrl.Ready(ctx, prPendingToForceReady.ID)
	assert.NoError(t, err)
	err = ctrl.Execute(ctx, prPendingToForceReady.ID)
	assert.Nil(t, err)

	PRCancel, err := mgr.PRMgr.PipelineRun.Create(ctx, &prmodels.Pipelinerun{
		ClusterID: cluster.ID,
		Status:    string(pipelinemodel.StatusPending),
	})
	assert.NoError(t, err)

	err = ctrl.Cancel(ctx, PRCancel.ID)
	assert.NoError(t, err)

	err = ctrl.Cancel(ctx, prDeployReady.ID)
	assert.NotNil(t, err)
}

func TestCheckRun(t *testing.T) {
	db, _ := orm.NewSqliteDB("")
	if err := db.AutoMigrate(&groupmodels.Group{}, &membermodels.Member{}, &applicationmodel.Application{},
		&clustermodel.Cluster{}, &prmodels.CheckRun{}, &prmodels.Pipelinerun{},
		&usermodel.User{}, &prmodels.Check{}); err != nil {
		panic(err)
	}
	param := managerparam.InitManager(db)
	ctx := context.Background()
	// nolint
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   uint(1),
	})

	ctrl := controller{
		clusterMgr: param.ClusterMgr,
		prMgr:      param.PRMgr,
		prSvc:      prservice.NewService(param),
	}

	_, err := param.UserMgr.Create(ctx, &usermodel.User{
		Name: "Tony",
	})
	assert.NoError(t, err)

	group, err := param.GroupMgr.Create(ctx, &groupmodels.Group{
		Name: "test",
	})
	assert.NoError(t, err)

	app, err := param.ApplicationMgr.Create(ctx, &applicationmodel.Application{
		Name:    "test",
		GroupID: group.ID,
	}, nil)
	assert.NoError(t, err)

	cluster, err := param.ClusterMgr.Create(ctx, &clustermodel.Cluster{
		Name:          "cluster",
		ApplicationID: app.ID,
	}, nil, nil)
	assert.NoError(t, err)

	// create a check under cluster
	check, err := ctrl.CreateCheck(ctx, &pipelinemodel.Check{
		Resource: common.Resource{
			ResourceID: cluster.ID,
			Type:       "clusters",
		},
	})
	assert.NoError(t, err)

	pr, err := ctrl.prMgr.PipelineRun.Create(ctx, &prmodels.Pipelinerun{
		ID:        1,
		Status:    string(prmodels.StatusPending),
		ClusterID: cluster.ID,
	})
	assert.NoError(t, err)

	_, err = ctrl.CreateCheckRun(ctx, pr.ID, &CreateOrUpdateCheckRunRequest{
		Name:      "test",
		Status:    string(prmodels.CheckStatusPending),
		CheckID:   check.ID,
		Message:   "hello",
		DetailURL: "https://www.google.com",
	})
	assert.NoError(t, err)

	err = ctrl.UpdateCheckRunByID(ctx, 1, &CreateOrUpdateCheckRunRequest{
		Status: string(prmodels.CheckStatusSuccess),
	})
	assert.NoError(t, err)

	cr, err := ctrl.GetCheckRunByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, string(cr.Status), string(prmodels.CheckStatusSuccess))

	prInDB, err := ctrl.GetPipelinerun(ctx, pr.ID)
	assert.NoError(t, err)
	assert.Equal(t, string(prInDB.Status), string(prmodels.StatusReady))

	checkRuns, err := ctrl.ListCheckRuns(ctx, pr.ID)
	assert.NoError(t, err)
	assert.Equal(t, len(checkRuns), 1)
}

func TestMessage(t *testing.T) {
	db, _ := orm.NewSqliteDB("")
	if err := db.AutoMigrate(&prmodels.PRMessage{}, &usermodel.User{}, &prmodels.Pipelinerun{}); err != nil {
		panic(err)
	}
	param := managerparam.InitManager(db)
	ctx := context.Background()
	// nolint
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   uint(1),
	})

	_, err := param.UserMgr.Create(ctx, &usermodel.User{
		Name: "Tony",
	})
	assert.NoError(t, err)

	pr, err := param.PRMgr.PipelineRun.Create(ctx, &prmodels.Pipelinerun{
		Action: prmodels.ActionBuildDeploy,
	})
	assert.NoError(t, err)

	ctrl := controller{
		prMgr:   param.PRMgr,
		userMgr: param.UserMgr,
		prSvc: prservice.NewService(&managerparam.Manager{
			PRMgr: param.PRMgr,
		}),
	}

	message1, err := ctrl.CreatePRMessage(ctx, pr.ID, &CreatePRMessageRequest{
		Content: "first",
	})
	assert.NoError(t, err)
	assert.Equal(t, message1.Content, "first")

	time.Sleep(time.Second)

	message2, err := ctrl.CreatePRMessage(ctx, pr.ID, &CreatePRMessageRequest{
		Content: "second",
	})
	assert.NoError(t, err)
	assert.Equal(t, message2.Content, "second")

	count, messages, err := ctrl.ListPRMessages(ctx, pr.ID, &q.Query{})
	assert.NoError(t, err)
	assert.Equal(t, count, 2)
	assert.Equal(t, len(messages), 2)
	assert.Equal(t, messages[0].Content, "first")
	assert.Equal(t, messages[1].Content, "second")
}
