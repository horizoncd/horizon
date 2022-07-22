package gitrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/config/gitlab"
	"github.com/stretchr/testify/assert"
)

/*
NOTE: gitlab params must set by environment variable.
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
	"rootGroupName": "xxx",
	"rootGroupID": xxx
}
EOF
)"
go test -v ./pkg/application/gitrepo

NOTE: when there is no GITLAB_PARAMS_FOR_TEST environment variable, skip this test.

*/
// nolint
var (
	ctx context.Context
	g   gitlablib.Interface
	r   ApplicationGitRepo

	rootGroupName string
	app           = "app"

	pipelineJSONBlob, applicationJSONBlob map[string]interface{}
	pipelineJSONStr                       = `{
            "buildxml":"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE project [<!ENTITY buildfile SYSTEM \"file:./build-user.xml\">]>\n<project basedir=\".\" default=\"deploy\" name=\"demo\">\n    <property name=\"ant\" value=\"ant\" />\n    <property name=\"baseline.dir\" value=\"${basedir}\"/>\n\n    <target name=\"package\">\n        <exec dir=\"${baseline.dir}\" executable=\"${ant}\" failonerror=\"true\">\n            <arg line=\"-buildfile overmind_build.xml -Denv=test -DenvName=mockserver.org\"/>\n        </exec>\n    </target>\n\n    <target name=\"deploy\">\n        <echo message=\"begin auto deploy......\"/>\n        <antcall target=\"package\"/>\n    </target>\n</project>"
        }`

	applicationJSONStr = `{
    "app":{
        "params":{
            "xmx":"512",
            "xms":"512",
            "maxPerm":"128",
            "mainClassName":"com.netease.horizon.WebApplication",
            "jvmExtra":"-Dserver.port=8080"
        },
        "resource":"x-small",
        "health":{
            "lifecycle":{
                "online":{
                    "url":"/online",
                    "timeoutSeconds":3,
                    "periodSeconds":15,
                    "retry":20
                },
                "offline":{
                    "url":"/offline",
                    "timeoutSeconds":3,
                    "periodSeconds":15,
                    "retry":20
                }
            },
            "probe":{
                "check":{
                    "url":"/check",
                    "initialDelaySeconds":200,
                    "timeoutSeconds":3,
                    "periodSeconds":15,
                    "failureThreshold":3
                },
                "status":{
                    "url":"/status",
                    "initialDelaySeconds":200,
                    "timeoutSeconds":3,
                    "periodSeconds":15,
                    "failureThreshold":3
                }
            },
            "port":8080
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

	ctx = context.WithValue(context.Background(), common.UserContextKey(), &userauth.DefaultInfo{
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
	r = &applicationGitlabRepo{
		gitlabLib: g,
		applicationRepoConf: &gitlab.Repo{
			Parent: &gitlab.Parent{
				Path: fmt.Sprintf("%v/%v", rootGroupName, "applications"),
				ID:   4280,
			},
			RecyclingParent: &gitlab.Parent{
				Path: fmt.Sprintf("%v/%v", rootGroupName, "recycling-applications"),
				ID:   4592,
			},
		},
	}

	defer func() {
		_ = r.DeleteApplication(ctx, app, 1)
		_ = g.DeleteGroup(ctx, fmt.Sprintf("%v/%v/%v-%v", rootGroupName, "recycling-applications", app, 1))
	}()

	err := r.CreateApplication(ctx, app, pipelineJSONBlob, applicationJSONBlob)
	assert.Nil(t, err)

	err = r.CreateApplication(ctx, app, pipelineJSONBlob, applicationJSONBlob)
	assert.NotNil(t, err)

	// update, exchange pipelineJSONBlob and applicationJSONBlob
	err = r.UpdateApplication(ctx, app, pipelineJSONBlob, applicationJSONBlob)
	assert.Nil(t, err)

	pipelineJSON, applicationJSON, err := r.GetApplication(ctx, app)
	assert.Nil(t, err)
	if reflect.DeepEqual(pipelineJSON, applicationJSONBlob) {
		t.Fatal("wrong pipeline")
	}

	if reflect.DeepEqual(applicationJSON, pipelineJSONBlob) {
		t.Fatal("wrong application")
	}

	env := "dev"
	pipelineJSON, applicationJSON, err = r.GetApplicationEnvTemplate(ctx, app, env)
	assert.Nil(t, err)
	if reflect.DeepEqual(pipelineJSON, applicationJSONBlob) {
		t.Fatal("wrong pipeline")
	}

	if reflect.DeepEqual(applicationJSON, pipelineJSONBlob) {
		t.Fatal("wrong application")
	}

	pipelineJSONBlob["test"] = "test"
	applicationJSONBlob["application"] = "application"
	err = r.UpdateApplicationEnvTemplate(ctx, app, env, pipelineJSONBlob, applicationJSONBlob)
	assert.Nil(t, err)

	pipelineJSON, applicationJSON, err = r.GetApplicationEnvTemplate(ctx, app, env)
	assert.Nil(t, err)
	if reflect.DeepEqual(pipelineJSON, applicationJSONBlob) {
		t.Fatal("wrong pipeline")
	}

	if reflect.DeepEqual(applicationJSON, pipelineJSONBlob) {
		t.Fatal("wrong application")
	}

	appToDelete := "appToDelete"
	err = r.CreateApplication(ctx, appToDelete, pipelineJSONBlob, applicationJSONBlob)
	assert.Nil(t, err)

	err = r.HardDeleteApplication(ctx, appToDelete)
	assert.Nil(t, err)
}
