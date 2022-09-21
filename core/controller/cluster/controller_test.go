package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/config"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	applicationgitrepomock "g.hz.netease.com/horizon/mock/pkg/application/gitrepo"
	applicationmanangermock "g.hz.netease.com/horizon/mock/pkg/application/manager"
	cdmock "g.hz.netease.com/horizon/mock/pkg/cluster/cd"
	commitmock "g.hz.netease.com/horizon/mock/pkg/cluster/code"
	clustergitrepomock "g.hz.netease.com/horizon/mock/pkg/cluster/gitrepo"
	clustermanagermock "g.hz.netease.com/horizon/mock/pkg/cluster/manager"
	registrymock "g.hz.netease.com/horizon/mock/pkg/cluster/registry"
	registryftymock "g.hz.netease.com/horizon/mock/pkg/cluster/registry/factory"
	tektonmock "g.hz.netease.com/horizon/mock/pkg/cluster/tekton"
	tektonftymock "g.hz.netease.com/horizon/mock/pkg/cluster/tekton/factory"
	tagmock "g.hz.netease.com/horizon/mock/pkg/tag/manager"
	outputmock "g.hz.netease.com/horizon/mock/pkg/templaterelease/output"
	trschemamock "g.hz.netease.com/horizon/mock/pkg/templaterelease/schema"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	clustercd "g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/code"
	codemodels "g.hz.netease.com/horizon/pkg/cluster/code"
	clustercommon "g.hz.netease.com/horizon/pkg/cluster/common"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	perror "g.hz.netease.com/horizon/pkg/errors"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	groupservice "g.hz.netease.com/horizon/pkg/group/service"
	harbordao "g.hz.netease.com/horizon/pkg/harbor/dao"
	harbormodels "g.hz.netease.com/horizon/pkg/harbor/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"g.hz.netease.com/horizon/pkg/server/global"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	tmodel "g.hz.netease.com/horizon/pkg/tag/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	trschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	gitlabschema "g.hz.netease.com/horizon/pkg/templaterelease/schema/gitlab"
	tagmodel "g.hz.netease.com/horizon/pkg/templateschematag/models"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	userservice "g.hz.netease.com/horizon/pkg/user/service"
	v1 "k8s.io/api/core/v1"

	"github.com/go-yaml/yaml"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

// nolint
var (
	ctx                                   context.Context
	c                                     Controller
	pr                                    *v1beta1.PipelineRun
	applicationSchema, pipelineSchema     map[string]interface{}
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}

	applicationSchemaJSON = `{
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
      "default": "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE project [<!ENTITY buildfile SYSTEM \"file:./build-user.xml\">]>\n<project basedir=\".\" default=\"deploy\" name=\"demo\">\n    <property name=\"ant\" value=\"ant\" />\n    <property name=\"baseline.dir\" value=\"${basedir}\"/>\n\n    <target name=\"package\">\n        <exec dir=\"${baseline.dir}\" executable=\"${ant}\" failonerror=\"true\">\n            <arg line=\"-buildfile overmind_build.xml -Denv=test -DenvName=mockserver.org\"/>\n        </exec>\n    </target>\n\n    <target name=\"deploy\">\n        <echo message=\"begin auto deploy......\"/>\n        <antcall target=\"package\"/>\n    </target>\n</project>"
    }
  }
}
`

	pipelineJSONStr = `{
		"buildxml":"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE project [<!ENTITY buildfile SYSTEM \"file:./build-user.xml\">]>\n<project basedir=\".\" default=\"deploy\" name=\"demo\">\n    <property name=\"ant\" value=\"ant\" />\n    <property name=\"baseline.dir\" value=\"${basedir}\"/>\n\n    <target name=\"package\">\n        <exec dir=\"${baseline.dir}\" executable=\"${ant}\" failonerror=\"true\">\n            <arg line=\"-buildfile overmind_build.xml -Denv=test -DenvName=mockserver.org\"/>\n        </exec>\n    </target>\n\n    <target name=\"deploy\">\n        <echo message=\"begin auto deploy......\"/>\n        <antcall target=\"package\"/>\n    </target>\n</project>"
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
                "cloudnative.music.netease.com/application":"testapp-1",
                "cloudnative.music.netease.com/cluster":"testcluster-1",
                "cloudnative.music.netease.com/environment":"env",
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
                                "imageID":"docker-pullable://harbor.mock.org/cloudnative/library/tekton-builder@sha256:14194e518981f5d893b85e170a28ba8aa80c2c610f63cfba814b6a460f48dc29"
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
                                "imageID":"docker-pullable://harbor.mock.org/cloudnative/library/tekton-builder@sha256:14194e518981f5d893b85e170a28ba8aa80c2c610f63cfba814b6a460f48dc29"
                            },
                            {
                                "terminated":{
                                    "exitCode":0,
                                    "reason":"Completed",
                                    "message":"[{\"key\":\"properties\",\"value\":\"harbor.mock.org/ndp-gjq/test-music-docker:helloworld-b1f57848-20210624143634 git@github.com:demo/demo.git helloworld b1f578488e3123e97ec00b671db143fb8f0abecf\",\"type\":\"TaskRunResult\"}]",
                                    "startedAt":"2021-06-24T06:36:34Z",
                                    "finishedAt":"2021-06-24T06:36:42Z",
                                    "containerID":"docker://9189624ad3981fd738ec5bf286f1fc5b688d71128b9827820ebc2541b2801dae"
                                },
                                "name":"image",
                                "container":"step-image",
                                "imageID":"docker-pullable://harbor.mock.org/cloudnative/library/kaniko-executor@sha256:473d6dfb011c69f32192e668d86a47c0235791e7e857c870ad70c5e86ec07e8c"
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
                                "imageID":"docker-pullable://harbor.mock.org/cloudnative/library/tekton-builder@sha256:14194e518981f5d893b85e170a28ba8aa80c2c610f63cfba814b6a460f48dc29"
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
		&harbormodels.Harbor{},
		&regionmodels.Region{}, &envregionmodels.EnvironmentRegion{},
		&prmodels.Pipelinerun{}, &tagmodel.ClusterTemplateSchemaTag{}, &tmodel.Tag{}, &envmodels.Environment{}); err != nil {
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

	os.Exit(m.Run())
}

func TestAll(t *testing.T) {
	t.Run("Test", test)
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
}

// nolint
func test(t *testing.T) {
	// for test
	conf := config.Config{}
	param := param.Param{
		Manager: managerparam.InitManager(nil),
	}
	NewController(&conf, &param)

	templateName := "javaapp"
	mockCtl := gomock.NewController(t)
	clusterGitRepo := clustergitrepomock.NewMockClusterGitRepo(mockCtl)
	applicationGitRepo := applicationgitrepomock.NewMockApplicationGitRepo(mockCtl)
	cd := cdmock.NewMockCD(mockCtl)
	tektonFty := tektonftymock.NewMockFactory(mockCtl)
	registryFty := registryftymock.NewMockFactory(mockCtl)
	commitGetter := commitmock.NewMockGitGetter(mockCtl)
	tagManager := tagmock.NewMockManager(mockCtl)

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
	harborDAO := harbordao.NewDAO(db)
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

	id, err := harborDAO.Create(ctx, &harbormodels.Harbor{
		Server:          "https://harbor.com",
		Token:           "xxx",
		PreheatPolicyID: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, id)

	region, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
		HarborID:    id,
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
		AutoFree:    true,
	})
	env, err = envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "test",
		DisplayName: "开发",
		AutoFree:    true,
	})
	assert.Nil(t, err)
	assert.NotNil(t, env)

	c = &controller{
		clusterMgr:           manager.ClusterMgr,
		clusterGitRepo:       clusterGitRepo,
		commitGetter:         commitGetter,
		cd:                   cd,
		applicationMgr:       appMgr,
		templateReleaseMgr:   trMgr,
		templateSchemaGetter: templateSchemaGetter,
		envMgr:               envMgr,
		envRegionMgr:         envRegionMgr,
		regionMgr:            regionMgr,
		groupSvc:             groupservice.NewService(manager),
		pipelinerunMgr:       manager.PipelinerunMgr,
		tektonFty:            tektonFty,
		registryFty:          registryFty,
		userManager:          manager.UserManager,
		userSvc:              userservice.NewService(manager),
		schemaTagManager:     manager.ClusterSchemaTagMgr,
		tagMgr:               tagManager,
		applicationGitRepo:   applicationGitRepo,
	}

	tagManager.EXPECT().ListByResourceTypeIDs(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	applicationGitRepo.EXPECT().GetApplicationEnvTemplate(ctx, gomock.Any(), gomock.Any()).
		Return(pipelineJSONBlob, applicationJSONBlob, nil).AnyTimes()
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
		GitRepoSSHURL: "ssh://xxxx",
		ValueFiles:    []string{},
	}).AnyTimes()
	imageName := "image"
	clusterGitRepo.EXPECT().UpdatePipelineOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("image-commit", nil).AnyTimes()
	cd.EXPECT().CreateCluster(ctx, gomock.Any()).Return(nil).AnyTimes()
	cd.EXPECT().Pause(ctx, gomock.Any()).Return(nil).AnyTimes()
	cd.EXPECT().Resume(ctx, gomock.Any()).Return(nil).AnyTimes()
	cd.EXPECT().Promote(ctx, gomock.Any()).Return(nil).AnyTimes()

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
		},
		Name:       "app-cluster",
		ExpireTime: "24h",
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

	UpdateGitURL := "git@github.com:demo/demo.git"
	updateClusterRequest := &UpdateClusterRequest{
		Base: &Base{
			Description: "new description",
			Git: &codemodels.Git{
				URL:       UpdateGitURL,
				Subfolder: "/new",
				Branch:    "new",
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

	resp, err = c.GetCluster(ctx, resp.ID)
	assert.Nil(t, err)
	assert.Equal(t, resp.Git.URL, UpdateGitURL)
	assert.Equal(t, resp.Git.Branch, "new")
	assert.Equal(t, resp.Git.Subfolder, "/new")
	assert.Equal(t, resp.FullPath, "/group/app/app-cluster")
	// NOTE: template name cannot be edited! template release can be edited
	assert.Equal(t, resp.Template.Name, "javaapp")
	assert.Equal(t, resp.Template.Release, "v1.0.1")
	assert.Equal(t, resp.TemplateInput.Application, applicationJSONBlob)
	assert.Equal(t, resp.TemplateInput.Pipeline, pipelineJSONBlob)

	count, respList, err := c.ListCluster(ctx, application.ID, []string{"test"}, "", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, count, 1)
	t.Logf("%v", respList[0])
	assert.Equal(t, respList[0].Template.Name, "javaapp")
	assert.Equal(t, respList[0].Template.Release, "v1.0.1")

	count, respList, err = c.ListCluster(ctx, application.ID, []string{}, "", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, count, 2)
	t.Logf("%+v", respList[0].Scope)
	t.Logf("%+v", respList[1].Scope)

	count, respList, err = c.ListCluster(ctx, application.ID, []string{"test", "dev"}, "", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, count, 2)
	t.Logf("%+v", respList[0].Scope)
	t.Logf("%+v", respList[1].Scope)

	getByName, err := c.GetClusterByName(ctx, "app-cluster")
	assert.Nil(t, err)
	t.Logf("%v", getByName)

	tekton := tektonmock.NewMockInterface(mockCtl)
	tektonFty.EXPECT().GetTekton(gomock.Any()).Return(tekton, nil).AnyTimes()
	tekton.EXPECT().CreatePipelineRun(ctx, gomock.Any()).Return("abc", nil)
	tekton.EXPECT().GetPipelineRunByID(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(pr, nil).AnyTimes()

	registry := registrymock.NewMockRegistry(mockCtl)
	registry.EXPECT().CreateProject(ctx, gomock.Any()).Return(1, nil)
	registryFty.EXPECT().GetByHarborConfig(ctx, gomock.Any()).Return(registry).AnyTimes()

	commitGetter.EXPECT().GetCommit(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(&code.Commit{
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
	clusterGitRepo.EXPECT().MergeBranch(ctx, gomock.Any(), gomock.Any(),
		gomock.Any()).Return("newest-commit", nil).AnyTimes()
	clusterGitRepo.EXPECT().GetRepoInfo(ctx, gomock.Any(), gomock.Any()).Return(&gitrepo.RepoInfo{
		GitRepoSSHURL: "ssh://xxxx.git",
		ValueFiles:    []string{"file1", "file2"},
	}).AnyTimes()

	cd.EXPECT().DeployCluster(ctx, gomock.Any()).Return(nil).AnyTimes()
	cd.EXPECT().GetClusterState(ctx, gomock.Any()).Return(nil, herrors.NewErrNotFound(herrors.PodsInK8S, "test"))
	internalDeployResp, err := c.InternalDeploy(ctx, resp.ID, &InternalDeployRequest{
		PipelinerunID: buildDeployResp.PipelinerunID,
	})
	assert.Nil(t, err)
	b, _ = json.Marshal(internalDeployResp)
	t.Logf("%v", string(b))

	clusterStatusResp, err := c.GetClusterStatus(ctx, resp.ID)
	assert.Nil(t, err)
	b, _ = json.Marshal(clusterStatusResp)
	t.Logf("%v", string(b))

	codeBranch := "master"
	commitID := "code-commit-id"
	commitMsg := "code-commit-msg"
	configDiff := "config-diff"
	commitGetter.EXPECT().GetCommit(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(&code.Commit{
		ID:      commitID,
		Message: commitMsg,
	}, nil)
	clusterGitRepo.EXPECT().CompareConfig(ctx, gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).Return(configDiff, nil).AnyTimes()

	getdiffResp, err := c.GetDiff(ctx, resp.ID, codemodels.GitRefTypeBranch, codeBranch)
	assert.Nil(t, err)

	assert.Equal(t, &GetDiffResponse{
		CodeInfo: &CodeInfo{
			Branch:    codeBranch,
			CommitID:  commitID,
			CommitMsg: commitMsg,
			Link:      common.InternalSSHToHTTPURL(UpdateGitURL) + common.CommitHistoryMiddle + codeBranch,
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

	// test deploy
	clusterGitRepo.EXPECT().GetPipelineOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, herrors.ErrPipelineOutputEmpty).Times(1)

	deployResp, err := c.Deploy(ctx, resp.ID, &DeployRequest{
		Title:       "deploy-title",
		Description: "deploy-description",
	})
	assert.Equal(t, herrors.ErrShouldBuildDeployFirst, perror.Cause(err))
	assert.Nil(t, deployResp)

	err = c.Pause(ctx, resp.ID)
	assert.Nil(t, err)

	err = c.Resume(ctx, resp.ID)
	assert.Nil(t, err)

	err = c.Promote(ctx, resp.ID)
	assert.Nil(t, err)

	clusterGitRepo.EXPECT().GetPipelineOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(&gitrepo.PipelineOutput{},
		nil).Times(1)

	deployResp, err = c.Deploy(ctx, resp.ID, &DeployRequest{
		Title:       "deploy-title",
		Description: "deploy-description",
	})
	assert.Equal(t, herrors.ErrShouldBuildDeployFirst, perror.Cause(err))
	assert.Nil(t, deployResp)

	clusterGitRepo.EXPECT().GetPipelineOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(&gitrepo.PipelineOutput{
		Image: &imageName,
	}, nil).AnyTimes()

	deployResp, err = c.Deploy(ctx, resp.ID, &DeployRequest{
		Title:       "deploy-title",
		Description: "deploy-description",
	})
	assert.Nil(t, err)
	assert.NotNil(t, deployResp)

	b, _ = json.Marshal(deployResp)
	t.Logf("%s", string(b))

	// test next
	cd.EXPECT().Next(ctx, gomock.Any()).Return(nil)
	err = c.Next(ctx, resp.ID)
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

	cd.EXPECT().Online(ctx, gomock.Any()).Return(execResp, nil)
	cd.EXPECT().Offline(ctx, gomock.Any()).Return(execResp, nil)

	execRequest := &ExecRequest{
		PodList: []string{"pod1", "pod2"},
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

	// test rollback
	clusterGitRepo.EXPECT().Rollback(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
		Return("rollback-commit", nil).AnyTimes()
	// update status to 'ok'
	err = manager.PipelinerunMgr.UpdateResultByID(ctx, buildDeployResp.PipelinerunID, &prmodels.Result{
		Result: string(prmodels.StatusOK),
	})
	assert.Nil(t, err)
	rollbackResp, err := c.Rollback(ctx, resp.ID, &RollbackRequest{
		PipelinerunID: buildDeployResp.PipelinerunID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, rollbackResp)
	b, _ = json.Marshal(rollbackResp)
	t.Logf("%s", string(b))

	cd.EXPECT().DeletePods(ctx, gomock.Any()).Return(
		map[string]clustercd.OperationResult{
			"pod1": {Result: true},
		}, nil)
	result, err := c.DeleteClusterPods(ctx, resp.ID, []string{"pod1"})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	value, ok := result["pod1"]
	assert.Equal(t, true, ok)
	assert.Equal(t, true, value.Result)

	// test GetDashboard
	grafanaResponse, err := c.GetDashboard(ctx, resp.ID)
	assert.NotNil(t, err)
	assert.Nil(t, grafanaResponse)

	_, err = c.GetClusterPods(ctx, resp.ID, 0, 19)
	assert.NotNil(t, err)

	podExist := "exist"
	podNotExist := "notexist"
	cd.EXPECT().GetPod(ctx, gomock.Any()).DoAndReturn(
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
    baseRegistry: harbor.mock.org
    ingressDomain: mock.org
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

	renderObect, err := c.GetClusterOutput(context.TODO(), 123)
	assert.Nil(t, err)
	out, err := yaml.Marshal(renderObect)
	assert.Nil(t, err)
	var ExpectOutPutStr = `syncDomainName:
  Description: sync domain name
  Value: app-cluster-demo.mock.org
`
	assert.Equal(t, string(out), ExpectOutPutStr)
}

var envValue = `
javaapp:
  env:
    environment: pre
    region: hz
    namespace: pre-54
    baseRegistry: harbor.mock.org
    ingressDomain: mock.org

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
	var expectOutPutStr = `{"syncDomainName":{"Description":"sync domain name","Value":"music-social-zone-pre.mock.org"}}` // nolint
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
    baseRegistry: harbor.mock.org
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
		clustercommon.RestartTimeKey: t1.Format(layout),
	}
	Pod1.Spec.Containers = []*clustercd.Container{containerV1, containerV2}

	// pod2: t2, imagev1, imagev2
	Pod2.Metadata.Annotations = map[string]string{
		clustercommon.RestartTimeKey: t2.Format(layout),
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
