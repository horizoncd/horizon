package application

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/horizoncd/horizon/pkg/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	appgitrepomock "github.com/horizoncd/horizon/mock/pkg/application/gitrepo"
	trschemamock "github.com/horizoncd/horizon/mock/pkg/templaterelease/schema"
	"github.com/horizoncd/horizon/pkg/application/gitrepo"
	"github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	groupservice "github.com/horizoncd/horizon/pkg/group/service"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	tmodels "github.com/horizoncd/horizon/pkg/template/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
	trschema "github.com/horizoncd/horizon/pkg/templaterelease/schema"
	userservice "github.com/horizoncd/horizon/pkg/user/service"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// nolint
var (
	ctx context.Context
	c   Controller

	applicationSchema, pipelineSchema     map[string]interface{}
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}

	appName = "app"

	applicationSchemaJSON = `{
  "type": "object",
  "properties": {
    "app": {
      "title": "App Config",
      "type": "object",
      "properties": {
        "params": {
          "title": "参数",
          "type": "object",
          "properties": {
            "mainClassName": {
              "type": "string"
            },
            "xmx": {
              "type": "string",
              "default": "512"
            },
            "xms": {
              "type": "string",
              "default": "512"
            },
            "maxPerm": {
              "type": "string",
              "default": "128"
            },
            "xdebugAddress": {
              "type": "string"
            },
            "jvmExtra": {
              "type": "string"
            }
          },
          "required": ["mainClassName"]
        },
        "resource": {
          "type": "string",
          "title": "规格",
          "description": "应用上建议选择tiny或者small规格（测试环境集群自动继承，节省资源使用），线上集群可选大规格"
        },
        "health": {
          "title": "健康检查",
          "type": "object",
          "properties": {
            "port": {
              "type": "integer"
            },
            "lifecycle":{
              "title": "优雅启停",
              "type": "object",
              "properties": {
                "online":{
                  "title": "上线",
                  "description": "上线接口会在应用启动之后进行调用，如果调用失败，则应用启动失败",
                  "$ref": "#/$defs/lifecycle"
                },
                "offline":{
                  "title": "下线",
                  "description": "下线接口会在应用停止之前进行调用，如果调用失败，则忽略",
                  "$ref": "#/$defs/lifecycle"
                }
              }
            },
            "probe":{
              "title": "健康检查",
              "type": "object",
              "properties": {
                "check":{
                  "title": "存活状态",
                  "description": "存活状态会在应用运行期间检测应用健康情况，检测失败时会对应用进行重启",
                  "$ref": "#/$defs/probe"
                },
                "status":{
                  "title": "就绪状态",
                  "description": "就绪状态会在应用运行期间检测应用是否处于上线状态，检测失败时显示下线状态",
                  "$ref": "#/$defs/probe"
                }
              }
            }
          },
          "dependencies": {
            "lifecycle": ["port"],
            "probe": ["port"]
          }
        }
      }
    }
  },

  "$defs": {
    "lifecycle": {
      "type": "object",
      "properties": {
        "url": {
          "title": "接口",
          "description": "接口路径",
          "type": "string"
        },
        "timeoutSeconds": {
          "title": "超时时间",
          "description": "请求接口的超时时间(单位为s)",
          "type": "integer"
        },
        "periodSeconds": {
          "title": "检测周期",
          "description": "连续两次检测之间的时间间隔(单位为s)",
          "type": "integer"
        },
        "retry": {
          "title": "重试次数",
          "description": "检测失败后重试的次数",
          "type": "integer"
        }
      }
    },
    "probe": {
      "type": "object",
      "properties": {
        "url": {
          "title": "接口",
          "description": "接口路径",
          "type": "string"
        },
        "initialDelaySeconds": {
          "title": "延迟启动",
          "description": "应用启动延迟等待该时间再进行检测",
          "type": "integer"
        },
        "timeoutSeconds": {
          "title": "超时时间",
          "description": "请求接口的超时时间(单位为s)",
          "type": "integer"
        },
        "periodSeconds": {
          "title": "重试次数",
          "description": "连续两次检测之间的时间间隔(单位为s)",
          "type": "integer"
        },
        "failureThreshold": {
          "title": "失败阈值",
          "description": "连续检测失败超过该次数，才认为最终检测失败",
          "type": "integer"
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
      "default": "xxxxxxxxxxxxxxxxxxx"
    }
  }
}`

	pipelineJSONStr = `{
            "buildxml":"PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPCFET0NUWVBFIHByb2plY3QgWzwhRU5USVRZIGJ1aWxkZmlsZSBTWVNURU0gImZpbGU6Li9idWlsZC11c2VyLnhtbCI+XT4KPHByb2plY3QgYmFzZWRpcj0iLiIgZGVmYXVsdD0iZGVwbG95IiBuYW1lPSJkZW1vIj4KICAgIDxwcm9wZXJ0eSBuYW1lPSJhbnQiIHZhbHVlPSJhbnQiIC8+CiAgICA8cHJvcGVydHkgbmFtZT0iYmFzZWxpbmUuZGlyIiB2YWx1ZT0iJHtiYXNlZGlyfSIvPgoKICAgIDx0YXJnZXQgbmFtZT0icGFja2FnZSI+CiAgICAgICAgPGV4ZWMgZGlyPSIke2Jhc2VsaW5lLmRpcn0iIGV4ZWN1dGFibGU9IiR7YW50fSIgZmFpbG9uZXJyb3I9InRydWUiPgogICAgICAgICAgICA8YXJnIGxpbmU9Ii1idWlsZGZpbGUgb3Zlcm1pbmRfYnVpbGQueG1sIC1EZW52PXRlc3QgLURlbnZOYW1lPXFhLWFsbGFuLmlnYW1lLjE2My5jb20iLz4KICAgICAgICA8L2V4ZWM+CiAgICA8L3RhcmdldD4KCiAgICA8dGFyZ2V0IG5hbWU9ImRlcGxveSI+CiAgICAgICAgPGVjaG8gbWVzc2FnZT0iYmVnaW4gYXV0byBkZXBsb3kuLi4uLi4iLz4KICAgICAgICA8YW50Y2FsbCB0YXJnZXQ9InBhY2thZ2UiLz4KICAgIDwvdGFyZ2V0Pgo8L3Byb2plY3Q+"
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
	manager *managerparam.Manager
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
	if err := db.AutoMigrate(&models.Application{}, &clustermodels.Cluster{}, &regionmodels.Region{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&groupmodels.Group{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&tmodels.Template{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&trmodels.TemplateRelease{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&membermodels.Member{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&eventmodels.Event{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   1,
	})

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

	os.Exit(m.Run())
}

// nolint
func Test(t *testing.T) {
	mockCtl := gomock.NewController(t)
	applicationGitRepo := appgitrepomock.NewMockApplicationGitRepo2(mockCtl)
	applicationGitRepo.EXPECT().CreateOrUpdateApplication(ctx, appName, gitrepo.CreateOrUpdateRequest{
		Version:      "",
		Environment:  common.ApplicationRepoDefaultEnv,
		BuildConf:    pipelineJSONBlob,
		TemplateConf: applicationJSONBlob,
	}).Return(nil).AnyTimes()
	applicationGitRepo.EXPECT().HardDeleteApplication(ctx, appName).Return(nil).AnyTimes()
	applicationGitRepo.EXPECT().GetApplication(ctx, appName, common.ApplicationRepoDefaultEnv).Return(&gitrepo.GetResponse{
		Manifest:     nil,
		BuildConf:    pipelineJSONBlob,
		TemplateConf: applicationJSONBlob,
	}, nil).AnyTimes()

	templateSchemaGetter := trschemamock.NewMockGetter(mockCtl)
	templateSchemaGetter.EXPECT().GetTemplateSchema(ctx, "javaapp", "v1.0.0", nil).
		Return(&trschema.Schemas{
			Application: &trschema.Schema{
				JSONSchema: applicationSchema,
			},
			Pipeline: &trschema.Schema{
				JSONSchema: pipelineSchema,
			},
		}, nil).AnyTimes()

	tr := &trmodels.TemplateRelease{
		TemplateName: "javaapp",
		ChartVersion: "v1.0.0",
		Name:         "v1.0.0",
		ChartName:    "javaapp",
	}
	_, err := manager.TemplateReleaseManager.Create(ctx, tr)
	assert.Nil(t, err)

	c = &controller{
		applicationGitRepo:   applicationGitRepo,
		templateSchemaGetter: templateSchemaGetter,
		applicationMgr:       manager.ApplicationManager,
		groupMgr:             manager.GroupManager,
		groupSvc:             groupservice.NewService(manager),
		templateReleaseMgr:   manager.TemplateReleaseManager,
		clusterMgr:           manager.ClusterMgr,
		userSvc:              userservice.NewService(manager),
		eventMgr:             manager.EventManager,
	}

	group, err := manager.GroupManager.Create(ctx, &groupmodels.Group{
		Name: "ABC",
		Path: "abc",
	})
	assert.Nil(t, err)

	createRequest := &CreateApplicationRequest{
		Base: Base{
			Description: "this is a description",
			Priority:    "P0",
			Template: &Template{
				Name:    "javaapp",
				Release: "v1.0.0",
			},
			Git: &codemodels.Git{
				URL:       "ssh://git@cloudnative.com:22222/music-cloud-native/horizon/horizon.git",
				Subfolder: "/",
				Branch:    "develop",
			},
			TemplateInput: &TemplateInput{
				Application: applicationJSONBlob,
				Pipeline:    pipelineJSONBlob,
			},
		},
		Name: appName,
	}

	// create application
	resp, err := c.CreateApplication(ctx, group.ID, createRequest)
	if err != nil {
		t.Logf("%+v", err)
		t.Fatal(err)
	}
	t.Logf("%+v", resp)

	// create application again, end with error
	_, err = c.CreateApplication(ctx, group.ID, createRequest)
	assert.NotNil(t, err)

	updatedDescription := "updated description"
	updateRequest := &UpdateApplicationRequest{
		Base: Base{
			Description: updatedDescription,
			Priority:    "P0",
			Template: &Template{
				Name:    "javaapp",
				Release: "v1.0.0",
			},
			Git: &codemodels.Git{
				URL:       "ssh://git@cloudnative.com:22222/music-cloud-native/horizon/horizon.git",
				Subfolder: "/",
				Branch:    "develop",
			},
			TemplateInput: &TemplateInput{
				Application: applicationJSONBlob,
				Pipeline:    pipelineJSONBlob,
			},
		},
	}

	resp, err = c.UpdateApplication(ctx, resp.ID, updateRequest)
	assert.Nil(t, err)
	t.Logf("%+v", resp)

	updateRequest.Priority = "no-exists"
	_, err = c.UpdateApplication(ctx, resp.ID, updateRequest)
	assert.NotNil(t, err)

	resp, err = c.GetApplication(ctx, resp.ID)
	t.Logf("resp: %+v", resp)
	assert.Nil(t, err)

	assert.Equal(t, resp.Description, updatedDescription)

	err = c.DeleteApplication(ctx, resp.ID, false)
	assert.Nil(t, err)

	resp, err = c.CreateApplication(ctx, group.ID, createRequest)
	if err != nil {
		t.Logf("%v", err)
		t.Fatal(err)
	}
	t.Logf("%v", resp)

	getResponseV1, err := c.GetApplication(ctx, resp.ID)

	getReponsev2, err := c.GetApplicationV2(ctx, getResponseV1.ID)
	assert.Nil(t, err)
	assert.Equal(t, getReponsev2.ID, getResponseV1.ID)
	assert.Equal(t, getReponsev2.Name, getResponseV1.Name)
	assert.Equal(t, getReponsev2.Description, getResponseV1.Description)
	assert.Equal(t, getReponsev2.Priority, getResponseV1.Priority)
	assert.Equal(t, getReponsev2.Git, getResponseV1.Git)
	assert.Equal(t, getReponsev2.BuildConfig, getResponseV1.TemplateInput.Pipeline)
	assert.Equal(t, getReponsev2.TemplateConfig, getResponseV1.TemplateInput.Application)
	assert.Nil(t, getReponsev2.Manifest, nil)
	assert.Equal(t, getReponsev2.FullPath, getResponseV1.FullPath)
	assert.Equal(t, getReponsev2.GroupID, getResponseV1.GroupID)
	assert.Equal(t, getReponsev2.CreatedAt, getResponseV1.CreatedAt)
	assert.Equal(t, getReponsev2.UpdatedAt, getResponseV1.UpdatedAt)

	err = c.DeleteApplication(ctx, resp.ID, true)
	assert.Nil(t, err)
}

func TestV2(t *testing.T) {
	appName = "appname2"
	mockCtl := gomock.NewController(t)
	applicationGitRepo := appgitrepomock.NewMockApplicationGitRepo2(mockCtl)
	applicationGitRepo.EXPECT().CreateOrUpdateApplication(ctx, appName, gitrepo.CreateOrUpdateRequest{
		Version:      common.MetaVersion2,
		Environment:  common.ApplicationRepoDefaultEnv,
		BuildConf:    nil,
		TemplateConf: applicationJSONBlob,
	}).Return(nil).AnyTimes()
	applicationGitRepo.EXPECT().HardDeleteApplication(ctx, appName).Return(nil).AnyTimes()
	applicationGitRepo.EXPECT().GetApplication(ctx, appName,
		common.ApplicationRepoDefaultEnv).Return(&gitrepo.GetResponse{
		Manifest:     nil,
		BuildConf:    nil,
		TemplateConf: applicationJSONBlob,
	}, nil).AnyTimes()

	templateSchemaGetter := trschemamock.NewMockGetter(mockCtl)
	templateSchemaGetter.EXPECT().GetTemplateSchema(ctx, "javaapp", "v1.0.0", nil).
		Return(&trschema.Schemas{
			Application: &trschema.Schema{
				JSONSchema: applicationSchema,
			},
			Pipeline: &trschema.Schema{
				JSONSchema: pipelineSchema,
			},
		}, nil).AnyTimes()

	tr := &trmodels.TemplateRelease{
		TemplateName: "javaapp",
		ChartVersion: "v1.0.0",
		Name:         "v1.0.0",
		ChartName:    "javaapp",
	}
	_, err := manager.TemplateReleaseManager.Create(ctx, tr)
	assert.Nil(t, err)
	c := &controller{
		applicationGitRepo:   applicationGitRepo,
		templateSchemaGetter: templateSchemaGetter,
		applicationMgr:       manager.ApplicationManager,
		groupMgr:             manager.GroupManager,
		groupSvc:             groupservice.NewService(manager),
		templateReleaseMgr:   manager.TemplateReleaseManager,
		clusterMgr:           manager.ClusterMgr,
		userSvc:              userservice.NewService(manager),
		eventMgr:             manager.EventManager,
	}

	group, err := manager.GroupManager.Create(ctx, &groupmodels.Group{
		Name: "cde",
		Path: "cde",
	})
	assert.Nil(t, err)

	P0 := "P0"
	TemplateName := "javaapp"
	TemplateVersion := "v1.0.0"
	Description := "this is an v2 application interface"
	createReq := &CreateOrUpdateApplicationRequestV2{
		Name:        appName,
		Description: Description,
		Priority:    &P0,
		Git: &codemodels.Git{
			URL:       "ssh://git@cloudnative.com:22222/music-cloud-native/horizon/horizon.git",
			Subfolder: "/",
			Branch:    "develop",
		},
		BuildConfig: nil,
		TemplateInfo: &codemodels.TemplateInfo{
			Name:    TemplateName,
			Release: TemplateVersion,
		},
		TemplateConfig: applicationJSONBlob,
		ExtraMembers:   nil,
	}
	resp, err := c.CreateApplicationV2(ctx, group.ID, createReq)
	assert.Nil(t, err)
	assert.Equal(t, resp.GroupID, group.ID)

	// get application
	getResponse, err := c.GetApplicationV2(ctx, resp.ID)
	assert.Nil(t, err)
	assert.Equal(t, getResponse.ID, resp.ID)
	assert.Equal(t, getResponse.Name, appName)
	assert.Equal(t, getResponse.Description, Description)
	assert.Equal(t, getResponse.Priority, P0)
	assert.Equal(t, getResponse.Git, createReq.Git)
	assert.Equal(t, getResponse.BuildConfig, createReq.BuildConfig)
	assert.Equal(t, getResponse.TemplateInfo, createReq.TemplateInfo)
	assert.Equal(t, getResponse.TemplateConfig, createReq.TemplateConfig)
	assert.Nil(t, getResponse.Manifest)
	t.Logf("%+v", getResponse.Manifest)
}

func Test_validateApplicationName(t *testing.T) {
	var (
		name string
		err  error
	)

	name = "a"
	err = validateApplicationName(name)
	assert.Nil(t, err)

	name = "a1"
	err = validateApplicationName(name)
	assert.Nil(t, err)

	name = "a-1"
	err = validateApplicationName(name)
	assert.Nil(t, err)

	name = "1"
	err = validateApplicationName(name)
	assert.NotNil(t, err)
	t.Logf("%v", err)

	name = "0a"
	err = validateApplicationName(name)
	assert.NotNil(t, err)
	t.Logf("%v", err)

	name = "9a"
	err = validateApplicationName(name)
	assert.NotNil(t, err)
	t.Logf("%v", err)

	name = "1aaa"
	err = validateApplicationName(name)
	assert.NotNil(t, err)
	t.Logf("%v", err)

	name = "a-"
	err = validateApplicationName(name)
	assert.NotNil(t, err)
	t.Logf("%v", err)

	name = "a-A"
	err = validateApplicationName(name)
	assert.NotNil(t, err)
	t.Logf("%v", err)
}

func TestListUserApplication(t *testing.T) {
	// init data
	var groups []*groupmodels.Group
	for i := 0; i < 5; i++ {
		name := "groupForAppFuzzily" + strconv.Itoa(i)
		group, err := manager.GroupManager.Create(ctx, &groupmodels.Group{
			Name:     name,
			Path:     name,
			ParentID: 0,
		})
		assert.Nil(t, err)
		assert.NotNil(t, group)
		groups = append(groups, group)
	}

	var applications []*models.Application
	for i := 0; i < 5; i++ {
		group := groups[i]
		name := "appFuzzily" + strconv.Itoa(i)
		application, err := manager.ApplicationManager.Create(ctx, &models.Application{
			GroupID:         group.ID,
			Name:            name,
			Priority:        "P3",
			GitURL:          "ssh://git.com",
			GitSubfolder:    "/test",
			GitRef:          "master",
			Template:        "javaapp",
			TemplateRelease: "v1.0.0",
		}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, application)
		applications = append(applications, application)
	}

	c = &controller{
		applicationMgr: manager.ApplicationManager,
		groupMgr:       manager.GroupManager,
		groupSvc:       groupservice.NewService(manager),
		memberManager:  manager.MemberManager,
	}

	// nolint
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Matt",
		ID:   uint(2),
	})

	_, err := manager.MemberManager.Create(ctx, &membermodels.Member{
		ResourceType: membermodels.TypeGroup,
		ResourceID:   groups[0].ID,
		Role:         "owner",
		MemberType:   membermodels.MemberUser,
		MemberNameID: 2,
	})
	assert.Nil(t, err)

	_, err = manager.MemberManager.Create(ctx, &membermodels.Member{
		ResourceType: membermodels.TypeApplication,
		ResourceID:   applications[1].ID,
		Role:         "owner",
		MemberType:   membermodels.MemberUser,
		MemberNameID: 2,
	})
	assert.Nil(t, err)

	resps, count, err := c.List(ctx, &q.Query{
		Keywords: q.KeyWords{
			common.ClusterQueryByUser: uint(2),
			common.ClusterQueryName:   "appFu",
		},
		PageNumber: 0,
		PageSize:   common.DefaultPageSize,
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, "appFuzzily1", resps[0].Name)
	assert.Equal(t, "appFuzzily0", resps[1].Name)
	for _, resp := range resps {
		t.Logf("%v", resp)
	}

	_, err = manager.MemberManager.Create(ctx, &membermodels.Member{
		ResourceType: membermodels.TypeGroup,
		ResourceID:   groups[2].ID,
		Role:         "owner",
		MemberType:   membermodels.MemberUser,
		MemberNameID: 2,
	})
	assert.Nil(t, err)
	resps, count, err = c.List(ctx, &q.Query{
		Keywords: q.KeyWords{
			common.ClusterQueryByUser: uint(2),
			common.ClusterQueryName:   "appFu",
		},
		PageNumber: 0,
		PageSize:   common.DefaultPageSize,
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, "appFuzzily2", resps[0].Name)
	assert.Equal(t, "appFuzzily1", resps[1].Name)
	for _, resp := range resps {
		t.Logf("%v", resp)
	}
}
