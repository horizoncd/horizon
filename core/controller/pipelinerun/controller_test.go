package pipelinerun

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	applicationmockmanager "g.hz.netease.com/horizon/mock/pkg/application/manager"
	commitmock "g.hz.netease.com/horizon/mock/pkg/cluster/code"
	clustergitrepomock "g.hz.netease.com/horizon/mock/pkg/cluster/gitrepo"
	clustermockmananger "g.hz.netease.com/horizon/mock/pkg/cluster/manager"
	pipelinemockmanager "g.hz.netease.com/horizon/mock/pkg/pipelinerun/manager"
	applicationmodel "g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/cluster/code"
	clustermodel "g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetDiff(t *testing.T) {
	mockCtl := gomock.NewController(t)
	ctx := context.TODO()

	mockCommitGetter := commitmock.NewMockCommitGetter(mockCtl)
	mockClusterManager := clustermockmananger.NewMockManager(mockCtl)
	mockApplicationMananger := applicationmockmanager.NewMockManager(mockCtl)
	mockPipelineManager := pipelinemockmanager.NewMockManager(mockCtl)
	mockClusterGitRepo := clustergitrepomock.NewMockClusterGitRepo(mockCtl)

	var ctl Controller = &controller{
		prMgr:          mockPipelineManager,
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
	mockCommitGetter.EXPECT().GetCommit(ctx, gitURL, gitCommit).Return(&code.Commit{
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
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
