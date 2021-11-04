package collector

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

//nolint:lll
func Test_resolveObjMetadata(t *testing.T) {
	jsonBody1 := `{
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
	var pr *v1beta1.PipelineRun
	if err := json.Unmarshal([]byte(jsonBody1), &pr); err != nil {
		t.Fatal(err)
	}
	type args struct {
		pr *v1beta1.PipelineRun
	}
	tests := []struct {
		name string
		args args
		want *ObjectMeta
	}{
		{
			name: "normal",
			args: args{
				pr: pr,
			},
			want: &ObjectMeta{
				Application:       "ndp-gjq",
				Cluster:           "test-music-docker",
				Environment:       "test",
				CreationTimestamp: "20210716165154",
				PipelineRun: &PipelineRunStatus{
					StatusMeta: StatusMeta{
						Name:            "test-music-docker-q58rp",
						Result:          "ok",
						DurationSeconds: 127,
					},
					Pipeline: "default",
					TaskRuns: map[string]TaskRunStatus{
						"test-music-docker-q58rp-build-g8khd": {
							StatusMeta: StatusMeta{
								Name:            "test-music-docker-q58rp-build-g8khd",
								Result:          "ok",
								DurationSeconds: 32,
							},
							Pod:  "test-music-docker-q58rp-build-g8khd-pod-mwsld",
							Task: "build",
							Steps: map[string]StepStatus{
								"git": {
									StatusMeta: StatusMeta{
										Name:            "git",
										Result:          "ok",
										DurationSeconds: 8,
									},
								},
								"compile": {
									StatusMeta: StatusMeta{
										Name:            "compile",
										Result:          "ok",
										DurationSeconds: 8,
									},
								},
								"image": {
									StatusMeta: StatusMeta{
										Name:            "image",
										Result:          "ok",
										DurationSeconds: 8,
									},
								},
							},
							StepsOrder: []string{"git", "compile", "image"},
						},
						"test-music-docker-q58rp-deploy-xzjkg": {
							StatusMeta: StatusMeta{
								Name:            "test-music-docker-q58rp-deploy-xzjkg",
								Result:          "ok",
								DurationSeconds: 95,
							},
							Pod:  "test-music-docker-q58rp-deploy-xzjkg-pod-zkcc4",
							Task: "deploy",
							Steps: map[string]StepStatus{
								"deploy": {
									StatusMeta: StatusMeta{
										Name:            "deploy",
										Result:          "ok",
										DurationSeconds: 90,
									},
								},
							},
							StepsOrder: []string{"deploy"},
						},
					},
					TasksOrder: []string{"build", "deploy"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveObjMetadata(tt.args.pr); !reflect.DeepEqual(got, tt.want) {
				gotB, _ := json.Marshal(got)
				wantB, _ := json.Marshal(tt.want)
				t.Errorf("ResolveObjMetadata() = %v\n, want %v", string(gotB), string(wantB))
			}
		})
	}
}
