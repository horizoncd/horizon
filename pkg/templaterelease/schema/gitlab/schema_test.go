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

package gitlab

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	gitlablibmock "github.com/horizoncd/horizon/mock/lib/gitlab"
	tmock "github.com/horizoncd/horizon/mock/pkg/template/manager"
	trmock "github.com/horizoncd/horizon/mock/pkg/templaterelease/manager"
	tmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/horizoncd/horizon/pkg/templaterelease/schema"
	"github.com/horizoncd/horizon/pkg/util/errors"
	"github.com/stretchr/testify/assert"
)

var (
	ctx                   = context.Background()
	templateName          = "javaapp"
	releaseName           = "v1.0.0"
	templateGitlabProject = "helm-template/javaapp"
)

func TestFunc(t *testing.T) {
	file := `
{
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
                            "title":"副本数",
                            "default": 1,
                            "minimum":0,
                            "maximum":{{ .maxReplicas | default 30}}
                        },
                        "resource":{
                            "type":"string",
                            "title":"规格",
                            "default":"x-small",
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
                            "default": 1,
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
                            "default": "all",
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
                            "description":"存活状态会在应用运行期间检测应用健康情况，检测失败时会对应用进行重启。接口如: /api/test"
                        },
                        "status":{
                            "type":"string",
                            "pattern":"^/.*$",
                            "title":"就绪状态",
                            "description":"就绪状态会在应用运行期间检测应用是否处于上线状态，检测失败时显示下线状态。接口如: /health/status"
                        },
                        "online":{
                            "type":"string",
                            "pattern":"^/.*$",
                            "title":"上线",
                            "description":"上线接口会在应用启动之后进行调用，如果调用失败，则应用启动失败。接口如: /health/online"
                        },
                        "offline":{
                            "type":"string",
                            "pattern":"^/.*$",
                            "title":"下线",
                            "description":"下线接口会在应用停止之前进行调用，如果调用失败，则忽略。接口如: /health/offline"
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
	params := make(map[string]string)
	params["maxReplicas"] = "122"
	files, err := schema.RenderFiles(params, []byte(file))
	assert.Nil(t, err)
	t.Logf("%s", string(files[0]))
}

func TestNoTag(t *testing.T) {
	mockCtl := gomock.NewController(t)
	gitlabLib := gitlablibmock.NewMockInterface(mockCtl)
	templateReleaseMgr := trmock.NewMockTemplateReleaseManager(mockCtl)
	templateMgr := tmock.NewMockTemplateManager(mockCtl)

	templateMgr.EXPECT().GetByName(ctx, templateName).
		Return(&tmodels.Template{
			Name:       templateName,
			Repository: templateGitlabProject,
		}, nil)
	templateReleaseMgr.EXPECT().GetByTemplateNameAndRelease(ctx, templateName,
		releaseName).Return(&tmodels.TemplateRelease{
		Model: global.Model{
			ID: 1,
		},
		TemplateName: templateName,
		ChartVersion: releaseName,
	}, nil)

	templateMgr.EXPECT().GetByName(ctx, templateName).
		Return(&tmodels.Template{
			Name:       templateName,
			Repository: templateGitlabProject,
		}, nil)
	templateReleaseMgr.EXPECT().GetByTemplateNameAndRelease(ctx, templateName,
		"release-not-exists").Return(nil, errors.E("", http.StatusNotFound))

	jsonSchema := `{"type": "object"}`
	var jsonSchemaMap map[string]interface{}
	_ = json.Unmarshal([]byte(jsonSchema), &jsonSchemaMap)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _pipelineSchemaPath).Return(
		[]byte(jsonSchema), nil)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _applicationSchemaPath).Return(
		[]byte(jsonSchema), nil)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _pipelineUISchemaPath).Return(
		[]byte(jsonSchema), nil)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _applicationUISchemaPath).Return(
		[]byte(jsonSchema), nil)

	g := &getter{
		templateMgr:        templateMgr,
		templateReleaseMgr: templateReleaseMgr,
		gitlabLib:          gitlabLib,
	}

	// release exists
	schema, err := g.GetTemplateSchema(ctx, templateName, releaseName, nil)
	assert.Nil(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, jsonSchemaMap, schema.Application.JSONSchema)
	assert.Equal(t, jsonSchemaMap, schema.Application.UISchema)
	assert.Equal(t, jsonSchemaMap, schema.Pipeline.JSONSchema)
	assert.Equal(t, jsonSchemaMap, schema.Pipeline.UISchema)

	// release not exists
	schema, err = g.GetTemplateSchema(ctx, templateName, "release-not-exists", nil)
	assert.Nil(t, schema)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, errors.Status(err))
}
