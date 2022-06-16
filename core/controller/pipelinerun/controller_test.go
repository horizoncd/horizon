package pipelinerun

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	applicationmockmanager "g.hz.netease.com/horizon/mock/pkg/application/manager"
	commitmock "g.hz.netease.com/horizon/mock/pkg/cluster/code"
	clustergitrepomock "g.hz.netease.com/horizon/mock/pkg/cluster/gitrepo"
	clustermockmananger "g.hz.netease.com/horizon/mock/pkg/cluster/manager"
	tektonmock "g.hz.netease.com/horizon/mock/pkg/cluster/tekton"
	tektoncollectormock "g.hz.netease.com/horizon/mock/pkg/cluster/tekton/collector"
	tektonftymock "g.hz.netease.com/horizon/mock/pkg/cluster/tekton/factory"
	pipelinemockmanager "g.hz.netease.com/horizon/mock/pkg/pipelinerun/manager"
	usermock "g.hz.netease.com/horizon/mock/pkg/user/manager"
	applicationmodel "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/cluster/code"
	clustermodel "g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/log"
	envmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	"g.hz.netease.com/horizon/pkg/pipelinerun/models"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	pipelinemodel "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	usermodel "g.hz.netease.com/horizon/pkg/user/models"
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
	mockPipelineManager := pipelinemockmanager.NewMockManager(mockCtl)
	mockClusterGitRepo := clustergitrepomock.NewMockClusterGitRepo(mockCtl)
	mockUserManager := usermock.NewMockManager(mockCtl)
	var ctl Controller = &controller{
		pipelinerunMgr: mockPipelineManager,
		clusterMgr:     mockClusterManager,
		applicationMgr: mockApplicationMananger,
		envMgr:         nil,
		tektonFty:      nil,
		commitGetter:   mockCommitGetter,
		clusterGitRepo: mockClusterGitRepo,
		userManager:    mockUserManager,
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

	pipelineBasic, err := ctl.Get(ctx, pipelineID)
	assert.Nil(t, err)
	assert.Equal(t, pipelineBasic.ID, pipelineID)
	assert.Equal(t, pipelineBasic.CreatedBy, UserInfo{
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
	count, pipelineBasics, err := ctl.List(ctx, clusterID, false, query)
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
	mockPipelineManager := pipelinemockmanager.NewMockManager(mockCtl)
	mockClusterGitRepo := clustergitrepomock.NewMockClusterGitRepo(mockCtl)

	var ctl Controller = &controller{
		pipelinerunMgr: mockPipelineManager,
		clusterMgr:     mockClusterManager,
		applicationMgr: mockApplicationMananger,
		envMgr:         nil,
		tektonFty:      nil,
		commitGetter:   mockCommitGetter,
		clusterGitRepo: mockClusterGitRepo,
	}

	var pipelineID uint = 99854
	var clusterID uint = 3213
	gitURL := "git@github.com:demo/demo.git"
	gitBranch := "master"
	gitCommit := "12388508504390230802832"
	configCommit := "123123"
	lastConfigCommit := "23232"
	mockPipelineManager.EXPECT().GetByID(ctx, pipelineID).Return(&models.Pipelinerun{
		ID:               0,
		ClusterID:        clusterID,
		GitURL:           gitURL,
		GitBranch:        gitBranch,
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
	mockCommitGetter.EXPECT().GetCommit(ctx, gitURL, nil, &gitCommit).Return(&code.Commit{
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

	expectResp := &GetDiffResponse{
		CodeInfo: &CodeInfo{
			Branch:    gitBranch,
			CommitID:  gitCommit,
			CommitMsg: commitMsg,
			Link:      common.InternalSSHToHTTPURL(gitURL) + common.CommitHistoryMiddle + gitCommit,
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

	pipelinerunMgr := manager.PipelinerunMgr
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
		pipelinerunMgr: pipelinerunMgr,
		clusterMgr:     clusterMgr,
		envMgr:         envMgr,
		tektonFty:      tektonFty,
	}

	logBytes := []byte("this is a log")
	tektonCollector.EXPECT().GetPipelineRunLog(ctx, gomock.Any()).Return(logBytes, nil).AnyTimes()

	logCh := make(chan log.Log)
	errCh := make(chan error)
	tekton.EXPECT().GetPipelineRunLogByID(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(logCh, errCh, nil)

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
		Status:    prmodels.ResultCreated,
		S3Bucket:  "",
		LogObject: "",
		PrObject:  "",
		CreatedBy: 1,
	})
	assert.Nil(t, err)
	tekton.EXPECT().StopPipelineRun(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	err = c.StopPipelinerun(ctx, pipelinerun.ID)
	assert.Nil(t, err)

	err = c.StopPipelinerunForCluster(ctx, pipelinerun.ClusterID)
	assert.Nil(t, err)
}
