package cloudevent

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	tektonmock "g.hz.netease.com/horizon/mock/pkg/cluster/tekton"
	tektoncollectormock "g.hz.netease.com/horizon/mock/pkg/cluster/tekton/collector"
	tektonftymock "g.hz.netease.com/horizon/mock/pkg/cluster/tekton/factory"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/collector"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	pipelinemodels "g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/models"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	"github.com/stretchr/testify/assert"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	triggers "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/mock/gomock"
)

var (
	ctx context.Context

	pipelineRun *v1beta1.PipelineRun

	manager *managerparam.Manager
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
	if err := db.AutoMigrate(&prmodels.Pipelinerun{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&pipelinemodels.Pipeline{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&appmodels.Application{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&clustermodels.Cluster{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&usermodels.User{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&membermodels.Member{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   uint(1),
	})

	prJSON := `{
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
            },
			"annotations":{
                "cloudnative.music.netease.com/operator":"demo@mail.com"
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

	if err := json.Unmarshal([]byte(prJSON), &pipelineRun); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tektonFty := tektonftymock.NewMockFactory(mockCtl)
	tekton := tektonmock.NewMockInterface(mockCtl)
	tektonCollector := tektoncollectormock.NewMockInterface(mockCtl)
	tektonFty.EXPECT().GetTekton(gomock.Any()).Return(tekton, nil).AnyTimes()
	tektonFty.EXPECT().GetTektonCollector(gomock.Any()).Return(tektonCollector, nil).AnyTimes()

	tektonCollector.EXPECT().Collect(ctx, gomock.Any(), gomock.Any()).Return(&collector.CollectResult{
		Bucket:    "bucket",
		LogObject: "log-object",
		PrObject:  "pr-object",
		Result:    "ok",
		StartTime: func() *metav1.Time {
			tt := time.Now()
			return &metav1.Time{
				Time: tt,
			}
		}(),
		CompletionTime: func() *metav1.Time {
			tt := time.Now()
			return &metav1.Time{
				Time: tt,
			}
		}(),
	}, nil)

	tekton.EXPECT().DeletePipelineRun(ctx, gomock.Any()).Return(nil)

	application, _ := manager.ApplicationManager.Create(ctx, &appmodels.Application{
		Name: "app",
	}, map[string]string{})
	user, _ := manager.UserManager.Create(ctx, &usermodels.User{
		Name: "user",
	})
	cluster, _ := manager.ClusterMgr.Create(ctx, &clustermodels.Cluster{
		ApplicationID: application.ID,
		Name:          "cluster",
	}, nil, nil)
	pipelinerunMgr := manager.PipelinerunMgr
	_, err := pipelinerunMgr.Create(ctx, &prmodels.Pipelinerun{
		ClusterID:   cluster.ID,
		Action:      "builddeploy",
		Status:      "created",
		Title:       "title",
		Description: "description",
		CIEventID:   pipelineRun.Labels[triggers.GroupName+triggers.EventIDLabelKey],
		CreatedBy:   user.ID,
	})
	assert.Nil(t, err)

	c := NewController(tektonFty, &param.Param{Manager: manager})

	err = c.CloudEvent(ctx, &WrappedPipelineRun{
		PipelineRun: pipelineRun,
	})
	assert.Nil(t, err)

	pr, err := pipelinerunMgr.GetLatestByClusterIDAndAction(ctx, uint(1), prmodels.ActionBuildDeploy)
	assert.Nil(t, err)
	assert.Equal(t, pr.Status, "ok")
	assert.Equal(t, pr.LogObject, "log-object")
}
