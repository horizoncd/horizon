package gitrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
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

	rootGroupName string
	rootGroup     *gitlab.Group
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

	pipelineJSONBlob2, applicationJSONBlob2 map[string]interface{}
	pipelineJSONStr2                        = `{
            "buildxml":"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE project [<!ENTITY buildfile SYSTEM \"file:./build-user.xml\">]>\n<project basedir=\".\" default=\"deploy\" name=\"demo\">\n    <property name=\"ant\" value=\"ant\" />\n    <property name=\"baseline.dir\" value=\"${basedir}\"/>\n\n    <target name=\"package\">\n        <exec dir=\"${baseline.dir}\" executable=\"${ant}\" failonerror=\"true\">\n            <arg line=\"-buildfile overmind_build.xml -Denv=test -DenvName=mockserver.org\"/>\n        </exec>\n    </target>\n\n    <target name=\"deploy\">\n        <echo message=\"begin auto deploy......\"/>\n        <antcall target=\"package\"/>\n    </target>\n</project>"
        }`

	applicationJSONStr2 = `{
    "app":{
        "params":{
            "xmx":"512",
            "xms":"512",
            "maxPerm":"128",
            "mainClassName":"com.netease.horizon.WebApplication",
            "jvmExtra":"-Dserver.port=8080"
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
func testInit() {
	var err error

	ctx = context.WithValue(context.Background(), common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
	})

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
	rootGroup, err = g.GetGroup(ctx, p.RootGroupName)
	if err != nil {
		panic(err)
	}
	rootGroupName = p.RootGroupName

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

	if err := json.Unmarshal([]byte(pipelineJSONStr2), &pipelineJSONBlob2); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(applicationJSONStr2), &applicationJSONBlob2); err != nil {
		panic(err)
	}
}

func TestV2(t *testing.T) {
	testInit()
	r, err := NewApplicationGitlabRepo2(ctx, rootGroup, g)
	assert.Nil(t, err)

	defer func() {
		_ = r.HardDeleteApplication(ctx, app)
		_ = g.DeleteGroup(ctx, fmt.Sprintf("%v/%v/%v-%v", rootGroupName, "recycling-applications", app, 1))
	}()

	versionV2 := "0.0.2"
	environment := "test"
	createReq := CreateOrUpdateRequest{
		Version:      versionV2,
		Environment:  environment,
		BuildConf:    pipelineJSONBlob,
		TemplateConf: applicationJSONBlob,
	}
	err = r.CreateOrUpdateApplication(ctx, app, createReq)
	assert.Nil(t, err)

	err = r.CreateOrUpdateApplication(ctx, app, createReq)
	assert.Nil(t, err)

	getReponse, err := r.GetApplication(ctx, app, environment)
	assert.Nil(t, err)
	t.Logf("%+v", getReponse.TemplateConf)
	t.Logf("%+v", applicationJSONBlob)

	assert.ObjectsAreEqual(getReponse.TemplateConf, applicationJSONBlob)
	assert.ObjectsAreEqual(getReponse.BuildConf, pipelineJSONBlob)

	versionV3 := "0.0.3"
	updateReq2 := CreateOrUpdateRequest{
		Version:      versionV3,
		Environment:  environment,
		BuildConf:    pipelineJSONBlob2,
		TemplateConf: nil,
	}
	err = r.CreateOrUpdateApplication(ctx, app, updateReq2)
	assert.Nil(t, err)
	getReponse2, err := r.GetApplication(ctx, app, environment)
	assert.Nil(t, err)
	assert.ObjectsAreEqual(getReponse2.BuildConf, pipelineJSONBlob2)
	assert.ObjectsAreEqual(getReponse2.TemplateConf, applicationJSONBlob)
	t.Logf("%+v", getReponse2.Manifest)

	updateReq3 := CreateOrUpdateRequest{
		Version:      versionV3,
		Environment:  environment,
		BuildConf:    nil,
		TemplateConf: applicationJSONBlob2,
	}
	err = r.CreateOrUpdateApplication(ctx, app, updateReq3)
	assert.Nil(t, err)
	getReponse3, err := r.GetApplication(ctx, app, environment)
	assert.Nil(t, err)
	assert.ObjectsAreEqual(getReponse3.BuildConf, pipelineJSONBlob2)
	assert.ObjectsAreEqual(getReponse3.TemplateConf, applicationJSONBlob2)
	t.Logf("%+v", getReponse3.Manifest)

	err = r.HardDeleteApplication(ctx, app)
	assert.Nil(t, err)
}
