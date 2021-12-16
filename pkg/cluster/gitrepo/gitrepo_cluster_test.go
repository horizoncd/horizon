package gitrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/middleware/user"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	gitlablibmock "g.hz.netease.com/horizon/mock/lib/gitlab"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	clustertagmodels "g.hz.netease.com/horizon/pkg/clustertag/models"
	"g.hz.netease.com/horizon/pkg/config/gitlab"
	gitlabconf "g.hz.netease.com/horizon/pkg/config/gitlab"
	harbormodels "g.hz.netease.com/horizon/pkg/harbor/models"
	k8sclustermodels "g.hz.netease.com/horizon/pkg/k8scluster/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	herrors "g.hz.netease.com/horizon/pkg/util/errors"
	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
)

/*
go test -v ./pkg/cluster/gitrepo

NOTE: when there is no GITLAB_PARAMS_FOR_TEST environment variable, skip this test.
*/

// nolint
var (
	ctx context.Context
	g   gitlablib.Interface

	sshURL        string
	rootGroupName string
	templateName  string

	pipelineJSONBlob, applicationJSONBlob map[string]interface{}
	pipelineJSONStr                       = `{
            "buildxml":"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE project [<!ENTITY buildfile SYSTEM \"file:./build-user.xml\">]>\n<project basedir=\".\" default=\"deploy\" name=\"demo\">\n    <property name=\"ant\" value=\"ant\" />\n    <property name=\"baseline.dir\" value=\"${basedir}\"/>\n\n    <target name=\"package\">\n        <exec dir=\"${baseline.dir}\" executable=\"${ant}\" failonerror=\"true\">\n            <arg line=\"-buildfile overmind_build.xml -Denv=test -DenvName=mockserver.org\"/>\n        </exec>\n    </target>\n\n    <target name=\"deploy\">\n        <echo message=\"begin auto deploy......\"/>\n        <antcall target=\"package\"/>\n    </target>\n</project>"
        }`

	applicationJSONStr = `{
    "app":{
        "health":{
            "check":"/api/test",
            "offline":"/health/offline",
            "online":"/health/online",
            "port":1024,
            "status":"/health/check"
        },
        "params":{
            "jvmExtra":"sdfa",
            "mainClassName":"fsda",
            "xms":"512",
            "xmx":"512"
        },
        "spec":{
            "replicas":1,
            "resource":"small"
        },
        "strategy":{
            "pauseType":"first",
            "stepsTotal":2
        }
    }
}`
)

type Param struct {
	Token         string `json:"token"`
	BaseURL       string `json:"baseURL"`
	RootGroupName string `json:"rootGroupName"`
	RootGroupID   int    `json:"rootGroupId"`
}

// nolint
func TestMain(m *testing.M) {
	var err error
	param := os.Getenv("GITLAB_PARAMS_FOR_TEST")

	if param == "" {
		return
	}

	var p *Param
	if err := json.Unmarshal([]byte(param), &p); err != nil {
		panic(err)
	}

	sshURL = "ssh://gitlab.com"

	g, err = gitlablib.New(p.Token, p.BaseURL, sshURL)
	if err != nil {
		panic(err)
	}

	ctx = context.WithValue(context.Background(), user.Key(), &userauth.DefaultInfo{
		Name: "Tony",
	})

	rootGroupName = p.RootGroupName
	templateName = "javaapp"

	if err := json.Unmarshal([]byte(pipelineJSONStr), &pipelineJSONBlob); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(applicationJSONStr), &applicationJSONBlob); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	r := &clusterGitRepo{
		gitlabLib: g,
		clusterRepoConf: &gitlab.Repo{
			Parent: &gitlab.Parent{
				Path: fmt.Sprintf("%v/%v", rootGroupName, "clusters"),
				ID:   4970,
			},
			RecyclingParent: &gitlab.Parent{
				Path: fmt.Sprintf("%v/%v", rootGroupName, "recycling-clusters"),
				ID:   4971,
			},
		},
		helmRepoMapper: map[string]string{
			"test": "https://test.helm.com",
		},
	}
	application := "app"
	cluster := "cluster"
	baseParams := &BaseParams{
		Cluster:             cluster,
		PipelineJSONBlob:    pipelineJSONBlob,
		ApplicationJSONBlob: applicationJSONBlob,
		TemplateRelease: &trmodels.TemplateRelease{
			TemplateName: templateName,
			Name:         "v1.0.0",
		},
		Application: &appmodels.Application{
			GroupID:  10,
			Name:     application,
			Priority: "P0",
		},
		Environment: "test",
	}
	createParams := &CreateClusterParams{
		BaseParams: baseParams,
		RegionEntity: &regionmodels.RegionEntity{
			Region: &regionmodels.Region{
				Name:        "hz",
				DisplayName: "HZ",
			},
			K8SCluster: &k8sclustermodels.K8SCluster{
				Server: "https://k8s.com",
			},
			Harbor: &harbormodels.Harbor{
				Server: "https://harbor.com",
			},
		},
	}

	updateParams := &UpdateClusterParams{
		BaseParams: baseParams,
	}

	defer func() {
		_ = r.DeleteCluster(ctx, application, cluster, 1)
		_ = g.DeleteProject(ctx, fmt.Sprintf("%v/%v/%v/%v-%v", rootGroupName,
			"recycling-clusters", application, cluster, 1))
	}()
	err := r.CreateCluster(ctx, createParams)
	assert.Nil(t, err)

	updateParams.Application.Priority = "P1"
	err = r.UpdateCluster(ctx, updateParams)
	assert.Nil(t, err)

	diff, err := r.CompareConfig(ctx, application, cluster, nil, nil)
	assert.Nil(t, err)
	t.Logf("\n%v\n", diff)

	files, err := r.GetCluster(ctx, application, cluster, templateName)
	assert.Nil(t, err)
	t.Logf("%v", files.PipelineJSONBlob)
	t.Logf("%v", files.ApplicationJSONBlob)
	assert.Equal(t, files.PipelineJSONBlob, pipelineJSONBlob)
	assert.Equal(t, files.ApplicationJSONBlob, applicationJSONBlob)

	commit, err := r.GetConfigCommit(ctx, application, cluster)
	assert.Nil(t, err)
	t.Logf("%v", commit)

	diff, err = r.CompareConfig(ctx, application, cluster, &commit.Master, &commit.Gitops)
	assert.Nil(t, err)
	t.Logf("\n%v\n", diff)

	updateImageCommit, err := r.UpdateImage(ctx, application, cluster, templateName, "newImage")
	assert.Nil(t, err)
	t.Logf("%v", updateImageCommit)

	repoInfo := r.GetRepoInfo(ctx, application, cluster)
	assert.NotNil(t, repoInfo)
	t.Logf("%v", repoInfo)

	com, err := r.MergeBranch(ctx, application, cluster)
	assert.Nil(t, err)
	t.Logf("%v", com)

	com, err = r.UpdateRestartTime(ctx, application, cluster, templateName)
	assert.Nil(t, err)
	t.Logf("%v", com)

	envValue, err := r.GetEnvValue(ctx, application, cluster, templateName)
	assert.Nil(t, err)
	t.Logf("%v", envValue)

	com, err = r.UpdateImage(ctx, application, cluster, templateName, "newImage2")
	assert.Nil(t, err)
	t.Logf("%v", com)

	com, err = r.Rollback(ctx, application, cluster, updateImageCommit)
	assert.Nil(t, err)
	t.Logf("%v", com)

	err = r.UpdateTags(ctx, application, cluster, templateName, []*clustertagmodels.ClusterTag{
		{
			Key:   "a",
			Value: "b",
		},
	})
	assert.Nil(t, err)
}

func TestGetClusterValueFile(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	gitlabmockLib := gitlablibmock.NewMockInterface(mockCtrl)

	repoConfig := gitlabconf.Repo{
		Parent: &gitlabconf.Parent{
			Path: "horizon",
			ID:   1,
		},
		RecyclingParent: &gitlabconf.Parent{
			Path: "horizon-recycle",
			ID:   2,
		},
	}

	var clusterGitRepoInstance ClusterGitRepo // nolint

	clusterGitRepoInstance = &clusterGitRepo{
		gitlabLib:       gitlabmockLib,
		clusterRepoConf: &repoConfig,
		helmRepoMapper:  nil,
	}

	// 1. test gitlab get file error
	gitlabmockLib.EXPECT().GetFile(gomock.Any(), gomock.Any(),
		_branchMaster, gomock.Any()).Return(
		nil, errors.New("gitlab getFile error")).Times(4)
	clusterValueFiles, err := clusterGitRepoInstance.GetClusterValueFiles(context.TODO(),
		"app", "cluster")
	assert.Nil(t, clusterValueFiles)
	assert.NotNil(t, err)

	// 2. test gitlab return ok
	gitlabmockLib.EXPECT().GetFile(gomock.Any(), gomock.Any(), _branchMaster, gomock.Any()).Return(
		[]byte("cluster: xxx"), nil).Times(4)
	clusterValueFiles, err = clusterGitRepoInstance.GetClusterValueFiles(context.TODO(),
		"app", "cluster")
	assert.Nil(t, err)
	assert.NotNil(t, clusterValueFiles)

	// 3. test gitlab return 404
	gitlabmockLib.EXPECT().GetFile(gomock.Any(), gomock.Any(), _branchMaster, gomock.Any()).Return(
		[]byte("cluster: xxx"), nil).Times(3)
	var herr = herrors.E(
		"Test", http.StatusNotFound)
	gitlabmockLib.EXPECT().GetFile(gomock.Any(), gomock.Any(), _branchMaster, gomock.Any()).Return(
		nil, herr).Times(1)

	clusterValueFile1, err := clusterGitRepoInstance.GetClusterValueFiles(context.TODO(),
		"app", "cluster")
	assert.NotNil(t, clusterValueFile1)
	assert.Nil(t, err)
}
