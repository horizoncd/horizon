package application

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	appgitrepomock "g.hz.netease.com/horizon/mock/pkg/application/gitrepo"
	trschemamock "g.hz.netease.com/horizon/mock/pkg/templaterelease/schema"
	"g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	groupservice "g.hz.netease.com/horizon/pkg/group/service"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	templatesvc "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	userservice "g.hz.netease.com/horizon/pkg/user/service"

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
	if err := db.AutoMigrate(&models.Application{}, &clustermodels.Cluster{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&groupmodels.Group{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&trmodels.TemplateRelease{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&membermodels.Member{}); err != nil {
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
	applicationGitRepo := appgitrepomock.NewMockApplicationGitRepo(mockCtl)
	applicationGitRepo.EXPECT().CreateApplication(ctx, appName, pipelineJSONBlob, applicationJSONBlob).Times(1).Return(nil)
	applicationGitRepo.EXPECT().UpdateApplication(ctx, appName, pipelineJSONBlob, applicationJSONBlob).Return(nil).AnyTimes()
	applicationGitRepo.EXPECT().DeleteApplication(ctx, appName, uint(1)).Return(nil).AnyTimes()
	applicationGitRepo.EXPECT().GetApplication(ctx, appName).Return(pipelineJSONBlob, applicationJSONBlob, nil).AnyTimes()

	templateSchemaGetter := trschemamock.NewMockGetter(mockCtl)
	templateSchemaGetter.EXPECT().GetTemplateSchema(ctx, "javaapp", "v1.0.0", nil).
		Return(&templatesvc.Schemas{
			Application: &templatesvc.Schema{
				JSONSchema: applicationSchema,
			},
			Pipeline: &templatesvc.Schema{
				JSONSchema: pipelineSchema,
			},
		}, nil).AnyTimes()

	c = &controller{
		applicationGitRepo:   applicationGitRepo,
		templateSchemaGetter: templateSchemaGetter,
		applicationMgr:       manager.ApplicationManager,
		groupMgr:             manager.GroupManager,
		groupSvc:             groupservice.NewService(manager),
		templateReleaseMgr:   manager.TemplateReleaseManager,
		clusterMgr:           manager.ClusterMgr,
		userSvc:              userservice.NewService(manager),
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
			Git: &Git{
				URL:       "ssh://git@g.hz.netease.com:22222/music-cloud-native/horizon/horizon.git",
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
		t.Logf("%v", err)
		t.Fatal(err)
	}
	t.Logf("%v", resp)

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
			Git: &Git{
				URL:       "ssh://git@g.hz.netease.com:22222/music-cloud-native/horizon/horizon.git",
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
	t.Logf("%v", resp)

	updateRequest.Priority = "no-exists"
	_, err = c.UpdateApplication(ctx, resp.ID, updateRequest)
	assert.NotNil(t, err)

	resp, err = c.GetApplication(ctx, resp.ID)
	t.Logf("resp: %v", resp)
	assert.Nil(t, err)

	assert.Equal(t, resp.Description, updatedDescription)

	err = c.DeleteApplication(ctx, resp.ID)
	assert.Nil(t, err)
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
			GitBranch:       "master",
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

	count, resps, err := c.ListUserApplication(ctx, "appFu", &q.Query{
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
	count, resps, err = c.ListUserApplication(ctx, "appFu", &q.Query{
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
