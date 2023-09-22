package cluster

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	mock_code "github.com/horizoncd/horizon/mock/pkg/cluster/code"
	mock_gitrepo "github.com/horizoncd/horizon/mock/pkg/cluster/gitrepo"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	clustergitrepo "github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	"github.com/horizoncd/horizon/pkg/git"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	prmodels "github.com/horizoncd/horizon/pkg/pr/models"
	prservice "github.com/horizoncd/horizon/pkg/pr/service"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	registrymodels "github.com/horizoncd/horizon/pkg/registry/models"
	usermodel "github.com/horizoncd/horizon/pkg/user/models"
)

func TestCreatePipelineRun(t *testing.T) {
	db, _ := orm.NewSqliteDB("")
	if err := db.AutoMigrate(&appmodels.Application{}, &models.Cluster{},
		&regionmodels.Region{}, &membermodels.Member{}, &registrymodels.Registry{},
		&prmodels.Pipelinerun{}, &groupmodels.Group{}, &prmodels.Check{},
		&usermodel.User{}, &eventmodels.Event{}); err != nil {
		panic(err)
	}
	param := managerparam.InitManager(db)
	ctx := context.Background()
	// nolint
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   uint(1),
	})
	mockCtl := gomock.NewController(t)
	mockGitGetter := mock_code.NewMockGitGetter(mockCtl)
	mockGitGetter.EXPECT().GetCommit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&git.Commit{Message: "test"}, nil).AnyTimes()

	mockClusterGitRepo := mock_gitrepo.NewMockClusterGitRepo(mockCtl)
	mockClusterGitRepo.EXPECT().GetConfigCommit(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&clustergitrepo.ClusterCommit{
			Master: "master",
			Gitops: "gitops",
		}, nil).AnyTimes()

	controller := &controller{
		prSvc:          prservice.NewService(param),
		prMgr:          param.PRMgr,
		clusterMgr:     param.ClusterMgr,
		applicationMgr: param.ApplicationMgr,
		regionMgr:      param.RegionMgr,
		clusterGitRepo: mockClusterGitRepo,
		commitGetter:   mockGitGetter,
		eventMgr:       param.EventMgr,
	}

	_, err := param.UserMgr.Create(ctx, &usermodel.User{
		Name: "Tony",
	})
	assert.NoError(t, err)

	group, err := param.GroupMgr.Create(ctx, &groupmodels.Group{
		Name: "test",
	})
	assert.NoError(t, err)

	app, err := param.ApplicationMgr.Create(ctx, &appmodels.Application{
		Name:    "test",
		GroupID: group.ID,
	}, nil)
	assert.NoError(t, err)

	registryID, err := param.RegistryMgr.Create(ctx, &registrymodels.Registry{
		Name: "test",
	})
	assert.NoError(t, err)

	region, err := param.RegionMgr.Create(ctx, &regionmodels.Region{
		Name:       "test",
		RegistryID: registryID,
	})

	assert.NoError(t, err)
	clusterGit, err := param.ClusterMgr.Create(ctx, &models.Cluster{
		Name:          "clusterGit",
		ApplicationID: app.ID,
		GitURL:        "hello",
		RegionName:    region.Name,
	}, nil, nil)
	assert.NoError(t, err)

	// request for build deploy
	requestBuildDeploy := &CreatePipelineRunRequest{
		Action:      prmodels.ActionBuildDeploy,
		Title:       "test",
		Description: "test",
		Git: &BuildDeployRequestGit{
			Commit: "test",
		},
	}
	pipelineBuildDeploy, err := controller.CreatePipelineRun(ctx, clusterGit.ID, requestBuildDeploy)
	assert.NoError(t, err)
	assert.Equal(t, "ready", pipelineBuildDeploy.Status)

	mockClusterGitRepo.EXPECT().GetCluster(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&clustergitrepo.ClusterFiles{
			PipelineJSONBlob:    map[string]interface{}{},
			ApplicationJSONBlob: map[string]interface{}{},
		}, nil)
	mockClusterGitRepo.EXPECT().CompareConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("diff", nil)
	requestDeploy := &CreatePipelineRunRequest{
		Action:      prmodels.ActionDeploy,
		Title:       "test",
		Description: "test",
	}
	pipelineDeploy, err := controller.CreatePipelineRun(ctx, clusterGit.ID, requestDeploy)
	assert.NoError(t, err)
	assert.Equal(t, "ready", pipelineDeploy.Status)

	requestRollback := &CreatePipelineRunRequest{
		Action:        prmodels.ActionRollback,
		Title:         "test",
		Description:   "test",
		PipelinerunID: pipelineBuildDeploy.ID,
	}
	_, err = controller.CreatePipelineRun(ctx, clusterGit.ID, requestRollback)
	assert.NotNil(t, err)

	_, err = param.PRMgr.Check.Create(ctx, &prmodels.Check{
		Resource: common.Resource{
			ResourceID: group.ID,
			Type:       common.ResourceGroup,
		},
	})
	assert.NoError(t, err)

	pipelineBuildDeployPending, err := controller.CreatePipelineRun(ctx, clusterGit.ID, requestBuildDeploy)
	assert.NoError(t, err)
	assert.Equal(t, "pending", pipelineBuildDeployPending.Status)
}
