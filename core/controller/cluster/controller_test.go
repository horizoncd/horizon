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

package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	tektoncollectormock "github.com/horizoncd/horizon/mock/pkg/cluster/tekton/collector"
	clustercd "github.com/horizoncd/horizon/pkg/cd"
	templatemodels "github.com/horizoncd/horizon/pkg/template/models"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/config"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/core/middleware/requestid"
	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	applicationgitrepomock "github.com/horizoncd/horizon/mock/pkg/application/gitrepo"
	applicationmanangermock "github.com/horizoncd/horizon/mock/pkg/application/manager"
	cdmock "github.com/horizoncd/horizon/mock/pkg/cd"
	commitmock "github.com/horizoncd/horizon/mock/pkg/cluster/code"
	clustergitrepomock "github.com/horizoncd/horizon/mock/pkg/cluster/gitrepo"
	clustermanagermock "github.com/horizoncd/horizon/mock/pkg/cluster/manager"
	registrymock "github.com/horizoncd/horizon/mock/pkg/cluster/registry"
	registryftymock "github.com/horizoncd/horizon/mock/pkg/cluster/registry/factory"
	tektonmock "github.com/horizoncd/horizon/mock/pkg/cluster/tekton"
	tektonftymock "github.com/horizoncd/horizon/mock/pkg/cluster/tekton/factory"
	outputmock "github.com/horizoncd/horizon/mock/pkg/templaterelease/output"
	trschemamock "github.com/horizoncd/horizon/mock/pkg/templaterelease/schema"
	appgitrepo "github.com/horizoncd/horizon/pkg/application/gitrepo"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	gitconfig "github.com/horizoncd/horizon/pkg/config/git"
	templateconfig "github.com/horizoncd/horizon/pkg/config/template"
	tokenconfig "github.com/horizoncd/horizon/pkg/config/token"
	envmodels "github.com/horizoncd/horizon/pkg/environment/models"
	"github.com/horizoncd/horizon/pkg/environment/service"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	"github.com/horizoncd/horizon/pkg/git"
	"github.com/horizoncd/horizon/pkg/git/gitlab"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	groupservice "github.com/horizoncd/horizon/pkg/group/service"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	prmodels "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	registrydao "github.com/horizoncd/horizon/pkg/registry/dao"
	registrymodels "github.com/horizoncd/horizon/pkg/registry/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	tmodel "github.com/horizoncd/horizon/pkg/tag/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
	trschema "github.com/horizoncd/horizon/pkg/templaterelease/schema"
	gitlabschema "github.com/horizoncd/horizon/pkg/templaterelease/schema/gitlab"
	schematagmodel "github.com/horizoncd/horizon/pkg/templateschematag/models"
	tokenmodels "github.com/horizoncd/horizon/pkg/token/models"
	tokenservice "github.com/horizoncd/horizon/pkg/token/service"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	userservice "github.com/horizoncd/horizon/pkg/user/service"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"gopkg.in/yaml.v3"
)

// nolint
var (
	ctx                                   context.Context
	c                                     *controller
	pr                                    *v1beta1.PipelineRun
	applicationSchema, pipelineSchema     map[string]interface{}
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}
	commitGetter                          codemodels.GitGetter
	applicationSchemaJSON                 = `{
    "type":"object",
    "properties":{
        "app":{
            "title":"",
            "type":"object",
            "properties":{
                "spec":{
                    "type":"object",
                    "title":"规格",
                    "properties":{
                        "replicas":{
                            "type":"integer",
                            "title":"副本数"
                        },
                        "resource":{
                            "type":"string",
                            "title":"规格",
                            "oneOf":[
                                {
                                    "enum":[
                                        "x-small"
                                    ],
                                    "title":"x-small(1C2G)"
                                },
                                {
                                    "enum":[
                                        "small"
                                    ],
                                    "title":"small(2C4G)"
                                },
                                {
                                    "enum":[
                                        "middle"
                                    ],
                                    "title":"middle(4C8G)"
                                },
                                {
                                    "enum":[
                                        "large"
                                    ],
                                    "title":"large(8C16G)"
                                },
                                {
                                    "enum":[
                                        "x-large"
                                    ],
                                    "title":"x-large(16C32G)"
                                }
                            ]
                        }
                    }
                },
                "strategy":{
                    "type":"object",
                    "title":"发布策略",
                    "properties":{
                        "stepsTotal":{
                            "type":"integer",
                            "title":"发布批次（多批次情况下，第一批默认为1个实例）",
                            "enum":[
                                1,
                                2,
                                3,
                                4,
                                5
                            ]
                        },
                        "pauseType":{
                            "type":"string",
                            "title":"暂停策略",
                            "oneOf":[
                                {
                                    "enum":[
                                        "first"
                                    ],
                                    "title":"第一批暂停"
                                },
                                {
                                    "enum":[
                                        "all"
                                    ],
                                    "title":"全部暂停"
                                },
                                {
                                    "enum":[
                                        "none"
                                    ],
                                    "title":"全不暂停"
                                }
                            ]
                        }
                    }
                },
                "params":{
                    "title":"参数",
                    "type":"object",
                    "properties":{
                        "mainClassName":{
                            "type":"string"
                        },
                        "xmx":{
                            "type":"string",
                            "default":"512",
                            "pattern":"^\\d*$"
                        },
                        "xms":{
                            "type":"string",
                            "default":"512",
                            "pattern":"^\\d*$"
                        },
                        "maxPerm":{
                            "type":"string",
                            "default":"512",
                            "pattern":"^\\d*$"
                        },
                        "xdebugAddress":{
                            "type":"string",
                            "pattern":"^\\d*$"
                        },
                        "jvmExtra":{
                            "type":"string"
                        }
                    },
                    "required":[
                        "mainClassName"
                    ]
                },
                "health":{
                    "title":"健康检查",
                    "type":"object",
                    "properties":{
                        "port":{
                            "type":"integer",
                            "minimum":1024,
                            "maximum":65535
                        },
                        "check":{
                            "type":"string",
                            "pattern":"^/.*$",
                            "title":"存活状态",
                            "description":"存活状态会在应用运行期间检测应用健康情况，检测失败时会对应用进行重启"
                        },
                        "status":{
                            "type":"string",
                            "pattern":"^/.*$",
                            "title":"就绪状态",
                            "description":"就绪状态会在应用运行期间检测应用是否处于上线状态，检测失败时显示下线状态"
                        },
                        "online":{
                            "type":"string",
                            "pattern":"^/.*$",
                            "title":"上线",
                            "description":"上线接口会在应用启动之后进行调用，如果调用失败，则应用启动失败"
                        },
                        "offline":{
                            "type":"string",
                            "pattern":"^/.*$",
                            "title":"下线",
                            "description":"下线接口会在应用停止之前进行调用，如果调用失败，则忽略"
                        }
                    },
                    "dependencies":{
                        "check":[
                            "port"
                        ],
                        "status":[
                            "port"
                        ],
                        "online":[
                            "port"
                        ],
                        "offline":[
                            "port"
                        ]
                    }
                }
            }
        }
    }
}
`
	pipelineSchemaJSON = `{
  "type": "object",
  "title": "Ant",
  "properties": {
    "buildxml": {
      "title": "build.xml",
      "type": "string",
      "default": "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE project [<!ENTITY buildfile SYSTEM \"file:./build-user.xml\">]>\n<project basedir=\".\" default=\"deploy\" name=\"demo\">\n    <property name=\"ant\" value=\"ant\" />\n    <property name=\"baseline.dir\" value=\"${basedir}\"/>\n\n    <target name=\"package\">\n        <exec dir=\"${baseline.dir}\" executable=\"${ant}\" failonerror=\"true\">\n            <arg line=\"-buildfile overmind_build.xml -Denv=test -DenvName=qa-game.cloudnative.com\"/>\n        </exec>\n    </target>\n\n    <target name=\"deploy\">\n        <echo message=\"begin auto deploy......\"/>\n        <antcall target=\"package\"/>\n    </target>\n</project>"
    }
  }
}
`

	pipelineJSONStr = `{
		"buildxml":"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE project [<!ENTITY buildfile SYSTEM \"file:./build-user.xml\">]>\n<project basedir=\".\" default=\"deploy\" name=\"demo\">\n    <property name=\"ant\" value=\"ant\" />\n    <property name=\"baseline.dir\" value=\"${basedir}\"/>\n\n    <target name=\"package\">\n        <exec dir=\"${baseline.dir}\" executable=\"${ant}\" failonerror=\"true\">\n            <arg line=\"-buildfile overmind_build.xml -Denv=test -DenvName=qa-game.cloudnative.com\"/>\n        </exec>\n    </target>\n\n    <target name=\"deploy\">\n        <echo message=\"begin auto deploy......\"/>\n        <antcall target=\"package\"/>\n    </target>\n</project>"
	}`
	applicationJSONStr = `{
    "app":{
        "spec":{
            "replicas":1,
            "resource":"small"
        },
        "strategy":{
            "stepsTotal":1,
            "pauseType":"first"
        },
        "params":{
            "xmx":"512",
            "xms":"512",
            "maxPerm":"128",
            "mainClassName":"com.netease.horizon.WebApplication",
            "jvmExtra":"-Dserver.port=8080"
        },
        "health":{
            "check":"/api/test",
            "status":"/health/status",
            "online":"/health/online",
            "offline":"/health/offline",
            "port":8080
        }
    }
}`

	prJson = `{
        "metadata":{
            "name":"test-music-docker-q58rp",
            "namespace":"tekton-resources",
            "creationTimestamp": "2021-07-16T08:51:54Z",
            "labels":{
                "app.kubernetes.io/managed-by":"Helm",
                "tekton.dev/pipeline":"default",
                "triggers.tekton.dev/eventlistener":"default-listener",
                "triggers.tekton.dev/trigger":"",
                "triggers.tekton.dev/triggers-eventid":"cttzw"
            }
        },
        "status":{
            "conditions":[
                {
                    "type":"Succeeded",
                    "status":"True",
                    "lastTransitionTime":"2021-06-24T06:38:18Z",
                    "reason":"Succeeded",
                    "message":"Tasks Completed: 2 (Failed: 0, Cancelled 0), Skipped: 0"
                }
            ],
            "startTime":"2021-06-24T06:36:11Z",
            "completionTime":"2021-06-24T06:38:18Z",
            "taskRuns":{
                "test-music-docker-q58rp-build-g8khd":{
                    "pipelineTaskName":"build",
                    "status":{
                        "conditions":[
                            {
                                "type":"Succeeded",
                                "status":"True",
                                "lastTransitionTime":"2021-06-24T06:36:43Z",
                                "reason":"Succeeded",
                                "message":"All Steps have completed executing"
                            }
                        ],
                        "podName":"test-music-docker-q58rp-build-g8khd-pod-mwsld",
                        "startTime":"2021-06-24T06:36:11Z",
                        "completionTime":"2021-06-24T06:36:43Z",
                        "steps":[
                            {
                                "terminated":{
                                    "exitCode":0,
                                    "reason":"Completed",
                                    "startedAt":"2021-06-24T06:36:18Z",
                                    "finishedAt":"2021-06-24T06:36:26Z",
                                    "containerID":"docker://3cccbd086c26e83e41fe8fcd86ef4e0f42e3963371c581e458df223b94da8d1e"
                                },
                                "name":"git",
                                "container":"step-git",
                                "imageID":"docker-pullable://harbor.cloudnative.com/cloudnative/library/tekton-builder@sha256:14194e518981f5d893b85e170a28ba8aa80c2c610f63cfba814b6a460f48dc29"
                            },
                            {
                                "terminated":{
                                    "exitCode":0,
                                    "reason":"Completed",
                                    "startedAt":"2021-06-24T06:36:26Z",
                                    "finishedAt":"2021-06-24T06:36:34Z",
                                    "containerID":"docker://58d06c0a4bfa8212620e5a85a42e9af0768a4adb9ade2219dc75aee4d650ff23"
                                },
                                "name":"compile",
                                "container":"step-compile",
                                "imageID":"docker-pullable://harbor.cloudnative.com/cloudnative/library/tekton-builder@sha256:14194e518981f5d893b85e170a28ba8aa80c2c610f63cfba814b6a460f48dc29"
                            },
                            {
                                "terminated":{
                                    "exitCode":0,
                                    "reason":"Completed",
                                    "message":"[{\"key\":\"properties\",\"value\":\"harbor.cloudnative.com/test-music-docker:helloworld-b1f57848-20210624143634 ssh://git@cloudnative.com:22222/demo/springboot-demo.git helloworld b1f578488e3123e97ec00b671db143fb8f0abecf\",\"type\":\"TaskRunResult\"}]",
                                    "startedAt":"2021-06-24T06:36:34Z",
                                    "finishedAt":"2021-06-24T06:36:42Z",
                                    "containerID":"docker://9189624ad3981fd738ec5bf286f1fc5b688d71128b9827820ebc2541b2801dae"
                                },
                                "name":"image",
                                "container":"step-image",
                                "imageID":"docker-pullable://harbor.cloudnative.com/cloudnative/library/kaniko-executor@sha256:473d6dfb011c69f32192e668d86a47c0235791e7e857c870ad70c5e86ec07e8c"
                            }
                        ]
                    }
                },
                "test-music-docker-q58rp-deploy-xzjkg":{
                    "pipelineTaskName":"deploy",
                    "status":{
                        "conditions":[
                            {
                                "type":"Succeeded",
                                "status":"True",
                                "lastTransitionTime":"2021-06-24T06:38:18Z",
                                "reason":"Succeeded",
                                "message":"All Steps have completed executing"
                            }
                        ],
                        "podName":"test-music-docker-q58rp-deploy-xzjkg-pod-zkcc4",
                        "startTime":"2021-06-24T06:36:43Z",
                        "completionTime":"2021-06-24T06:38:18Z",
                        "steps":[
                            {
                                "terminated":{
                                    "exitCode":0,
                                    "reason":"Completed",
                                    "startedAt":"2021-06-24T06:36:48Z",
                                    "finishedAt":"2021-06-24T06:38:18Z",
                                    "containerID":"docker://fb2579fe83579e1918b5dcedc35f3686cad8ac632cc750d6d92f556341b5f7bb"
                                },
                                "name":"deploy",
                                "container":"step-deploy",
                                "imageID":"docker-pullable://harbor.cloudnative.com/cloudnative/library/tekton-builder@sha256:14194e518981f5d893b85e170a28ba8aa80c2c610f63cfba814b6a460f48dc29"
                            }
                        ]
                    }
                }
            }
        }
    }
	`

	db, _   = orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
)

const secondsInOneDay = 24 * 3600

// nolint
func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&appmodels.Application{}, &models.Cluster{}, &groupmodels.Group{},
		&trmodels.TemplateRelease{}, &membermodels.Member{}, &usermodels.User{},
		&registrymodels.Registry{}, eventmodels.Event{}, &templatemodels.Template{},
		&regionmodels.Region{}, &envregionmodels.EnvironmentRegion{}, &eventmodels.Event{},
		&prmodels.Pipelinerun{}, &schematagmodel.ClusterTemplateSchemaTag{}, &tmodel.Tag{},
		&envmodels.Environment{}, &tokenmodels.Token{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   uint(1),
	})
	ctx = context.WithValue(ctx, requestid.HeaderXRequestID, "requestid")

	if err := json.Unmarshal([]byte(applicationSchemaJSON), &applicationSchema); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(pipelineSchemaJSON), &pipelineSchema); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(pipelineJSONStr), &pipelineJSONBlob); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(applicationJSONStr), &applicationJSONBlob); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(prJson), &pr); err != nil {
		panic(err)
	}

	commitGetter, _ = codemodels.NewGitGetter(ctx, []*gitconfig.Repo{
		{
			Kind:  gitlab.Kind,
			URL:   "https://cloudnative.com",
			Token: "123456",
		},
	})
	os.Exit(m.Run())
}

func TestAll(t *testing.T) {
	t.Run("Test", test)
	t.Run("TestV2", testV2)
	t.Run("TestUpgrade", testUpgrade)
	t.Run("TestGetClusterOutPut", testGetClusterOutPut)
	t.Run("TestRenderOutPutObject", testRenderOutPutObject)
	t.Run("TestRenderOutPutObjectMissingKey", testRenderOutPutObjectMissingKey)
	t.Run("TestIsClusterActuallyHealthy", testIsClusterActuallyHealthy)
	t.Run("TestImageURL", testImageURL)
	t.Run("TestPinyin", testPinyin)
	t.Run("TestListClusterByNameFuzzily", testListClusterByNameFuzzily)
	t.Run("TestListUserClustersByNameFuzzily", testListUserClustersByNameFuzzily)
	t.Run("TestListClusterWithExpiry", testListClusterWithExpiry)
	t.Run("TestControllerFreeOrDeleteClusterFailed", testControllerFreeOrDeleteClusterFailed)
	t.Run("TestGetClusterStatusV2", testGetClusterStatusV2)
}

// nolint
func test(t *testing.T) {
	// for test
	conf := config.Config{}
	param := param.Param{
		AutoFreeSvc: service.New([]string{"test", "dev"}),
		Manager:     managerparam.InitManager(nil),
	}
	NewController(&conf, &param)

	templateName := "javaapp"
	mockCtl := gomock.NewController(t)
	clusterGitRepo := clustergitrepomock.NewMockClusterGitRepo(mockCtl)
	applicationGitRepo := applicationgitrepomock.NewMockApplicationGitRepo2(mockCtl)
	cd := cdmock.NewMockLegacyCD(mockCtl)
	k8sutil := cdmock.NewMockK8sUtil(mockCtl)
	tektonFty := tektonftymock.NewMockFactory(mockCtl)
	registryFty := registryftymock.NewMockRegistryGetter(mockCtl)
	commitGetter := commitmock.NewMockGitGetter(mockCtl)
	tagManager := manager.TagManager

	templateSchemaGetter := trschemamock.NewMockGetter(mockCtl)
	expectparams := make(map[string]string)
	expectparams[gitlabschema.ClusterIDKey] = "1"

	templateSchemaGetter.EXPECT().GetTemplateSchema(gomock.Any(), templateName, "v1.0.0", gomock.Any()).
		Return(&trschema.Schemas{
			Application: &trschema.Schema{
				JSONSchema: applicationSchema,
			},
			Pipeline: &trschema.Schema{
				JSONSchema: pipelineSchema,
			},
		}, nil).AnyTimes()
	templateSchemaGetter.EXPECT().GetTemplateSchema(gomock.Any(), templateName, "v1.0.1", gomock.Any()).
		Return(&trschema.Schemas{
			Application: &trschema.Schema{
				JSONSchema: applicationSchema,
			},
			Pipeline: &trschema.Schema{
				JSONSchema: pipelineSchema,
			},
		}, nil).AnyTimes()

	appMgr := manager.ApplicationManager
	trMgr := manager.TemplateReleaseManager
	envMgr := manager.EnvMgr
	regionMgr := manager.RegionMgr
	groupMgr := manager.GroupManager
	registryDAO := registrydao.NewDAO(db)
	envRegionMgr := manager.EnvRegionMgr

	// init data
	group, err := groupMgr.Create(ctx, &groupmodels.Group{
		Name:     "group",
		Path:     "group",
		ParentID: 0,
	})
	assert.Nil(t, err)
	assert.NotNil(t, group)

	gitURL := "ssh://git.com"
	application, err := appMgr.Create(ctx, &appmodels.Application{
		GroupID:         group.ID,
		Name:            "app",
		Priority:        "P3",
		GitURL:          gitURL,
		GitSubfolder:    "/test",
		GitRef:          "master",
		Template:        "javaapp",
		TemplateRelease: "v1.0.0",
	}, nil)

	tr, err := trMgr.Create(ctx, &trmodels.TemplateRelease{
		TemplateName: templateName,
		Name:         "v1.0.0",
		ChartVersion: "v1.0.0-test",
		ChartName:    templateName,
	})
	assert.Nil(t, err)
	assert.NotNil(t, tr)

	id, err := registryDAO.Create(ctx, &registrymodels.Registry{
		Server: "https://harbor.com",
		Token:  "xxx",
	})
	assert.Nil(t, err)
	assert.NotNil(t, id)

	region, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
		RegistryID:  id,
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	er, err := envRegionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: "test",
		RegionName:      "hz",
	})
	er, err = envRegionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: "dev",
		RegionName:      "hz",
	})
	assert.Nil(t, err)
	assert.NotNil(t, er)

	env, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "dev",
		DisplayName: "开发",
	})
	env, err = envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "test",
		DisplayName: "开发",
	})
	assert.Nil(t, err)
	assert.NotNil(t, env)

	c = &controller{
		clusterMgr:           manager.ClusterMgr,
		clusterGitRepo:       clusterGitRepo,
		commitGetter:         commitGetter,
		cd:                   cd,
		k8sutil:              k8sutil,
		applicationMgr:       appMgr,
		templateReleaseMgr:   trMgr,
		templateSchemaGetter: templateSchemaGetter,
		envMgr:               envMgr,
		envRegionMgr:         envRegionMgr,
		regionMgr:            regionMgr,
		autoFreeSvc:          param.AutoFreeSvc,
		groupSvc:             groupservice.NewService(manager),
		pipelinerunMgr:       manager.PipelinerunMgr,
		tektonFty:            tektonFty,
		registryFty:          registryFty,
		userManager:          manager.UserManager,
		userSvc:              userservice.NewService(manager),
		schemaTagManager:     manager.ClusterSchemaTagMgr,
		tagMgr:               tagManager,
		applicationGitRepo:   applicationGitRepo,
		eventMgr:             manager.EventManager,
		tokenSvc: tokenservice.NewService(manager, tokenconfig.Config{
			JwtSigningKey:         "horizon",
			CallbackTokenExpireIn: time.Hour * 2,
		}),
	}

	commitGetter.EXPECT().GetHTTPLink(gomock.Any()).Return("https://cloudnative.com:22222/demo/springboot-demo", nil).AnyTimes()
	applicationGitRepo.EXPECT().GetApplication(ctx, gomock.Any(), gomock.Any()).
		Return(&appgitrepo.GetResponse{
			Manifest:     nil,
			BuildConf:    pipelineJSONBlob,
			TemplateConf: applicationJSONBlob,
		}, nil).AnyTimes()
	clusterGitRepo.EXPECT().CreateCluster(ctx, gomock.Any()).Return(nil).Times(2)
	clusterGitRepo.EXPECT().UpdateCluster(ctx, gomock.Any()).Return(nil).Times(1)
	clusterGitRepo.EXPECT().GetCluster(ctx, "app",
		"app-cluster", templateName).Return(&gitrepo.ClusterFiles{
		PipelineJSONBlob:    pipelineJSONBlob,
		ApplicationJSONBlob: applicationJSONBlob,
	}, nil).AnyTimes()
	clusterGitRepo.EXPECT().GetCluster(ctx, "app",
		"app-cluster-mergepatch", "javaapp").Return(&gitrepo.ClusterFiles{
		PipelineJSONBlob:    pipelineJSONBlob,
		ApplicationJSONBlob: applicationJSONBlob,
	}, nil).AnyTimes()
	clusterGitRepo.EXPECT().GetConfigCommit(ctx, gomock.Any(), gomock.Any()).Return(&gitrepo.ClusterCommit{
		Master: "master-commit",
		Gitops: "gitops-commit",
	}, nil).AnyTimes()
	clusterGitRepo.EXPECT().GetEnvValue(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(&gitrepo.EnvValue{
		Namespace: "test-1",
	}, nil).AnyTimes()
	clusterGitRepo.EXPECT().GetRepoInfo(ctx, gomock.Any(), gomock.Any()).Return(&gitrepo.RepoInfo{
		GitRepoURL: "ssh://xxxx",
		ValueFiles: []string{},
	}).AnyTimes()
	imageName := "image"
	clusterGitRepo.EXPECT().UpdatePipelineOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("image-commit", nil).AnyTimes()
	clusterGitRepo.EXPECT().DefaultBranch().Return("master").AnyTimes()
	cd.EXPECT().CreateCluster(ctx, gomock.Any()).Return(nil).AnyTimes()

	clusterGitRepo.EXPECT().UpdateTags(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	createClusterRequest := &CreateClusterRequest{
		Base: &Base{
			Description: "cluster description",
			Git: &codemodels.Git{
				Branch: "develop",
			},
			TemplateInput: &TemplateInput{
				Application: applicationJSONBlob,
				Pipeline:    pipelineJSONBlob,
			},
			Tags: tagmodels.TagsBasic{
				{
					Key:   "key1",
					Value: "value1",
				},
			},
		},
		Name:       "app-cluster",
		ExpireTime: "24h0m0s",
	}

	resp, err := c.CreateCluster(ctx, application.ID, "test", "hz", createClusterRequest, false)
	assert.Nil(t, err)
	t.Logf("%v", resp.ExpireTime)

	createClusterRequest.Name = "app-cluster-new"
	_, err = c.CreateCluster(ctx, application.ID, "dev", "hz", createClusterRequest, false)
	assert.Nil(t, err)
	b, _ := json.MarshalIndent(resp, "", "  ")
	t.Logf("%v", string(b))

	assert.Equal(t, resp.Git.URL, gitURL)
	assert.Equal(t, resp.Git.Branch, "develop")
	assert.Equal(t, resp.Git.Subfolder, "/test")
	assert.Equal(t, resp.FullPath, "/group/app/app-cluster")
	t.Logf("%v", resp.ExpireTime)

	UpdateGitURL := "ssh://git@cloudnative.com:22222/demo/springboot-demo.git"
	updateClusterRequest := &UpdateClusterRequest{
		Base: &Base{
			Description: "new description",
			Git: &codemodels.Git{
				URL:       UpdateGitURL,
				Subfolder: "/new",
				Branch:    "new",
			},
			Tags: tagmodels.TagsBasic{
				{
					Key:   "key1",
					Value: "value2",
				},
			},
			TemplateInput: &TemplateInput{
				Application: applicationJSONBlob,
				Pipeline:    pipelineJSONBlob,
			},
			Template: &Template{
				Name:    "tomcat7_jdk8",
				Release: "v1.0.1",
			},
		},
	}

	newTr, err := trMgr.Create(ctx, &trmodels.TemplateRelease{
		TemplateName: templateName,
		ChartName:    templateName,
		Name:         "v1.0.1",
		ChartVersion: "v1.0.1",
	})
	assert.Nil(t, err)
	assert.NotNil(t, newTr)

	resp, err = c.UpdateCluster(ctx, resp.ID, updateClusterRequest, false)
	assert.Nil(t, err)
	assert.Equal(t, resp.Git.URL, UpdateGitURL)
	assert.Equal(t, resp.Git.Branch, "new")
	assert.Equal(t, resp.Git.Subfolder, "/new")
	assert.Equal(t, resp.FullPath, "/group/app/app-cluster")
	// NOTE: template name cannot be edited! template release can be edited
	assert.Equal(t, resp.Template.Name, "javaapp")
	assert.Equal(t, resp.Template.Release, "v1.0.1")
	assert.Equal(t, 1, len(resp.Base.Tags))
	assert.Equal(t, "value2", resp.Base.Tags[0].Value)

	resp, err = c.GetCluster(ctx, resp.ID)
	assert.Nil(t, err)
	assert.Equal(t, "24h0m0s", resp.ExpireTime)
	assert.Equal(t, resp.Git.URL, UpdateGitURL)
	assert.Equal(t, resp.Git.Branch, "new")
	assert.Equal(t, resp.Git.Subfolder, "/new")
	assert.Equal(t, resp.FullPath, "/group/app/app-cluster")
	// NOTE: template name cannot be edited! template release can be edited
	assert.Equal(t, resp.Template.Name, "javaapp")
	assert.Equal(t, resp.Template.Release, "v1.0.1")
	assert.Equal(t, resp.TemplateInput.Application, applicationJSONBlob)
	assert.Equal(t, resp.TemplateInput.Pipeline, pipelineJSONBlob)
	assert.Equal(t, 1, len(resp.Base.Tags))
	assert.Equal(t, "value2", resp.Base.Tags[0].Value)

	resp, err = c.UpdateCluster(ctx, resp.ID, &UpdateClusterRequest{
		Base:       &Base{},
		ExpireTime: "48h0m0s",
	}, false)
	assert.Nil(t, err)
	resp, err = c.GetCluster(ctx, resp.ID)
	assert.Nil(t, err)
	assert.Equal(t, "48h0m0s", resp.ExpireTime)
	assert.Equal(t, 1, len(resp.Base.Tags))

	resp, err = c.UpdateCluster(ctx, resp.ID, &UpdateClusterRequest{
		Base:       &Base{Tags: tagmodels.TagsBasic{}},
		ExpireTime: "48h0m0s",
	}, false)
	assert.Nil(t, err)
	resp, err = c.GetCluster(ctx, resp.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp.Base.Tags))

	count, respList, err := c.ListByApplication(ctx,
		q.New(q.KeyWords{
			common.ParamApplicationID:      application.ID,
			common.ClusterQueryEnvironment: []string{"test"},
		}))
	assert.Nil(t, err)
	t.Logf("%+v", respList)
	assert.Equal(t, 1, count)
	assert.Equal(t, respList[0].Template.Name, "javaapp")
	assert.Equal(t, respList[0].Template.Release, "v1.0.1")

	count, respList, err = c.ListByApplication(ctx,
		q.New(q.KeyWords{common.ParamApplicationID: application.ID}))
	assert.Nil(t, err)
	assert.Equal(t, count, 2)
	t.Logf("%+v", respList[0].Scope)
	t.Logf("%+v", respList[1].Scope)

	count, respList, err = c.ListByApplication(ctx,
		q.New(q.KeyWords{
			common.ParamApplicationID:      application.ID,
			common.ClusterQueryEnvironment: []string{"test", "dev"},
		}))
	assert.Nil(t, err)
	assert.Equal(t, count, 2)
	t.Logf("%+v", respList[0].Scope)
	t.Logf("%+v", respList[1].Scope)

	getByName, err := c.GetClusterByName(ctx, "app-cluster")
	assert.Nil(t, err)
	t.Logf("%v", getByName)

	tekton := tektonmock.NewMockInterface(mockCtl)
	tektonFty.EXPECT().GetTekton(gomock.Any()).Return(tekton, nil).AnyTimes()
	tekton.EXPECT().CreatePipelineRun(ctx, gomock.Any()).Return("abc", nil).Times(2)
	tekton.EXPECT().GetPipelineRunByID(ctx, gomock.Any()).Return(pr, nil).AnyTimes()
	tektonCollector := tektoncollectormock.NewMockInterface(mockCtl)

	tektonFty.EXPECT().GetTektonCollector(gomock.Any()).Return(tektonCollector, nil).AnyTimes()
	tektonCollector.EXPECT().GetPipelineRun(ctx, gomock.Any()).Return(pr, nil).AnyTimes()

	commitGetter.EXPECT().GetCommit(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(&git.Commit{
		ID:      "code-commit-id",
		Message: "msg",
	}, nil)

	buildDeployResp, err := c.BuildDeploy(ctx, resp.ID, &BuildDeployRequest{
		Title:       "title",
		Description: "description",
		Git: &BuildDeployRequestGit{
			Branch: "develop",
		},
	})
	assert.Nil(t, err)
	b, _ = json.Marshal(buildDeployResp)
	t.Logf("%v", string(b))

	clusterGitRepo.EXPECT().GetRestartTime(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil).AnyTimes()
	clusterGitRepo.EXPECT().MergeBranch(ctx, gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).Return("newest-commit", nil).AnyTimes()
	clusterGitRepo.EXPECT().GetRepoInfo(ctx, gomock.Any(), gomock.Any()).Return(&gitrepo.RepoInfo{
		GitRepoURL: "ssh://xxxx.git",
		ValueFiles: []string{"file1", "file2"},
	}).AnyTimes()

	cd.EXPECT().DeployCluster(ctx, gomock.Any()).Return(nil).AnyTimes()
	cd.EXPECT().GetClusterStateV1(ctx, gomock.Any()).Return(nil, herrors.NewErrNotFound(herrors.PodsInK8S, "test"))
	internalDeployResp, err := c.InternalDeploy(ctx, resp.ID, &InternalDeployRequest{
		PipelinerunID: buildDeployResp.PipelinerunID,
	})
	assert.Nil(t, err)
	b, _ = json.Marshal(internalDeployResp)
	t.Logf("%v", string(b))

	// v2
	// InternalDeployV2 needs a new context with jwt token string
	user, err := common.UserFromContext(ctx)
	assert.Nil(t, err)
	_, err = c.userManager.Create(ctx, &usermodels.User{
		Name: user.GetName(),
	})
	assert.Nil(t, err)
	token, err := c.tokenSvc.CreateJWTToken(strconv.Itoa(int(user.GetID())), time.Hour,
		tokenservice.WithPipelinerunID(buildDeployResp.PipelinerunID))
	assert.Nil(t, err)
	newCtx := common.WithContextJWTTokenString(ctx, token)

	clusterGitRepo.EXPECT().GetConfigCommit(gomock.Any(), application.Name, resp.Name).
		Return(&gitrepo.ClusterCommit{
			Master: "master",
			Gitops: "gitops",
		}, nil).AnyTimes()
	clusterGitRepo.EXPECT().CompareConfig(gomock.Any(), application.Name, resp.Name, gomock.Any(), gomock.Any()).
		Return("config-diff", nil).AnyTimes()
	clusterGitRepo.EXPECT().UpdatePipelineOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("image-commit", nil).AnyTimes()
	clusterGitRepo.EXPECT().MergeBranch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).Return("newest-commit", nil).AnyTimes()
	clusterGitRepo.EXPECT().GetRepoInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&gitrepo.RepoInfo{
		GitRepoURL: "ssh://xxxx.git",
		ValueFiles: []string{"file1", "file2"},
	}).AnyTimes()
	clusterGitRepo.EXPECT().GetEnvValue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&gitrepo.EnvValue{
		Namespace: "test-1",
	}, nil).AnyTimes()
	cd.EXPECT().CreateCluster(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	cd.EXPECT().DeployCluster(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	internalDeployRespV2, err := c.InternalDeployV2(newCtx, resp.ID, &InternalDeployRequestV2{
		PipelinerunID: buildDeployResp.PipelinerunID,
		Output:        nil,
	})
	assert.Nil(t, err)
	b, _ = json.Marshal(internalDeployRespV2)
	t.Logf("%v", string(b))

	clusterStatusResp, err := c.GetClusterStatus(ctx, resp.ID)
	assert.Nil(t, err)
	b, _ = json.Marshal(clusterStatusResp)
	t.Logf("%v", string(b))

	buildStatusResp, err := c.GetClusterPipelinerunStatus(ctx, resp.ID)
	assert.Nil(t, err)
	b, _ = json.Marshal(buildStatusResp)
	t.Logf("%v", string(b))

	codeBranch := "master"
	commitID := "code-commit-id"
	commitMsg := "code-commit-msg"
	configDiff := "config-diff"
	commitGetter.EXPECT().GetCommitHistoryLink(gomock.Any(), gomock.Any()).Return("https://cloudnative.com:22222/demo/springboot-demo/-/commits/"+codeBranch, nil).AnyTimes()
	commitGetter.EXPECT().GetCommit(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(&git.Commit{
		ID:      commitID,
		Message: commitMsg,
	}, nil)
	clusterGitRepo.EXPECT().CompareConfig(ctx, gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).Return(configDiff, nil).AnyTimes()

	getdiffResp, err := c.GetDiff(ctx, resp.ID, codemodels.GitRefTypeBranch, codeBranch)
	assert.Nil(t, err)

	link, _ := commitGetter.GetCommitHistoryLink(UpdateGitURL, codeBranch)
	assert.Equal(t, &GetDiffResponse{
		CodeInfo: &CodeInfo{
			Branch:    codeBranch,
			CommitID:  commitID,
			CommitMsg: commitMsg,
			Link:      link,
		},
		ConfigDiff: configDiff,
	}, getdiffResp)
	b, _ = json.Marshal(getdiffResp)
	t.Logf("%s", string(b))

	// test restart
	clusterGitRepo.EXPECT().UpdateRestartTime(ctx, gomock.Any(), gomock.Any(),
		gomock.Any()).Return("update-image-commit", nil)

	restartResp, err := c.Restart(ctx, resp.ID)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	b, _ = json.Marshal(restartResp)
	t.Logf("%s", string(b))
	pr, err := manager.PipelinerunMgr.GetByID(ctx, restartResp.PipelinerunID)
	assert.Nil(t, err)
	assert.Equal(t, string(prmodels.StatusOK), pr.Status)
	assert.NotNil(t, pr.FinishedAt)

	// test deploy
	clusterGitRepo.EXPECT().GetPipelineOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, herrors.ErrPipelineOutputEmpty).Times(1)
	commitGetter.EXPECT().GetCommit(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(&git.Commit{
		ID:      commitID,
		Message: commitMsg,
	}, nil).AnyTimes()
	deployResp, err := c.Deploy(ctx, resp.ID, &DeployRequest{
		Title:       "deploy-title",
		Description: "deploy-description",
	})
	assert.Equal(t, herrors.ErrShouldBuildDeployFirst, perror.Cause(err))
	assert.Nil(t, deployResp)

	clusterGitRepo.EXPECT().GetPipelineOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil,
		nil).Times(1)

	deployResp, err = c.Deploy(ctx, resp.ID, &DeployRequest{
		Title:       "deploy-title",
		Description: "deploy-description",
	})
	assert.Equal(t, herrors.ErrShouldBuildDeployFirst, perror.Cause(err))
	assert.Nil(t, deployResp)

	type PipelineOutput struct {
		Image *string
	}

	clusterGitRepo.EXPECT().GetPipelineOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&PipelineOutput{Image: &imageName}, nil).AnyTimes()

	deployResp, err = c.Deploy(ctx, resp.ID, &DeployRequest{
		Title:       "deploy-title",
		Description: "deploy-description",
	})
	assert.Nil(t, err)
	assert.NotNil(t, deployResp)

	b, _ = json.Marshal(deployResp)
	t.Logf("%s", string(b))

	pr, err = manager.PipelinerunMgr.GetByID(ctx, deployResp.PipelinerunID)
	assert.Nil(t, err)
	assert.Equal(t, string(prmodels.StatusCreated), pr.Status)

	// test next
	k8sutil.EXPECT().ExecuteAction(ctx, gomock.Any()).Return(nil)
	err = c.ExecuteAction(ctx, resp.ID, "next", schema.GroupVersionResource{})
	assert.Nil(t, err)

	// test Online and Offline
	execResp := map[string]clustercd.ExecResp{
		"pod1": {
			Result: true,
		},
		"pod2": {
			Result: false,
			Stderr: "error",
		},
	}

	k8sutil.EXPECT().Exec(ctx, gomock.Any()).Return(execResp, nil).Times(3)

	execRequest := &ExecRequest{
		PodList:  []string{"pod1", "pod2"},
		Commands: []string{"echo 'hello, world'"},
	}

	onlineResp, err := c.Online(ctx, resp.ID, execRequest)
	assert.Nil(t, err)
	assert.NotNil(t, onlineResp)
	b, _ = json.Marshal(onlineResp)
	t.Logf("%s", string(b))

	offlineResp, err := c.Offline(ctx, resp.ID, execRequest)
	assert.Nil(t, err)
	assert.NotNil(t, offlineResp)
	b, _ = json.Marshal(offlineResp)
	t.Logf("%s", string(b))

	shellResp, err := c.Exec(ctx, resp.ID, execRequest)
	assert.Nil(t, err)
	assert.NotNil(t, shellResp)
	b, _ = json.Marshal(shellResp)
	t.Logf("%s", string(b))

	valueFile := gitrepo.ClusterValueFile{
		FileName: common.GitopsFileTags,
	}
	err = yaml.Unmarshal([]byte(`javaapp:
  tags: 
    test_key: test_value`), &valueFile.Content)
	assert.Nil(t, err)

	clusterGitRepo.EXPECT().GetClusterValueFiles(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]gitrepo.ClusterValueFile{valueFile}, nil)
	// test rollback
	clusterGitRepo.EXPECT().Rollback(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
		Return("rollback-commit", nil).AnyTimes()
	clusterGitRepo.EXPECT().GetClusterTemplate(ctx, application.Name, resp.Name).
		Return(&gitrepo.ClusterTemplate{
			Name:    resp.Template.Name,
			Release: resp.Template.Release,
		}, nil).AnyTimes()
	clusterGitRepo.EXPECT().GetManifest(ctx, application.Name, resp.Name, gomock.Any()).
		Return(nil, herrors.NewErrNotFound(herrors.GitlabResource, "")).Times(2)
	// update status to 'ok'
	err = manager.PipelinerunMgr.UpdateResultByID(ctx, buildDeployResp.PipelinerunID, &prmodels.Result{
		Result: string(prmodels.StatusOK),
	})
	assert.Nil(t, err)

	c.tagMgr = manager.TagManager
	rollbackResp, err := c.Rollback(ctx, resp.ID, &RollbackRequest{
		PipelinerunID: buildDeployResp.PipelinerunID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, rollbackResp)
	b, _ = json.Marshal(rollbackResp)
	t.Logf("%s", string(b))
	pr, err = manager.PipelinerunMgr.GetByID(ctx, rollbackResp.PipelinerunID)
	assert.Nil(t, err)
	assert.Equal(t, string(prmodels.StatusOK), pr.Status)
	assert.NotNil(t, pr.FinishedAt)
	tags, err := manager.TagManager.ListByResourceTypeID(ctx, common.ResourceCluster, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "test_key", tags[0].Key)
	assert.Equal(t, "test_value", tags[0].Value)
	c.tagMgr = tagManager

	k8sutil.EXPECT().DeletePods(ctx, gomock.Any()).Return(
		map[string]clustercd.OperationResult{
			"pod1": {Result: true},
		}, nil)
	result, err := c.DeleteClusterPods(ctx, resp.ID, []string{"pod1"})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	value, ok := result["pod1"]
	assert.Equal(t, true, ok)
	assert.Equal(t, true, value.Result)

	podExist := "exist"
	podNotExist := "notexist"
	k8sutil.EXPECT().GetPod(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, param *clustercd.GetPodParams) (*v1.Pod, error) {
			if param.Pod == podExist {
				return &v1.Pod{}, nil
			} else {
				return nil, herrors.NewErrNotFound(herrors.PodsInK8S, "")
			}
		},
	).Times(2)
	_, err = c.GetClusterPod(ctx, resp.ID, podExist)
	assert.Nil(t, err)

	_, err = c.GetClusterPod(ctx, resp.ID, podNotExist)
	assert.NotNil(t, err)
	_, ok = perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.Equal(t, true, ok)

	patchJSONStr := `{
		"app": {
		  "params": {
			"xmx": "1024",
			"jvmExtra": "-Dserver.port=8181"
		  }
		}
	  }`

	mergedJSONStr := `{
		"app": {
		  "spec": {
			"replicas": 1,
			"resource": "small"
		  },
		  "strategy": {
			"stepsTotal": 1,
			"pauseType": "first"
		  },
		  "params": {
			"xmx": "1024",
			"xms": "512",
			"maxPerm": "128",
			"mainClassName": "com.netease.horizon.WebApplication",
			"jvmExtra": "-Dserver.port=8181"
		  },
		  "health": {
			"check": "/api/test",
			"status": "/health/status",
			"online": "/health/online",
			"offline": "/health/offline",
			"port": 8080
		  }
		}
	  }`

	patchJsonBlob := map[string]interface{}{}
	err = json.Unmarshal([]byte(patchJSONStr), &patchJsonBlob)
	assert.Nil(t, err)
	createClusterRequest = &CreateClusterRequest{
		Base: &Base{
			Description: "cluster description",
			Git: &codemodels.Git{
				Branch: "develop",
			},
			TemplateInput: &TemplateInput{
				Application: patchJsonBlob,
				Pipeline:    pipelineJSONBlob,
			},
		},
		Name: "app-cluster-mergepatch",
	}

	clusterGitRepo.EXPECT().CreateCluster(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, params *gitrepo.CreateClusterParams) error {
			blob := map[string]interface{}{}
			err := json.Unmarshal([]byte(mergedJSONStr), &blob)
			assert.Nil(t, err)
			assertMapEqual(t, blob, params.ApplicationJSONBlob)
			return nil
		},
	).Times(1)
	resp, err = c.CreateCluster(ctx, application.ID, "test", "hz", createClusterRequest, true)
	assert.Nil(t, err)

	patchJSONStr = `{
		"app": {
		  "params": {
			"xmx": "2048",
			"jvmExtra": "-Dserver.port=8282"
		  }
		}
	  }`
	mergedJSONStr = `{
		"app": {
		  "spec": {
			"replicas": 1,
			"resource": "small"
		  },
		  "strategy": {
			"stepsTotal": 1,
			"pauseType": "first"
		  },
		  "params": {
			"xmx": "2048",
			"xms": "512",
			"maxPerm": "128",
			"mainClassName": "com.netease.horizon.WebApplication",
			"jvmExtra": "-Dserver.port=8282"
		  },
		  "health": {
			"check": "/api/test",
			"status": "/health/status",
			"online": "/health/online",
			"offline": "/health/offline",
			"port": 8080
		  }
		}
	  }`

	patchJsonBlob = map[string]interface{}{}
	err = json.Unmarshal([]byte(patchJSONStr), &patchJsonBlob)
	assert.Nil(t, err)
	updateClusterRequest = &UpdateClusterRequest{
		Base: &Base{
			Description: "new description",
			Git: &codemodels.Git{
				URL:       UpdateGitURL,
				Subfolder: "/new",
				Branch:    "new",
			},
			TemplateInput: &TemplateInput{
				Application: patchJsonBlob,
				Pipeline:    pipelineJSONBlob,
			},
			Template: &Template{
				Name:    "tomcat7_jdk8",
				Release: "v1.0.1",
			},
		},
	}
	clusterGitRepo.EXPECT().UpdateCluster(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, params *gitrepo.UpdateClusterParams) error {
			blob := map[string]interface{}{}
			err := json.Unmarshal([]byte(mergedJSONStr), &blob)
			assert.Nil(t, err)
			assertMapEqual(t, blob, params.ApplicationJSONBlob)
			return nil
		},
	).Times(1)
	_, err = c.UpdateCluster(ctx, resp.ID, updateClusterRequest, true)
	assert.Nil(t, err)
}

func testV2(t *testing.T) {
	// for test
	conf := config.Config{}
	param := param.Param{
		AutoFreeSvc: service.New([]string{"dev", "test"}),
		Manager:     managerparam.InitManager(nil),
	}
	NewController(&conf, &param)
	templateName := "rollout"
	templateVersion := "v1.0.0"
	mockCtl := gomock.NewController(t)
	clusterGitRepo := clustergitrepomock.NewMockClusterGitRepo(mockCtl)
	applicationGitRepo := applicationgitrepomock.NewMockApplicationGitRepo2(mockCtl)
	templateSchemaGetter := trschemamock.NewMockGetter(mockCtl)
	registryFty := registryftymock.NewMockRegistryGetter(mockCtl)
	mockCd := cdmock.NewMockCD(mockCtl)

	templateSchemaGetter.EXPECT().GetTemplateSchema(gomock.Any(), templateName, "v1.0.0", gomock.Any()).
		Return(&trschema.Schemas{
			Application: &trschema.Schema{
				JSONSchema: applicationSchema,
			},
			Pipeline: &trschema.Schema{
				JSONSchema: pipelineSchema,
			},
		}, nil).Times(1)

	appMgr := manager.ApplicationManager
	trMgr := manager.TemplateReleaseManager
	envMgr := manager.EnvMgr
	regionMgr := manager.RegionMgr
	groupMgr := manager.GroupManager
	envRegionMgr := manager.EnvRegionMgr

	registryDAO := registrydao.NewDAO(db)
	id, err := registryDAO.Create(ctx, &registrymodels.Registry{
		Server: "https://harbor.com",
		Token:  "xxx",
	})
	assert.Nil(t, err)
	assert.NotNil(t, id)

	region, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
		RegistryID:  id,
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	er, err := envRegionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: "test2",
		RegionName:      "hz",
	})
	assert.Nil(t, err)
	assert.NotNil(t, er)
	er, err = envRegionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: "dev2",
		RegionName:      "hz",
	})
	assert.Nil(t, err)
	assert.NotNil(t, er)

	env, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "dev2",
		DisplayName: "开发",
	})
	assert.Nil(t, err)
	assert.NotNil(t, env)
	env, err = envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "test2",
		DisplayName: "开发",
	})
	assert.Nil(t, err)
	assert.NotNil(t, env)

	// init data
	group, err := groupMgr.Create(ctx, &groupmodels.Group{
		Name:     "group1",
		Path:     "group1",
		ParentID: 0,
	})
	t.Logf("%+v", err)
	t.Logf("%+v", group)
	assert.Nil(t, err)
	assert.NotNil(t, group)
	gitURL := "ssh://git.com"

	applicationName := "app2"
	appGitSubFolder := "/test"
	appGitRef := "master"
	priority := "P3"
	application, err := appMgr.Create(ctx, &appmodels.Application{
		GroupID:         group.ID,
		Name:            applicationName,
		Priority:        appmodels.Priority(priority),
		GitURL:          gitURL,
		GitSubfolder:    appGitSubFolder,
		GitRef:          appGitRef,
		Template:        templateName,
		TemplateRelease: templateVersion,
	}, nil)
	assert.Nil(t, err)

	tr, err := trMgr.Create(ctx, &trmodels.TemplateRelease{
		TemplateName: templateName,
		Name:         "v1.0.0",
		ChartVersion: "v1.0.0-test",
		ChartName:    templateName,
	})
	assert.Nil(t, err)
	assert.NotNil(t, tr)

	c = &controller{
		clusterMgr:           manager.ClusterMgr,
		clusterGitRepo:       clusterGitRepo,
		applicationMgr:       appMgr,
		templateReleaseMgr:   trMgr,
		templateSchemaGetter: templateSchemaGetter,
		envMgr:               envMgr,
		envRegionMgr:         envRegionMgr,
		regionMgr:            regionMgr,
		groupSvc:             groupservice.NewService(manager),
		pipelinerunMgr:       manager.PipelinerunMgr,
		userManager:          manager.UserManager,
		autoFreeSvc:          service.New([]string{"dev2", "test2"}),
		userSvc:              userservice.NewService(manager),
		schemaTagManager:     manager.ClusterSchemaTagMgr,
		applicationGitRepo:   applicationGitRepo,
		tagMgr:               manager.TagManager,
		registryFty:          registryFty,
		cd:                   mockCd,
		eventMgr:             manager.EventManager,
	}
	applicationGitRepo.EXPECT().GetApplication(gomock.Any(), applicationName, gomock.Any()).
		Return(&appgitrepo.GetResponse{
			Manifest:     nil,
			BuildConf:    pipelineJSONBlob,
			TemplateConf: applicationJSONBlob,
		}, nil).Times(1)
	clusterGitRepo.EXPECT().CreateCluster(ctx, gomock.Any()).Return(nil).Times(1)

	createClusterName := "app-cluster2"
	createReq := &CreateClusterRequestV2{
		Name:        createClusterName,
		Description: "cluster description",
		ExpireTime:  "24h0m0s",
		Git: &codemodels.Git{
			Branch: "develop",
		},
		Tags: tagmodels.TagsBasic{
			{
				Key:   "key1",
				Value: "value1",
			},
		},
		BuildConfig: pipelineJSONBlob,
		TemplateInfo: &codemodels.TemplateInfo{
			Name:    templateName,
			Release: templateVersion,
		},
		TemplateConfig: applicationJSONBlob,
		ExtraMembers:   nil,
	}
	resp, err := c.CreateClusterV2(ctx, &CreateClusterParamsV2{
		CreateClusterRequestV2: createReq,
		ApplicationID:          application.ID,
		Environment:            "test2",
		Region:                 "hz",
		MergePatch:             false,
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.ApplicationID, application.ID)
	assert.Equal(t, resp.FullPath, "/"+group.Path+"/"+application.Name+"/"+createClusterName)
	t.Logf("%+v", resp)

	// then get cluster
	clusterGitRepo.EXPECT().GetCluster(ctx, applicationName, createClusterName, templateName).Return(&gitrepo.ClusterFiles{
		PipelineJSONBlob:    pipelineJSONBlob,
		ApplicationJSONBlob: applicationJSONBlob,
		Manifest:            nil,
	}, nil).Times(1)

	getClusterResp, err := c.GetClusterV2(ctx, resp.ID)
	assert.Nil(t, err)
	assert.Equal(t, getClusterResp.ID, resp.ID)
	assert.Equal(t, getClusterResp.Name, createClusterName)
	assert.Equal(t, getClusterResp.Priority, priority)
	assert.Equal(t, createReq.ExpireTime, getClusterResp.ExpireTime)
	assert.NotNil(t, getClusterResp.Scope)
	assert.Equal(t, getClusterResp.FullPath, "/"+group.Path+"/"+application.Name+"/"+createClusterName)
	assert.Equal(t, getClusterResp.ApplicationName, application.Name)
	assert.Equal(t, getClusterResp.ApplicationID, application.ID)
	assertMapEqual(t, getClusterResp.TemplateConfig, applicationJSONBlob)
	assertMapEqual(t, getClusterResp.BuildConfig, pipelineJSONBlob)
	assert.Equal(t, getClusterResp.TemplateInfo.Name, templateName)
	assert.Equal(t, getClusterResp.TemplateInfo.Release, templateVersion)
	assert.Nil(t, getClusterResp.Manifest)
	assert.Equal(t, getClusterResp.Status, "")
	assert.Equal(t, 1, len(getClusterResp.Tags))
	assert.Equal(t, getClusterResp.Tags[0].Value, "value1")
	t.Logf("%+v", getClusterResp)

	// update v2
	clusterGitRepo.EXPECT().GetCluster(ctx, applicationName, createClusterName, templateName).Return(&gitrepo.ClusterFiles{
		PipelineJSONBlob:    pipelineJSONBlob,
		ApplicationJSONBlob: applicationJSONBlob,
		Manifest:            nil,
	}, nil).Times(1)

	updateRequestV2 := &UpdateClusterRequestV2{
		Tags: tagmodels.TagsBasic{
			{
				Key:   "key1",
				Value: "value2",
			},
		},
		Description:    "",
		ExpireTime:     "336h0m0s",
		BuildConfig:    pipelineJSONBlob,
		TemplateInfo:   nil,
		TemplateConfig: applicationJSONBlob,
	}
	// update manifest not exist (not v2 repo)
	err = c.UpdateClusterV2(ctx, getClusterResp.ID, updateRequestV2, false)
	t.Logf("%+v", err)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "parameter is invalid"))

	var manifest = make(map[string]interface{})
	manifest["Version"] = common.MetaVersion2

	clusterGitRepo.EXPECT().GetCluster(ctx, applicationName, createClusterName, templateName).Return(&gitrepo.ClusterFiles{
		PipelineJSONBlob:    pipelineJSONBlob,
		ApplicationJSONBlob: applicationJSONBlob,
		Manifest:            manifest,
	}, nil).Times(1)
	clusterGitRepo.EXPECT().UpdateCluster(ctx, gomock.Any()).Return(nil).Times(1)
	clusterGitRepo.EXPECT().UpdateTags(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	templateSchemaGetter.EXPECT().GetTemplateSchema(gomock.Any(), templateName, "v1.0.0", gomock.Any()).
		Return(&trschema.Schemas{
			Application: &trschema.Schema{
				JSONSchema: applicationSchema,
			},
			Pipeline: &trschema.Schema{
				JSONSchema: pipelineSchema,
			},
		}, nil).Times(1)
	err = c.UpdateClusterV2(ctx, getClusterResp.ID, updateRequestV2, false)
	t.Logf("%+v", err)
	assert.Nil(t, err)

	registry := registrymock.NewMockRegistry(mockCtl)
	registryFty.EXPECT().GetRegistryByConfig(gomock.Any(), gomock.Any()).Return(registry, nil).Times(1)
	registry.EXPECT().DeleteImage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	clusterGitRepo.EXPECT().DeleteCluster(gomock.Any(), applicationName,
		createClusterName, getClusterResp.ID).Return(nil).Times(1)
	mockCd.EXPECT().DeleteCluster(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	err = c.DeleteCluster(ctx, getClusterResp.ID, false)
	assert.Nil(t, err)
	time.Sleep(time.Second * 5)
}

func testUpgrade(t *testing.T) {
	// for test
	conf := config.Config{}
	parameter := param.Param{
		AutoFreeSvc: service.New([]string{"dev", "test"}),
		Manager:     managerparam.InitManager(nil),
	}
	NewController(&conf, &parameter)
	templateName := "javaapp"
	templateRelease := "v1.0.1"
	mockCtl := gomock.NewController(t)
	clusterGitRepo := clustergitrepomock.NewMockClusterGitRepo(mockCtl)
	applicationGitRepo := applicationgitrepomock.NewMockApplicationGitRepo2(mockCtl)
	templateSchemaGetter := trschemamock.NewMockGetter(mockCtl)
	registryFty := registryftymock.NewMockRegistryGetter(mockCtl)
	mockCd := cdmock.NewMockCD(mockCtl)

	appMgr := manager.ApplicationManager
	trMgr := manager.TemplateReleaseManager
	envMgr := manager.EnvMgr
	regionMgr := manager.RegionMgr
	groupMgr := manager.GroupManager
	envRegionMgr := manager.EnvRegionMgr

	// init data
	group, err := groupMgr.Create(ctx, &groupmodels.Group{
		Name:     "group-upgrade",
		Path:     "group-upgrade",
		ParentID: 0,
	})
	t.Logf("%+v", err)
	t.Logf("%+v", group)
	assert.Nil(t, err)
	assert.NotNil(t, group)
	gitURL := "ssh://git.com"

	applicationName := "app-upgrade"
	appGitSubFolder := "/test"
	appGitRef := "master"
	priority := "P3"
	application, err := appMgr.Create(ctx, &appmodels.Application{
		GroupID:         group.ID,
		Name:            applicationName,
		Priority:        appmodels.Priority(priority),
		GitURL:          gitURL,
		GitSubfolder:    appGitSubFolder,
		GitRef:          appGitRef,
		Template:        templateName,
		TemplateRelease: templateRelease,
	}, nil)
	assert.Nil(t, err)

	tr, err := trMgr.Create(ctx, &trmodels.TemplateRelease{
		TemplateName: templateName,
		Name:         templateRelease,
		ChartVersion: templateRelease + "-test",
		ChartName:    templateName,
	})
	assert.Nil(t, err)
	assert.NotNil(t, tr)

	templateUpgradeMapper := templateconfig.UpgradeMapper{
		"javaapp": {
			Name:    "rollout",
			Release: "v1.0.0",
			BuildConfig: templateconfig.BuildConfig{
				Language:    "java",
				Environment: "javaapp",
			},
		},
	}

	c = &controller{
		clusterMgr:            manager.ClusterMgr,
		clusterGitRepo:        clusterGitRepo,
		applicationMgr:        appMgr,
		templateReleaseMgr:    trMgr,
		templateSchemaGetter:  templateSchemaGetter,
		envMgr:                envMgr,
		envRegionMgr:          envRegionMgr,
		regionMgr:             regionMgr,
		groupSvc:              groupservice.NewService(manager),
		pipelinerunMgr:        manager.PipelinerunMgr,
		userManager:           manager.UserManager,
		autoFreeSvc:           parameter.AutoFreeSvc,
		userSvc:               userservice.NewService(manager),
		schemaTagManager:      manager.ClusterSchemaTagMgr,
		applicationGitRepo:    applicationGitRepo,
		tagMgr:                manager.TagManager,
		registryFty:           registryFty,
		cd:                    mockCd,
		eventMgr:              manager.EventManager,
		templateUpgradeMapper: templateUpgradeMapper,
	}

	applicationGitRepo.EXPECT().GetApplication(ctx, gomock.Any(), gomock.Any()).
		Return(&appgitrepo.GetResponse{
			Manifest:     nil,
			BuildConf:    pipelineJSONBlob,
			TemplateConf: applicationJSONBlob,
		}, nil).AnyTimes()
	templateSchemaGetter.EXPECT().GetTemplateSchema(gomock.Any(), templateName, templateRelease, gomock.Any()).
		Return(&trschema.Schemas{
			Application: &trschema.Schema{
				JSONSchema: applicationSchema,
			},
			Pipeline: &trschema.Schema{
				JSONSchema: pipelineSchema,
			},
		}, nil).Times(1)
	clusterGitRepo.EXPECT().CreateCluster(ctx, gomock.Any()).Return(nil).Times(1)
	clusterGitRepo.EXPECT().GetEnvValue(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(&gitrepo.EnvValue{
		Namespace: "test-1",
	}, nil).AnyTimes()

	createClusterName := "app-cluster-upgrade"
	createClusterRequest := &CreateClusterRequest{
		Base: &Base{
			Description: "cluster-upgrade description",
			Git: &codemodels.Git{
				Branch: "develop",
			},
			TemplateInput: &TemplateInput{
				Application: applicationJSONBlob,
				Pipeline:    pipelineJSONBlob,
			},
			Template: &Template{
				Name:    templateName,
				Release: templateRelease,
			},
		},
		Name: createClusterName,
	}

	resp, err := c.CreateCluster(ctx, application.ID, "test", "hz", createClusterRequest, false)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.Application.ID, application.ID)
	assert.Equal(t, resp.FullPath, "/"+group.Path+"/"+application.Name+"/"+createClusterName)

	clusterGitRepo.EXPECT().GetClusterTemplate(ctx, application.Name, resp.Name).
		Return(&gitrepo.ClusterTemplate{
			Name:    resp.Template.Name,
			Release: resp.Template.Release,
		}, nil).AnyTimes()
	clusterGitRepo.EXPECT().UpgradeCluster(ctx, gomock.Any()).Return("", nil).Times(1)
	clusterGitRepo.EXPECT().DefaultBranch().Return("master").AnyTimes()
	clusterGitRepo.EXPECT().CompareConfig(ctx, gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).Return("", nil).Times(1)

	err = c.Upgrade(ctx, resp.ID)
	assert.Nil(t, err)

	registry := registrymock.NewMockRegistry(mockCtl)
	registryFty.EXPECT().GetRegistryByConfig(gomock.Any(), gomock.Any()).Return(registry, nil).Times(1)
	registry.EXPECT().DeleteImage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	clusterGitRepo.EXPECT().DeleteCluster(gomock.Any(), applicationName,
		createClusterName, resp.ID).Return(nil).Times(1)
	mockCd.EXPECT().DeleteCluster(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	err = c.DeleteCluster(ctx, resp.ID, false)
	assert.Nil(t, err)
	time.Sleep(time.Second * 5)
}

func assertMapEqual(t *testing.T, expected, got map[string]interface{}) {
	expectedBuf, err := json.Marshal(expected)
	if err != nil {
		t.Error(err)
		return
	}
	gotBuf, err := json.Marshal(got)
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(expectedBuf, gotBuf) {
		t.Errorf("expected %s,\n got %s", string(expectedBuf), string(gotBuf))
		return
	}
}

func testGetClusterOutPut(t *testing.T) {
	mockCtl := gomock.NewController(t)
	appManagerMock := applicationmanangermock.NewMockManager(mockCtl)
	clusterManagerMock := clustermanagermock.NewMockManager(mockCtl)
	outputMock := outputmock.NewMockGetter(mockCtl)
	clusterGitRepoMock := clustergitrepomock.NewMockClusterGitRepo(mockCtl)
	c := controller{
		clusterMgr:     clusterManagerMock,
		applicationMgr: appManagerMock,
		outputGetter:   outputMock,
		clusterGitRepo: clusterGitRepoMock,
	}

	var applicationID uint = 102
	template := "javaapp"
	templateRelease := "v1.0.0"
	clusterName := "app-cluster-demo"
	clusterManagerMock.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.Cluster{
		Model:           global.Model{},
		ApplicationID:   applicationID,
		Template:        template,
		TemplateRelease: templateRelease,
		Name:            clusterName,
	}, nil).Times(1)

	applicationName := "app-demo"
	appManagerMock.EXPECT().GetByID(gomock.Any(), applicationID).Return(&appmodels.Application{
		Model:   global.Model{},
		GroupID: 0,
		Name:    applicationName,
	}, nil).Times(1)

	envValueFile := gitrepo.ClusterValueFile{
		FileName: "env.yaml",
	}
	var envValue = `
javaapp:
  env:
    environment: pre
    region: hz
    namespace: pre-54
    baseRegistry: harbor.cloudnative.com
    ingressDomain: cloudnative.com
  horizon:
    cluster: app-cluster-demo
`
	err := yaml.Unmarshal([]byte(envValue), &(envValueFile.Content))
	assert.Nil(t, err)
	var clusterValueFiles = make([]gitrepo.ClusterValueFile, 0)
	clusterValueFiles = append(clusterValueFiles, envValueFile)
	clusterGitRepoMock.EXPECT().GetClusterValueFiles(gomock.Any(), applicationName, clusterName).Return(
		clusterValueFiles, nil).Times(1)

	var outPutStr = `
syncDomainName:
  Description: sync domain name
  Value: {{ .Values.horizon.cluster}}.{{ .Values.env.ingressDomain}}`
	outputMock.EXPECT().GetTemplateOutPut(gomock.Any(), template, templateRelease).Return(outPutStr, nil).Times(1)

	renderObject, err := c.GetClusterOutput(context.TODO(), 123)
	assert.Nil(t, err)
	builder := &strings.Builder{}
	encoder := yaml.NewEncoder(builder)
	encoder.SetIndent(2)
	err = encoder.Encode(renderObject)
	assert.Nil(t, err)
	var ExpectOutPutStr = `syncDomainName:
  Description: sync domain name
  Value: app-cluster-demo.cloudnative.com
`
	assert.Equal(t, ExpectOutPutStr, builder.String())
}

var envValue = `
javaapp:
  env:
    environment: pre
    region: hz
    namespace: pre-54
    baseRegistry: harbor.cloudnative.com
    ingressDomain: cloudnative.com

`
var horizonValue = `
javaapp:
  horizon:
    application: music-social-zone
    cluster: music-social-zone-pre
    template:
      name: javaapp
      release: v1.0.0
    priority: P2
`
var applicationValue = `
javaapp:
  app:
    health:
      check: /api/test
      offline: /health/offline
      online: /health/active
      port: 8888
      status: /health/status
    params:
      jvmExtra: -Dserver.port=8888
      mainClassName: com.netease.music.social.zone.WebApplication
      xdebugAddress: "10000"
      xms: "2048"
      xmx: "2048"
    spec:
      resource: large
`

func testRenderOutPutObject(t *testing.T) {
	var envValueFile, horizonValueFile, applicationValueFile gitrepo.ClusterValueFile
	err := yaml.Unmarshal([]byte(envValue), &(envValueFile.Content))
	assert.Nil(t, err)

	err = yaml.Unmarshal([]byte(horizonValue), &(horizonValueFile.Content))
	assert.Nil(t, err)

	err = yaml.Unmarshal([]byte(applicationValue), &(applicationValueFile.Content))
	assert.Nil(t, err)

	var outPutStr = `syncDomainName:
  Description: sync domain name
  Value: {{ .Values.horizon.cluster}}.{{ .Values.env.ingressDomain}}`
	outPutRenderJSONObject, err := RenderOutputObject(outPutStr, "javaapp",
		horizonValueFile, envValueFile, applicationValueFile)

	assert.Nil(t, err)
	t.Logf("outPutRenderStr = \n%+v", outPutRenderJSONObject)

	jsonBytes, err := json.Marshal(outPutRenderJSONObject)
	assert.Nil(t, err)
	t.Logf("outPutRenderStr = \n%+s", string(jsonBytes))
	var expectOutPutStr = `{"syncDomainName":{"Description":"sync domain name","Value":"music-social-zone-pre.cloudnative.com"}}` // nolint
	assert.Equal(t, expectOutPutStr, string(jsonBytes))
}

func testRenderOutPutObjectMissingKey(t *testing.T) {
	var envValueFile, horizonValueFile, applicationValueFile gitrepo.ClusterValueFile
	var envValue = `
javaapp:
  env:
    environment: pre
    region: hz
    namespace: pre-54
    baseRegistry: harbor.cloudnative.com
`
	var horizonValue = `
javaapp:
  horizon:
    application: music-social-zone
    template:
      name: javaapp
      release: v1.0.0
    priority: P2
`
	err := yaml.Unmarshal([]byte(envValue), &(envValueFile.Content))
	assert.Nil(t, err)

	err = yaml.Unmarshal([]byte(horizonValue), &(horizonValueFile.Content))
	assert.Nil(t, err)

	err = yaml.Unmarshal([]byte(applicationValue), &(applicationValueFile.Content))
	assert.Nil(t, err)
	var outPutStr = `syncDomainName:
  Description: sync domain name
  Value: {{ .Values.horizon.cluster}}.{{ .Values.env.ingressDomain}}`
	outPutRenderJSONObject, err := RenderOutputObject(outPutStr, "javaapp",
		horizonValueFile, envValueFile)

	assert.Nil(t, err)
	t.Logf("outPutRenderStr = \n%+v", outPutRenderJSONObject)

	jsonBytes, err := json.Marshal(outPutRenderJSONObject)
	assert.Nil(t, err)
	t.Logf("outPutRenderStr = \n%+s", string(jsonBytes))
	var expectOutPutStr = `{"syncDomainName":{"Description":"sync domain name","Value":"."}}` // nolint
	assert.Equal(t, expectOutPutStr, string(jsonBytes))
}

func testIsClusterActuallyHealthy(t *testing.T) {
	layout := "2006-01-02 15:04:05"
	var t0 time.Time
	t1, err := time.Parse(layout, "2022-09-17 17:50:00")
	assert.Nil(t, err)
	t2, err := time.Parse(layout, "2022-09-15 17:50:00")
	assert.Nil(t, err)
	tActual, err := time.Parse(layout, "2022-09-16 17:50:00")
	assert.Nil(t, err)
	imageV1 := "v1"
	imageV2 := "v2"
	cs := &clustercd.ClusterState{}
	assert.Equal(t, false, isClusterActuallyHealthy(ctx, cs, imageV1, t0, 0))

	containerV1 := &clustercd.Container{
		Image: imageV1,
	}
	containerV2 := &clustercd.Container{
		Image: imageV2,
	}
	Pod1 := &clustercd.ClusterPod{}
	Pod2 := &clustercd.ClusterPod{}
	Pod3 := &clustercd.ClusterPod{}

	// pod1: t1, imagev1, imagev2
	Pod1.Metadata.Annotations = map[string]string{
		common.ClusterRestartTimeKey: t1.Format(layout),
	}
	Pod1.Spec.Containers = []*clustercd.Container{containerV1, containerV2}

	// pod2: t2, imagev1, imagev2
	Pod2.Metadata.Annotations = map[string]string{
		common.ClusterRestartTimeKey: t2.Format(layout),
	}
	Pod2.Spec.InitContainers = []*clustercd.Container{containerV1, containerV2}

	// pod2: imagev2
	Pod3.Spec.InitContainers = []*clustercd.Container{containerV2}

	cs.PodTemplateHash = "test"
	cs.Versions = map[string]*clustercd.ClusterVersion{}

	// none replicas is expected
	cs.Versions[cs.PodTemplateHash] = &clustercd.ClusterVersion{
		Pods: map[string]*clustercd.ClusterPod{"Pod3": Pod3},
	}
	assert.Equal(t, true, isClusterActuallyHealthy(ctx, cs, imageV1, tActual, 0))

	// one imagev1 pod is expected
	cs.Versions[cs.PodTemplateHash].Pods["Pod1"] = Pod1
	assert.Equal(t, true, isClusterActuallyHealthy(ctx, cs, imageV1, tActual, 1))

	// two imagev1 pods is expected
	cs.Versions[cs.PodTemplateHash].Pods["Pod1-copy"] = Pod1
	assert.Equal(t, true, isClusterActuallyHealthy(ctx, cs, imageV1, tActual, 2))

	// t2 pod is unexpected
	cs.Versions[cs.PodTemplateHash].Pods["Pod2"] = Pod2
	assert.Equal(t, false, isClusterActuallyHealthy(ctx, cs, imageV1, tActual, 2))

	// three t1 pods is not expected
	assert.Equal(t, false, isClusterActuallyHealthy(ctx, cs, imageV1, tActual, 3))
}
