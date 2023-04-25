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

package metrics

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	prmodels "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const _layout = "2006-01-02T15:04:05Z"

func TestWrappedPipelineRun_ResolveMetadata(t *testing.T) {
	type fields struct {
		PipelineRun *v1beta1.PipelineRun
	}
	tests := []struct {
		name   string
		fields fields
		want   *PrMetadata
	}{
		{
			name: "normal1",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pr1",
						Namespace: "ns1",
						Labels: map[string]string{
							"tekton.dev/pipeline": "p1",
						},
					},
				},
			},
			want: &PrMetadata{
				Name:      "pr1",
				Namespace: "ns1",
				Pipeline:  "p1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wpr := &WrappedPipelineRun{
				PipelineRun: tt.fields.PipelineRun,
			}
			if got := wpr.ResolveMetadata(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolveMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrappedPipelineRun_ResolvePrResult(t *testing.T) {
	type fields struct {
		PipelineRun *v1beta1.PipelineRun
	}
	tests := []struct {
		name   string
		fields fields
		want   *PrResult
	}{
		{
			name: "resultOK1",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonSuccessful),
									},
								},
							),
						},
						PipelineRunStatusFields: v1beta1.PipelineRunStatusFields{
							StartTime:      parseTime("2021-07-08T06:36:11Z"),
							CompletionTime: parseTime("2021-07-08T06:38:18Z"),
						},
					},
				},
			},
			want: &PrResult{
				DurationSeconds: 127,
				Result:          string(prmodels.StatusOK),
				StartTime:       parseTime("2021-07-08T06:36:11Z"),
				CompletionTime:  parseTime("2021-07-08T06:38:18Z"),
			},
		},
		{
			name: "resultOK2",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonCompleted),
									},
								},
							),
						},
						PipelineRunStatusFields: v1beta1.PipelineRunStatusFields{
							StartTime:      parseTime("2021-07-08T06:36:11Z"),
							CompletionTime: parseTime("2021-07-08T06:37:18Z"),
						},
					},
				},
			},
			want: &PrResult{
				DurationSeconds: 67,
				Result:          string(prmodels.StatusOK),
				StartTime:       parseTime("2021-07-08T06:36:11Z"),
				CompletionTime:  parseTime("2021-07-08T06:37:18Z"),
			},
		},
		{
			name: "resultFailed1",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonFailed),
									},
								},
							),
						},
						PipelineRunStatusFields: v1beta1.PipelineRunStatusFields{
							StartTime:      parseTime("2021-07-08T06:36:11Z"),
							CompletionTime: parseTime("2021-07-08T06:38:10Z"),
						},
					},
				},
			},
			want: &PrResult{
				DurationSeconds: 119,
				Result:          string(prmodels.StatusFailed),
				StartTime:       parseTime("2021-07-08T06:36:11Z"),
				CompletionTime:  parseTime("2021-07-08T06:38:10Z"),
			},
		},
		{
			name: "resultFailed2",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonTimedOut),
									},
								},
							),
						},
						PipelineRunStatusFields: v1beta1.PipelineRunStatusFields{
							StartTime:      parseTime("2021-07-08T06:36:11Z"),
							CompletionTime: parseTime("2021-07-08T06:38:14Z"),
						},
					},
				},
			},
			want: &PrResult{
				DurationSeconds: 123,
				Result:          string(prmodels.StatusFailed),
				StartTime:       parseTime("2021-07-08T06:36:11Z"),
				CompletionTime:  parseTime("2021-07-08T06:38:14Z"),
			},
		},
		{
			name: "resultCancelled",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonCancelled),
									},
								},
							),
						},
						PipelineRunStatusFields: v1beta1.PipelineRunStatusFields{
							StartTime:      parseTime("2021-07-08T06:36:11Z"),
							CompletionTime: parseTime("2021-07-08T06:38:14Z"),
						},
					},
				},
			},
			want: &PrResult{
				DurationSeconds: 123,
				Result:          string(prmodels.StatusCancelled),
				StartTime:       parseTime("2021-07-08T06:36:11Z"),
				CompletionTime:  parseTime("2021-07-08T06:38:14Z"),
			},
		},
		{
			name: "resultUnknown",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: "",
									},
								},
							),
						},
						PipelineRunStatusFields: v1beta1.PipelineRunStatusFields{
							StartTime:      parseTime("2021-07-08T06:36:11Z"),
							CompletionTime: parseTime("2021-07-08T06:38:14Z"),
						},
					},
				},
			},
			want: &PrResult{
				DurationSeconds: 123,
				Result:          string(prmodels.StatusUnknown),
				StartTime:       parseTime("2021-07-08T06:36:11Z"),
				CompletionTime:  parseTime("2021-07-08T06:38:14Z"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wpr := &WrappedPipelineRun{
				PipelineRun: tt.fields.PipelineRun,
			}
			if got := wpr.ResolvePrResult(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolvePrResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:lll
func TestWrappedPipelineRun_ResolveTrAndStepResults(t *testing.T) {
	jsonBody1 := `{
    "pipelineRun":{
        "metadata":{
            "name":"test-music-docker-q58rp",
            "namespace":"tekton-resources",
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
	}
	`

	jsonBody2 := `{
    "pipelineRun":{
        "metadata":{
            "labels":{
                "app.kubernetes.io/managed-by":"Helm",
                "tekton.dev/pipeline":"default",
                "triggers.tekton.dev/eventlistener":"default-listener",
                "triggers.tekton.dev/trigger":"",
                "triggers.tekton.dev/triggers-eventid":"ksw7h"
            },
            "name":"music-datastream-serverless-fetest-task-ld9s6",
            "namespace":"tekton-resources"
        },
        "status":{
            "completionTime":"2021-06-17T06:33:37Z",
            "conditions":[
                {
                    "lastTransitionTime":"2021-06-17T06:33:37Z",
                    "message":"Tasks Completed: 1 (Failed: 1, Cancelled 0), Skipped: 1",
                    "reason":"Failed",
                    "status":"False",
                    "type":"Succeeded"
                }
            ],
            "skippedTasks":[
                {
                    "name":"deploy"
                }
            ],
            "startTime":"2021-06-17T06:33:27Z",
            "taskRuns":{
                "music-datastream-serverless-fetest-task-ld9s6-build-2dg26":{
                    "pipelineTaskName":"build",
                    "status":{
                        "completionTime":"2021-06-17T06:33:36Z",
                        "conditions":[
                            {
                                "lastTransitionTime":"2021-06-17T06:33:36Z",
                                "message":"\"step-compile\" exited with code 1 (image: \"docker-pullable://harbor.mockserver.org/staffyun163music/cloudnative/tekton/builder@sha256:08fe6aa9f2cac16791e3c448b90fd73e1492e201a96359f7cf23550b37d62b72\"); for logs run: kubectl -n tekton-resources logs music-datastream-serverless-fetest-task-ld9s6-build-2dg26-8wp4g -c step-compile\n",
                                "reason":"Failed",
                                "status":"False",
                                "type":"Succeeded"
                            }
                        ],
                        "podName":"music-datastream-serverless-fetest-task-ld9s6-build-2dg26-8wp4g",
                        "startTime":"2021-06-17T06:33:28Z",
                        "steps":[
                            {
                                "container":"step-git",
                                "imageID":"docker-pullable://harbor.mockserver.org/staffyun163music/cloudnative/tekton/builder@sha256:08fe6aa9f2cac16791e3c448b90fd73e1492e201a96359f7cf23550b37d62b72",
                                "name":"git",
                                "terminated":{
                                    "containerID":"docker://041709ac75cb00e3ce31fbacf465b750f54e73a47672fce4edee8c3810c1af69",
                                    "exitCode":0,
                                    "finishedAt":"2021-06-17T06:33:35Z",
                                    "reason":"Completed",
                                    "startedAt":"2021-06-17T06:33:34Z"
                                }
                            },
                            {
                                "container":"step-compile",
                                "imageID":"docker-pullable://harbor.mockserver.org/staffyun163music/cloudnative/tekton/builder@sha256:08fe6aa9f2cac16791e3c448b90fd73e1492e201a96359f7cf23550b37d62b72",
                                "name":"compile",
                                "terminated":{
                                    "containerID":"docker://ce4d00d5c973995d3f70737e315b060d64620abb5f255e49de91a017f1146159",
                                    "exitCode":1,
                                    "finishedAt":"2021-06-17T06:33:36Z",
                                    "reason":"Error",
                                    "startedAt":"2021-06-17T06:33:36Z"
                                }
                            },
                            {
                                "container":"step-image",
                                "imageID":"docker-pullable://harbor.mockserver.org/staffyun163music/cloudnative/kaniko-executor@sha256:473d6dfb011c69f32192e668d86a47c0235791e7e857c870ad70c5e86ec07e8c",
                                "name":"image",
                                "terminated":{
                                    "containerID":"docker://17108818edcd23b49fa0327fb329ea20a6ab43e8d6156bf51b3edacaafb42305",
                                    "exitCode":1,
                                    "finishedAt":"2021-06-17T06:33:36Z",
                                    "reason":"Error",
                                    "startedAt":"2021-06-17T06:33:36Z"
                                }
                            }
                        ]
                    }
                }
            }
        }
    }
	}`
	var wpr1, wpr2 *WrappedPipelineRun
	_ = json.Unmarshal([]byte(jsonBody1), &wpr1)
	_ = json.Unmarshal([]byte(jsonBody2), &wpr2)
	type fields struct {
		PipelineRun *v1beta1.PipelineRun
	}
	tests := []struct {
		name   string
		fields fields
		want   TrResults
		want1  StepResults
	}{
		{
			name: "normal",
			fields: fields{
				PipelineRun: wpr1.PipelineRun,
			},
			want: TrResults{
				{
					Name:            "test-music-docker-q58rp-build-g8khd",
					Pod:             "test-music-docker-q58rp-build-g8khd-pod-mwsld",
					Task:            "build",
					StartTime:       parseTime("2021-06-24T06:36:11Z"),
					CompletionTime:  parseTime("2021-06-24T06:36:43Z"),
					DurationSeconds: 32,
					Result:          string(prmodels.StatusOK),
				},
				{
					Name:            "test-music-docker-q58rp-deploy-xzjkg",
					Pod:             "test-music-docker-q58rp-deploy-xzjkg-pod-zkcc4",
					Task:            "deploy",
					StartTime:       parseTime("2021-06-24T06:36:43Z"),
					CompletionTime:  parseTime("2021-06-24T06:38:18Z"),
					DurationSeconds: 95,
					Result:          string(prmodels.StatusOK),
				},
			},
			want1: StepResults{
				{
					Step:            "git",
					Task:            "build",
					TaskRun:         "test-music-docker-q58rp-build-g8khd",
					StartTime:       parseTime("2021-06-24T06:36:18Z"),
					CompletionTime:  parseTime("2021-06-24T06:36:26Z"),
					DurationSeconds: 8,
					Result:          string(prmodels.StatusOK),
				},
				{
					Step:            "compile",
					Task:            "build",
					TaskRun:         "test-music-docker-q58rp-build-g8khd",
					StartTime:       parseTime("2021-06-24T06:36:26Z"),
					CompletionTime:  parseTime("2021-06-24T06:36:34Z"),
					DurationSeconds: 8,
					Result:          string(prmodels.StatusOK),
				},
				{
					Step:            "image",
					Task:            "build",
					TaskRun:         "test-music-docker-q58rp-build-g8khd",
					StartTime:       parseTime("2021-06-24T06:36:34Z"),
					CompletionTime:  parseTime("2021-06-24T06:36:42Z"),
					DurationSeconds: 8,
					Result:          string(prmodels.StatusOK),
				},
				{
					Step:            "deploy",
					Task:            "deploy",
					TaskRun:         "test-music-docker-q58rp-deploy-xzjkg",
					StartTime:       parseTime("2021-06-24T06:36:48Z"),
					CompletionTime:  parseTime("2021-06-24T06:38:18Z"),
					DurationSeconds: 90,
					Result:          string(prmodels.StatusOK),
				},
			},
		},
		{
			name: "failed",
			fields: fields{
				PipelineRun: wpr2.PipelineRun,
			},
			want: TrResults{
				{
					Name:            "music-datastream-serverless-fetest-task-ld9s6-build-2dg26",
					Pod:             "music-datastream-serverless-fetest-task-ld9s6-build-2dg26-8wp4g",
					Task:            "build",
					StartTime:       parseTime("2021-06-17T06:33:28Z"),
					CompletionTime:  parseTime("2021-06-17T06:33:36Z"),
					DurationSeconds: 8,
					Result:          string(prmodels.StatusFailed),
				},
			},
			want1: StepResults{
				{
					Step:            "git",
					Task:            "build",
					TaskRun:         "music-datastream-serverless-fetest-task-ld9s6-build-2dg26",
					StartTime:       parseTime("2021-06-17T06:33:34Z"),
					CompletionTime:  parseTime("2021-06-17T06:33:35Z"),
					DurationSeconds: 1,
					Result:          string(prmodels.StatusOK),
				},
				{
					Step:            "compile",
					Task:            "build",
					TaskRun:         "music-datastream-serverless-fetest-task-ld9s6-build-2dg26",
					StartTime:       parseTime("2021-06-17T06:33:36Z"),
					CompletionTime:  parseTime("2021-06-17T06:33:36Z"),
					DurationSeconds: 0,
					Result:          string(prmodels.StatusFailed),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wpr := &WrappedPipelineRun{
				PipelineRun: tt.fields.PipelineRun,
			}
			got, got1 := wpr.ResolveTrAndStepResults()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolveTrAndStepResults() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ResolveTrAndStepResults() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_durationSeconds(t *testing.T) {
	type args struct {
		beginTime *metav1.Time
		endTime   *metav1.Time
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "0",
			args: args{
				beginTime: parseTime("2021-07-08T06:38:14Z"),
				endTime:   parseTime("2021-07-08T06:38:14Z"),
			},
			want: 0,
		},
		{
			name: "negative",
			args: args{
				beginTime: parseTime("2021-07-08T06:38:18Z"),
				endTime:   parseTime("2021-07-08T06:38:14Z"),
			},
			want: -4,
		},
		{
			name: "positive normal",
			args: args{
				beginTime: parseTime("2021-07-08T06:38:18Z"),
				endTime:   parseTime("2021-07-08T06:41:28Z"),
			},
			want: 190,
		},
		{
			name: "nil1",
			args: args{
				beginTime: nil,
				endTime:   parseTime("2021-07-08T06:38:14Z"),
			},
			want: -1,
		},
		{
			name: "nil2",
			args: args{
				beginTime: parseTime("2021-07-08T06:38:14Z"),
				endTime:   nil,
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := durationSeconds(tt.args.beginTime, tt.args.endTime); got != tt.want {
				t.Errorf("durationSeconds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func parseTime(str string) *metav1.Time {
	t, _ := time.Parse(_layout, str)
	mt := metav1.NewTime(t.Local())
	return &mt
}
