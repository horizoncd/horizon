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
	"strings"
	"testing"

	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

//nolint:lll
func TestObserve(t *testing.T) {
	wprBody := `{
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
                                    "message":"[{\"key\":\"properties\",\"value\":\"harbor.cloudnative.com/ndp/test-music-docker:helloworld-b1f57848-20210624143634 ssh://git@cloudnative.com:22222/demo/springboot-demo.git helloworld b1f578488e3123e97ec00b671db143fb8f0abecf\",\"type\":\"TaskRunResult\"}]",
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

	prHistogramMetric := `
        # HELP horizon_pipelinerun_duration_seconds PipelineRun duration info
        # TYPE horizon_pipelinerun_duration_seconds histogram
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="0"} 0
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="5"} 0
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="10"} 0
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="20"} 0
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="30"} 0
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="40"} 0
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="50"} 0
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="60"} 0
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="90"} 0
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="120"} 0
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="150"} 1
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="180"} 1
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="240"} 1
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="300"} 1
        horizon_pipelinerun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",template="serverless",le="+Inf"} 1
        horizon_pipelinerun_duration_seconds_sum{application="horizon",environment="test",pipeline="default",result="ok",template="serverless"} 127
        horizon_pipelinerun_duration_seconds_count{application="horizon",environment="test",pipeline="default",result="ok",template="serverless"} 1
		`

	trHistogramMetric := `
        # HELP horizon_taskrun_duration_seconds Taskrun duration info
        # TYPE horizon_taskrun_duration_seconds histogram
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="0"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="5"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="10"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="20"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="30"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="40"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="50"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="60"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="90"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="120"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="150"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="180"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="240"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="300"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless",le="+Inf"} 1
        horizon_taskrun_duration_seconds_sum{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless"} 32
        horizon_taskrun_duration_seconds_count{application="horizon",environment="test",pipeline="default",result="ok",task="build",template="serverless"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="0"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="5"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="10"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="20"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="30"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="40"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="50"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="60"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="90"} 0
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="120"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="150"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="180"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="240"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="300"} 1
        horizon_taskrun_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless",le="+Inf"} 1
        horizon_taskrun_duration_seconds_sum{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless"} 95
        horizon_taskrun_duration_seconds_count{application="horizon",environment="test",pipeline="default",result="ok",task="deploy",template="serverless"} 1
		`

	stepHistogramMetric := `
        # HELP horizon_step_duration_seconds Step duration info
        # TYPE horizon_step_duration_seconds histogram
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="0"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="5"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="10"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="20"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="30"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="40"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="50"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="60"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="90"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="120"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="150"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="180"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="240"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="300"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless",le="+Inf"} 1
        horizon_step_duration_seconds_sum{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless"} 8
        horizon_step_duration_seconds_count{application="horizon",environment="test",pipeline="default",result="ok",step="compile",task="build",template="serverless"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="0"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="5"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="10"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="20"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="30"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="40"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="50"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="60"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="90"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="120"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="150"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="180"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="240"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="300"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless",le="+Inf"} 1
        horizon_step_duration_seconds_sum{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless"} 90
        horizon_step_duration_seconds_count{application="horizon",environment="test",pipeline="default",result="ok",step="deploy",task="deploy",template="serverless"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="0"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="5"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="10"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="20"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="30"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="40"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="50"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="60"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="90"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="120"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="150"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="180"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="240"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="300"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless",le="+Inf"} 1
        horizon_step_duration_seconds_sum{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless"} 8
        horizon_step_duration_seconds_count{application="horizon",environment="test",pipeline="default",result="ok",step="git",task="build",template="serverless"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="0"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="5"} 0
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="10"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="20"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="30"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="40"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="50"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="60"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="90"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="120"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="150"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="180"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="240"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="300"} 1
        horizon_step_duration_seconds_bucket{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless",le="+Inf"} 1
        horizon_step_duration_seconds_sum{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless"} 8
        horizon_step_duration_seconds_count{application="horizon",environment="test",pipeline="default",result="ok",step="image",task="build",template="serverless"} 1
		`

	var wpr1 *WrappedPipelineRun
	_ = json.Unmarshal([]byte(wprBody), &wpr1)
	businessDatas := &global.HorizonMetaData{
		Application: "horizon",
		Environment: "test",
		Template:    "serverless",
	}
	Observe(FormatPipelineResults(wpr1.PipelineRun), businessDatas)
	if err := testutil.CollectAndCompare(_prHistogram, strings.NewReader(prHistogramMetric)); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := testutil.CollectAndCompare(_trHistogram, strings.NewReader(trHistogramMetric)); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := testutil.CollectAndCompare(_stepHistogram, strings.NewReader(stepHistogramMetric)); err != nil {
		t.Fatalf("err: %v", err)
	}
}
