package gitrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	gitlablib "github.com/horizoncd/horizon/lib/gitlab"
	gitlablibmock "github.com/horizoncd/horizon/mock/lib/gitlab"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	templateconfig "github.com/horizoncd/horizon/pkg/config/template"
	config "github.com/horizoncd/horizon/pkg/config/templaterepo"
	perror "github.com/horizoncd/horizon/pkg/errors"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	registrymodels "github.com/horizoncd/horizon/pkg/registry/models"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
	"github.com/horizoncd/horizon/pkg/templaterepo/chartmuseumbase"
	utilcommon "github.com/horizoncd/horizon/pkg/util/common"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
)

/*
go test -v ./pkg/cluster/gitrepo

NOTE: when there is no GITLAB_PARAMS_FOR_TEST environment variable, skip this test.

env name is GITLAB_PARAMS_FOR_TEST, and the value is a json string, look like:
{
	"token": "xxx",
	"baseURL": "http://cicd.mockserver.org",
	"rootGroupName": "xxx",
	"rootGroupID": xxx
}

1. token is used for auth, see https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html for more information.
2. baseURL is the basic URL for gitlab.
3. rootGroupName is a root group, our unit tests will do some operations under this group.
4. rootGroupID is the ID for this root group.


You can run this unit test just like this:

export GITLAB_PARAMS_FOR_TEST="$(cat <<\EOF
{
	"token": "xxx",
	"baseURL": "http://cicd.mockserver.org",
	"rootGroupName": "xxx"
}
EOF
)"
go test -v ./pkg/application/gitrepo

NOTE: when there is no GITLAB_PARAMS_FOR_TEST environment variable, skip this test.


NOTE: if your gitlab default branch is main.
set the env 'defaultBranch' to main
*/

// nolint
var (
	ctx           context.Context
	g             gitlablib.Interface
	defaultBranch string

	sshURL        string
	rootGroupName string
	rootGroup     *gitlab.Group
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
}

// nolint
func TestMain(m *testing.M) {
	var err error
	ctx = context.WithValue(context.Background(), common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
	})

	param := os.Getenv("GITLAB_PARAMS_FOR_TEST")
	if param == "" {
		return
	}

	defaultBranch = os.Getenv("defaultBranch")
	if defaultBranch == "" {
		defaultBranch = "master"
	}

	var p *Param
	if err := json.Unmarshal([]byte(param), &p); err != nil {
		panic(err)
	}

	sshURL = "ssh://gitlab.com"

	g, err = gitlablib.New(p.Token, p.BaseURL)
	if err != nil {
		panic(err)
	}
	rootGroup, err = g.GetGroup(ctx, p.RootGroupName)
	if err != nil {
		panic(err)
	}

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
	repo, _ := chartmuseumbase.NewRepo(config.Repo{Host: "https://harbor.cloudnative.com"})

	r, err := NewClusterGitlabRepo(ctx, rootGroup, repo, g, defaultBranch)
	assert.Nil(t, err)

	application := "app"
	cluster := "cluster"

	baseParams := &BaseParams{
		Cluster:             cluster,
		PipelineJSONBlob:    pipelineJSONBlob,
		ApplicationJSONBlob: applicationJSONBlob,
		TemplateRelease: &trmodels.TemplateRelease{
			TemplateName: templateName,
			ChartName:    templateName,
			ChartVersion: "v1.0.0",
		},
		Application: &appmodels.Application{
			GroupID:  10,
			Name:     application,
			Priority: "P0",
		},
		Environment: "test",
		RegionEntity: &regionmodels.RegionEntity{
			Region: &regionmodels.Region{
				Name:        "hz",
				DisplayName: "HZ",
				Server:      "https://k8s.com",
			},
			Registry: &registrymodels.Registry{
				Server: "https://harbor.com",
			},
		},
	}
	createParams := &CreateClusterParams{
		BaseParams: baseParams,
	}
	updateParams := &UpdateClusterParams{
		BaseParams: baseParams,
	}

	defer func() {
		_ = r.DeleteCluster(ctx, application, cluster, 1)
		_ = g.DeleteProject(ctx, fmt.Sprintf("%v/%v/%v/%v-%v", rootGroupName,
			common.GitopsGroupRecyclingClusters, application, cluster, 1))
	}()
	err = r.CreateCluster(ctx, createParams)
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

	template, err := r.GetClusterTemplate(ctx, application, cluster)
	assert.Nil(t, err)
	assert.Equal(t, template.Name, templateName)
	assert.Equal(t, template.Release, "v1.0.0")

	_, err = r.GetManifest(ctx, application, cluster, nil)
	assert.NotNil(t, err)
	_, ok := perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.True(t, ok)

	commit, err := r.GetConfigCommit(ctx, application, cluster)
	assert.Nil(t, err)
	t.Logf("%v", commit)

	diff, err = r.CompareConfig(ctx, application, cluster, &commit.Master, &commit.Gitops)
	assert.Nil(t, err)
	t.Logf("\n%v\n", diff)

	imageName := "newImage"

	updateImageCommit, err := r.UpdatePipelineOutput(ctx, application, cluster, templateName, PipelineOutput{
		Image: &imageName,
	})
	assert.Nil(t, err)
	t.Logf("%v", updateImageCommit)

	repoInfo := r.GetRepoInfo(ctx, application, cluster)
	assert.NotNil(t, repoInfo)
	t.Logf("%v", repoInfo)

	com, err := r.MergeBranch(ctx, application, cluster,
		GitOpsBranch, r.DefaultBranch(), utilcommon.UintPtr(123))
	assert.Nil(t, err)
	t.Logf("%v", com)

	com, err = r.UpdateRestartTime(ctx, application, cluster, templateName)
	assert.Nil(t, err)
	t.Logf("%v", com)

	envValue, err := r.GetEnvValue(ctx, application, cluster, templateName)
	assert.Nil(t, err)
	t.Logf("%v", envValue)

	imageName = "newImage2"
	com, err = r.UpdatePipelineOutput(ctx, application, cluster, templateName, PipelineOutput{
		Image: &imageName,
	})
	assert.Nil(t, err)
	t.Logf("%v", com)

	com, err = r.Rollback(ctx, application, cluster, updateImageCommit)
	assert.Nil(t, err)
	t.Logf("%v", com)

	err = r.UpdateTags(ctx, application, cluster, templateName, []*tagmodels.Tag{
		{
			Key:   "a",
			Value: "b",
		},
	})
	assert.Nil(t, err)

	restartTime, err := r.GetRestartTime(ctx, application, cluster, templateName)
	assert.Nil(t, err)
	assert.NotEmpty(t, restartTime)
}

func TestV2(t *testing.T) {
	repo, _ := chartmuseumbase.NewRepo(config.Repo{Host: "https://harbor.cloudnative.com"})
	r, err := NewClusterGitlabRepo(ctx, rootGroup, repo, g, defaultBranch)
	assert.Nil(t, err)

	application := "appv2"
	cluster := "clusterv2"

	baseParams := &BaseParams{
		Cluster:             cluster,
		PipelineJSONBlob:    nil,
		ApplicationJSONBlob: applicationJSONBlob,
		TemplateRelease: &trmodels.TemplateRelease{
			TemplateName: templateName,
			ChartName:    templateName,
			ChartVersion: "v1.0.0",
		},
		Application: &appmodels.Application{
			GroupID:  10,
			Name:     application,
			Priority: "P0",
		},
		Environment: "test",
		RegionEntity: &regionmodels.RegionEntity{
			Region: &regionmodels.Region{
				Name:        "hz",
				DisplayName: "HZ",
				Server:      "https://k8s.com",
			},
			Registry: &registrymodels.Registry{
				Server: "https://harbor.com",
			},
		},
		Version: common.MetaVersion2,
	}
	createParams := &CreateClusterParams{
		BaseParams: baseParams,
	}

	defer func() {
		_ = r.DeleteCluster(ctx, application, cluster, 1)
		_ = g.DeleteProject(ctx, fmt.Sprintf("%v/%v/%v/%v-%v", rootGroupName,
			"recycling-clusters", application, cluster, 1))
	}()
	err = r.CreateCluster(ctx, createParams)
	assert.Nil(t, err)
	files, err := r.GetCluster(ctx, application, cluster, templateName)
	t.Logf("%+v", files)
	assert.Nil(t, err)
	assert.NotNil(t, files.Manifest)
	assert.NotNil(t, files.ApplicationJSONBlob)
	assert.Nil(t, files.PipelineJSONBlob)
	t.Logf("%+v", files)

	manifest, err := r.GetManifest(ctx, application, cluster, nil)
	assert.Nil(t, err)
	assert.Equal(t, manifest.Version, common.MetaVersion2)

	// do not update Region \ update Version \ application yaml \ add pipeline
	baseParams.RegionEntity = nil
	baseParams.TemplateRelease.ChartVersion = "v2.0.0"
	var applicationJSONBlob2 map[string]interface{}
	applicationJSONStr2 := `{
    "app":{
        "health":{
            "check":"/api/test2",
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
            "resource":"middle"
        },
        "strategy":{
            "pauseType":"first",
            "stepsTotal":2
        }
    }
}`
	if err := json.Unmarshal([]byte(applicationJSONStr2), &applicationJSONBlob2); err != nil {
		panic(err)
	}
	baseParams.ApplicationJSONBlob = applicationJSONBlob2
	baseParams.PipelineJSONBlob = pipelineJSONBlob
	updateParam := &UpdateClusterParams{BaseParams: baseParams}
	err = r.UpdateCluster(ctx, updateParam)
	assert.Nil(t, err)

	files, err = r.GetCluster(ctx, application, cluster, templateName)
	t.Logf("%+v", files)
	assert.Nil(t, err)
	assert.NotNil(t, files.Manifest)
	assert.NotNil(t, files.ApplicationJSONBlob)
	assert.NotNil(t, files.PipelineJSONBlob)
}

func TestUpgradeToV2(t *testing.T) {
	repo, _ := chartmuseumbase.NewRepo(config.Repo{Host: "https://harbor.cloudnative.com"})
	r, err := NewClusterGitlabRepo(ctx, rootGroup, repo, g, defaultBranch)
	assert.Nil(t, err)

	application := "appUpgrade"
	cluster := "clusterUpgrade"

	baseParams := &BaseParams{
		Cluster:             cluster,
		PipelineJSONBlob:    pipelineJSONBlob,
		ApplicationJSONBlob: applicationJSONBlob,
		TemplateRelease: &trmodels.TemplateRelease{
			TemplateName: templateName,
			ChartName:    templateName,
			ChartVersion: "v1.0.0",
		},
		Application: &appmodels.Application{
			GroupID:  10,
			Name:     application,
			Priority: "P0",
		},
		Environment: "test",
		RegionEntity: &regionmodels.RegionEntity{
			Region: &regionmodels.Region{
				Name:        "hz",
				DisplayName: "HZ",
				Server:      "https://k8s.com",
			},
			Registry: &registrymodels.Registry{
				Server: "https://harbor.com",
			},
		},
	}
	createParams := &CreateClusterParams{
		BaseParams: baseParams,
	}
	defer func() {
		_ = r.DeleteCluster(ctx, application, cluster, 1)
		_ = g.DeleteProject(ctx, fmt.Sprintf("%v/%v/%v/%v-%v", rootGroupName,
			"recycling-clusters", application, cluster, 1))
	}()

	err = r.CreateCluster(ctx, createParams)
	assert.Nil(t, err)
	files, err := r.GetCluster(ctx, application, cluster, templateName)
	t.Logf("%+v", files)
	assert.Nil(t, err)
	assert.Nil(t, files.Manifest)
	assert.NotNil(t, files.ApplicationJSONBlob)
	assert.NotNil(t, files.PipelineJSONBlob)
	t.Logf("%+v", files)
	template, err := r.GetClusterTemplate(ctx, application, cluster)
	assert.Nil(t, err)
	assert.Equal(t, template.Name, templateName)
	assert.Equal(t, template.Release, "v1.0.0")
	targetTemplate := "rollout"
	targetRelease := "v1.2.0"
	upgradeCommit, err := r.UpgradeCluster(ctx, &UpgradeValuesParam{
		Application: application,
		Cluster:     cluster,
		Template:    template,
		TargetRelease: &trmodels.TemplateRelease{
			TemplateName: targetTemplate,
			ChartName:    targetTemplate,
			ChartVersion: targetRelease,
		},
		BuildConfig: &templateconfig.BuildConfig{
			Language:    "java",
			Environment: "javaapp",
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, upgradeCommit)
	files, err = r.GetCluster(ctx, application, cluster, templateName)
	t.Logf("%+v", files)
	assert.Nil(t, err)
	assert.NotNil(t, files.Manifest)
	assert.NotNil(t, files.ApplicationJSONBlob)
	assert.NotNil(t, files.PipelineJSONBlob)
	t.Logf("%+v", files)
	template, err = r.GetClusterTemplate(ctx, application, cluster)
	assert.Nil(t, err)
	assert.Equal(t, template.Name, targetTemplate)
	assert.Equal(t, template.Release, targetRelease)
}

func TestHardDeleteCluster(t *testing.T) {
	application := "app"
	cluster := "cluster"

	baseParams := &BaseParams{
		Cluster:             cluster,
		PipelineJSONBlob:    pipelineJSONBlob,
		ApplicationJSONBlob: applicationJSONBlob,
		TemplateRelease: &trmodels.TemplateRelease{
			TemplateName: templateName,
			ChartVersion: "v1.0.0",
		},
		Application: &appmodels.Application{
			GroupID:  10,
			Name:     application,
			Priority: "P0",
		},
		Environment: "test",
		RegionEntity: &regionmodels.RegionEntity{
			Region: &regionmodels.Region{
				Name:        "hz",
				DisplayName: "HZ",
				Server:      "https://k8s.com",
			},
			Registry: &registrymodels.Registry{
				Server: "https://harbor.com",
			},
		},
	}
	createParams := &CreateClusterParams{
		BaseParams: baseParams,
	}
	repo, _ := chartmuseumbase.NewRepo(config.Repo{Host: "https://harbor.cloudnative.com"})
	r, err := NewClusterGitlabRepo(ctx, rootGroup, repo, g, defaultBranch)
	assert.Nil(t, err)
	err = r.CreateCluster(ctx, createParams)
	assert.Nil(t, err)

	err = r.HardDeleteCluster(ctx, application, cluster)
	assert.Nil(t, err)
}

func TestGetClusterValueFile(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	gitlabmockLib := gitlablibmock.NewMockInterface(mockCtrl)
	gitlabmockLib.EXPECT().GetCreatedGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&gitlab.Group{}, nil).AnyTimes()

	var clusterGitRepoInstance ClusterGitRepo // nolint
	clusterGitRepoInstance, err := NewClusterGitlabRepo(ctx, rootGroup, &chartmuseumbase.Repo{},
		gitlabmockLib, defaultBranch)
	assert.Nil(t, err)

	// 1. test gitlab get file error
	gitlabmockLib.EXPECT().GetFile(gomock.Any(), gomock.Any(),
		defaultBranch, gomock.Any()).Return(
		nil, errors.New("gitlab getFile error")).Times(5)
	clusterValueFiles, err := clusterGitRepoInstance.GetClusterValueFiles(context.TODO(),
		"app", "cluster")
	assert.Nil(t, clusterValueFiles)
	assert.NotNil(t, err)

	// 2. test gitlab return ok
	gitlabmockLib.EXPECT().GetFile(gomock.Any(), gomock.Any(), defaultBranch, gomock.Any()).Return(
		[]byte("cluster: xxx"), nil).Times(5)
	clusterValueFiles, err = clusterGitRepoInstance.GetClusterValueFiles(context.TODO(),
		"app", "cluster")
	assert.Nil(t, err)
	assert.NotNil(t, clusterValueFiles)

	// 3. test gitlab return 404
	gitlabmockLib.EXPECT().GetFile(gomock.Any(), gomock.Any(), defaultBranch, gomock.Any()).Return(
		[]byte("cluster: xxx"), nil).Times(4)
	var herr = herrors.NewErrNotFound(herrors.GitlabResource, "test")
	gitlabmockLib.EXPECT().GetFile(gomock.Any(), gomock.Any(), defaultBranch, gomock.Any()).Return(
		nil, herr).Times(1)

	clusterValueFile1, err := clusterGitRepoInstance.GetClusterValueFiles(context.TODO(),
		"app", "cluster")
	assert.NotNil(t, clusterValueFile1)
	assert.Nil(t, err)
}

// nolint
func TestClusterGitRepo_UpdatePipelineOutput(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	templateName := "java"
	output := `
java:
  image: harbor.cloudnative.com/music-job-console/music-job-console-1:dev-d094e34f-20220118150928
`
	gitlabmockLib := gitlablibmock.NewMockInterface(mockCtrl)
	gitlabmockLib.EXPECT().GetFile(gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).Return(
		[]byte(output), nil).AnyTimes()
	gitlabmockLib.EXPECT().WriteFiles(gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx, pid, branch, commitMsg, startBranch, actions interface{}) (string, error) {
			output = actions.([]gitlablib.CommitAction)[0].Content
			return "", nil
		}).AnyTimes()
	gitlabmockLib.EXPECT().GetCreatedGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&gitlab.Group{}, nil).AnyTimes()

	var clusterGitRepoInstance ClusterGitRepo // nolint
	clusterGitRepoInstance, err := NewClusterGitlabRepo(ctx, rootGroup, &chartmuseumbase.Repo{},
		gitlabmockLib, defaultBranch)
	assert.Nil(t, err)

	expectedOutput := `java:
  git:
    branch: bbb
    commitID: ccc
    url: aaa
  image: harbor.cloudnative.com/music-job-console/music-job-console-1:dev-d094e34f-20221109170000
`
	url := "aaa"
	branch := "bbb"
	commit := "ccc"
	image := "harbor.cloudnative.com/music-job-console/music-job-console-1:dev-d094e34f-20221109170000"
	_, err = clusterGitRepoInstance.UpdatePipelineOutput(ctx, "", "", templateName, PipelineOutput{
		Git: &Git{
			URL:      &url,
			Branch:   &branch,
			CommitID: &commit,
		},
		Image: &image,
	})
	assert.Nil(t, err)
	fmt.Println(output)
	assert.Equal(t, expectedOutput, output)
}
