package gitrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/middleware/user"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/config/gitlab"
	harbormodels "g.hz.netease.com/horizon/pkg/harbor/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"

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

	rootGroupName string

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

	g, err = gitlablib.New(p.Token, p.BaseURL, "")
	if err != nil {
		panic(err)
	}

	ctx = context.WithValue(context.Background(), user.Key(), &userauth.DefaultInfo{
		Name: "Tony",
	})

	rootGroupName = p.RootGroupName

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
	}
	application := "app"
	cluster := "cluster"
	params := &Params{
		Cluster:     cluster,
		HelmRepoURL: "https://helm.com",
		Environment: "test",
		RegionEntity: &regionmodels.RegionEntity{
			Region: &regionmodels.Region{
				Name:        "hz",
				DisplayName: "HZ",
			},
			Harbor: &harbormodels.Harbor{
				Server: "https://harbor.com",
			},
		},
		PipelineJSONBlob:    pipelineJSONBlob,
		ApplicationJSONBlob: applicationJSONBlob,
		TemplateRelease: &trmodels.TemplateRelease{
			TemplateName: "javaapp",
			Name:         "v1.0.0",
		},
		Application: &appmodels.Application{
			GroupID:  10,
			Name:     application,
			Priority: "P0",
		},
	}

	defer func() {
		_ = r.DeleteCluster(ctx, application, cluster, 1)
		_ = g.DeleteProject(ctx, fmt.Sprintf("%v/%v/%v/%v-%v", rootGroupName,
			"recycling-clusters", application, cluster, 1))
	}()
	err := r.CreateCluster(ctx, params)
	assert.Nil(t, err)

	params.Application.Priority = "P1"
	err = r.UpdateCluster(ctx, params)
	assert.Nil(t, err)

	diff, err := r.CompareConfig(ctx, application, params.Cluster)
	assert.Nil(t, err)
	t.Logf("\n%v\n", diff)
}
