package collector

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/horizoncd/horizon/lib/s3"
	tektonmock "github.com/horizoncd/horizon/mock/pkg/cluster/tekton"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/log"
)

func getPipelineRunLog(pr *v1beta1.PipelineRun) (<-chan log.Log, <-chan error, error) {
	logCh := make(chan log.Log)
	pipeline := pr.ObjectMeta.Labels["tekton.dev/pipeline"]
	task := "test-task"
	step := "test-step"
	go func() {
		defer close(logCh)
		for i := 0; i < 10; i++ {
			logCh <- log.Log{
				Pipeline: pipeline,
				Task:     task,
				Step:     step,
				Log:      "line" + strconv.Itoa(i),
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	return logCh, nil, nil
}

func TestNewS3Collector(t *testing.T) {
	s3Driver := &s3.Driver{}
	ctl := gomock.NewController(t)
	tek := tektonmock.NewMockInterface(ctl)
	s := NewS3Collector(s3Driver, tek)
	assert.NotNil(t, s)
}

//nolint:lll
func TestS3Collector_Collect(t *testing.T) {
	jsonBody1 := `{
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
	var pr *v1beta1.PipelineRun
	if err := json.Unmarshal([]byte(jsonBody1), &pr); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	ctl := gomock.NewController(t)
	tek := tektonmock.NewMockInterface(ctl)
	tek.EXPECT().GetPipelineRunLog(ctx, pr).Return(getPipelineRunLog(pr))

	backend := s3mem.New()
	_ = backend.CreateBucket("bucket")
	faker := gofakes3.New(backend)
	ts := httptest.NewServer(faker.Server())
	defer ts.Close()

	params := &s3.Params{
		AccessKey:        "accessKey",
		SecretKey:        "secretKey",
		Region:           "us-east-1",
		Endpoint:         ts.URL,
		Bucket:           "bucket",
		ContentType:      "text/plain",
		SkipVerify:       true,
		S3ForcePathStyle: true,
	}

	d, err := s3.NewDriver(*params)
	assert.Nil(t, err)

	c := NewS3Collector(d, tek)

	businessDatas := &global.HorizonMetaData{
		Application: "app",
		Cluster:     "cluster",
		Environment: "test",
	}

	// collect
	collectResult, err := c.Collect(ctx, pr, businessDatas)
	assert.Nil(t, err)
	b, _ := json.Marshal(collectResult)
	t.Logf("%v", string(b))

	// 1. getLatestPipelineRunLog
	b, err = c.GetPipelineRunLog(ctx, collectResult.LogObject)
	assert.Nil(t, err)
	t.Logf(string(b))
	// 2. getLatestPipelineRunObject
	obj, err := c.GetPipelineRunObject(ctx, collectResult.PrObject)
	assert.Nil(t, err)
	assert.NotNil(t, obj)
	objectMeta := resolveObjMetadata(pr, businessDatas)
	if !reflect.DeepEqual(objectMeta, obj.Metadata) {
		t.Fatalf("pipelineRun objectMeta: expected %v, got %v", objectMeta, obj.Metadata)
	}
}
