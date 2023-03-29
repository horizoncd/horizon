package envtemplate

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/horizoncd/horizon/pkg/core/common"
	"github.com/horizoncd/horizon/pkg/core/controller/build"
	"github.com/horizoncd/horizon/lib/orm"
	appgitrepomock "github.com/horizoncd/horizon/mock/pkg/application/gitrepo"
	trschemamock "github.com/horizoncd/horizon/mock/pkg/templaterelease/schema"
	"github.com/horizoncd/horizon/pkg/application/gitrepo"
	"github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	envmodels "github.com/horizoncd/horizon/pkg/environment/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	templatesvc "github.com/horizoncd/horizon/pkg/templaterelease/schema"
	"github.com/stretchr/testify/assert"
)

// nolint
var (
	ctx context.Context
	c   Controller

	applicationSchema, pipelineSchema     map[string]interface{}
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}

	appName = "app"
	env     = "dev"

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
	if err := db.AutoMigrate(&models.Application{}, &envmodels.Environment{}, &membermodels.Member{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
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
	applicationGitRepo.EXPECT().CreateOrUpdateApplication(ctx, gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	applicationGitRepo.EXPECT().GetApplication(ctx, gomock.Any(), gomock.Any()).
		Return(&gitrepo.GetResponse{
			Manifest:     nil,
			BuildConf:    pipelineJSONBlob,
			TemplateConf: applicationJSONBlob,
		}, nil).AnyTimes()

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

	envMgr := manager.EnvMgr
	applicationMgr := manager.ApplicationManager
	c = &controller{
		applicationGitRepo:   applicationGitRepo,
		templateSchemaGetter: templateSchemaGetter,
		applicationMgr:       applicationMgr,
		envMgr:               envMgr,
	}

	_, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name: env,
	})
	assert.Nil(t, err)

	app, err := applicationMgr.Create(ctx, &models.Application{
		Name:            appName,
		Template:        "javaapp",
		TemplateRelease: "v1.0.0",
	}, nil)
	assert.Nil(t, err)

	updateRequest := &UpdateEnvTemplateRequest{
		EnvTemplate: &EnvTemplate{
			Application: applicationJSONBlob,
			Pipeline:    pipelineJSONBlob,
		},
	}

	err = c.UpdateEnvTemplate(ctx, app.ID, env, updateRequest)
	assert.Nil(t, err)

	resp, err := c.GetEnvTemplate(ctx, app.ID, env)
	assert.Nil(t, err)
	b, _ := json.Marshal(resp)
	t.Logf("%v", string(b))
}

func TestV2(t *testing.T) {
	mockCtl := gomock.NewController(t)
	applicationGitRepo := appgitrepomock.NewMockApplicationGitRepo2(mockCtl)
	applicationGitRepo.EXPECT().CreateOrUpdateApplication(ctx, gomock.Any(), gitrepo.CreateOrUpdateRequest{
		Version:      common.MetaVersion2,
		Environment:  env,
		BuildConf:    pipelineJSONBlob,
		TemplateConf: applicationJSONBlob,
	}).
		Return(nil).AnyTimes()
	applicationGitRepo.EXPECT().GetApplication(ctx, gomock.Any(), gomock.Any()).
		Return(&gitrepo.GetResponse{
			Manifest:     nil,
			BuildConf:    pipelineJSONBlob,
			TemplateConf: applicationJSONBlob,
		}, nil).AnyTimes()

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

	envMgr := manager.EnvMgr
	applicationMgr := manager.ApplicationManager
	c = &controller{
		applicationGitRepo:   applicationGitRepo,
		templateSchemaGetter: templateSchemaGetter,
		applicationMgr:       applicationMgr,
		envMgr:               envMgr,
		buildSchema: &build.Schema{
			JSONSchema: pipelineSchema,
			UISchema:   nil,
		},
	}

	_, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name: env,
	})
	assert.Nil(t, err)

	app, err := applicationMgr.Create(ctx, &models.Application{
		Name:            appName,
		Template:        "javaapp",
		TemplateRelease: "v1.0.0",
	}, nil)
	assert.Nil(t, err)

	updateRequest := &UpdateEnvTemplateRequest{
		EnvTemplate: &EnvTemplate{
			Application: applicationJSONBlob,
			Pipeline:    pipelineJSONBlob,
		},
	}

	err = c.UpdateEnvTemplateV2(ctx, app.ID, env, updateRequest)
	assert.Nil(t, err)

	resp, err := c.GetEnvTemplate(ctx, app.ID, env)
	assert.Nil(t, err)
	b, _ := json.Marshal(resp)
	t.Logf("%v", string(b))
}
