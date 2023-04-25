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

package tekton

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

func TestGetPipelineRunStatus(t *testing.T) {
	type args struct {
		pr *v1beta1.PipelineRun
	}
	tests := []struct {
		name string
		args args
		want *PipelineRunStatus
	}{
		{
			name: "normal",
			args: args{
				pr: &v1beta1.PipelineRun{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pr",
					},
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: []apis.Condition{
								{
									Type:   apis.ConditionSucceeded,
									Reason: string(v1beta1.PipelineRunReasonRunning),
								},
							},
						},
						PipelineRunStatusFields: v1beta1.PipelineRunStatusFields{
							TaskRuns: map[string]*v1beta1.PipelineRunTaskRunStatus{
								"taskrun1": {
									PipelineTaskName: "build",
									Status: &v1beta1.TaskRunStatus{
										Status: duckv1beta1.Status{
											Conditions: []apis.Condition{
												{
													Type:   apis.ConditionSucceeded,
													Reason: string(v1beta1.PipelineRunReasonSuccessful),
												},
											},
										},
									},
								},
								"taskrun2": {
									PipelineTaskName: "deploy",
									Status: &v1beta1.TaskRunStatus{
										Status: duckv1beta1.Status{
											Conditions: []apis.Condition{
												{
													Type:   apis.ConditionSucceeded,
													Reason: string(v1beta1.PipelineRunReasonRunning),
												},
											},
										},
									},
								},
							},
							PipelineSpec: &v1beta1.PipelineSpec{
								Tasks: []v1beta1.PipelineTask{
									{
										Name: "build",
									}, {
										Name: "deploy",
									},
								},
							},
						},
					},
				},
			},
			want: &PipelineRunStatus{
				Name: "pr",
				RunningTask: &RunningTask{
					Name:   "deploy",
					Status: "Running",
				},
				Status: "Running",
			},
		},
		{
			name: "nil",
			args: args{
				pr: nil,
			},
			want: nil,
		}, {
			name: "nil status",
			args: args{
				pr: &v1beta1.PipelineRun{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pr",
					},
				},
			},
			want: &PipelineRunStatus{
				Name:        "pr",
				RunningTask: nil,
				Status:      "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPipelineRunStatus(context.Background(), tt.args.pr); !reflect.DeepEqual(got, tt.want) {
				b, _ := json.MarshalIndent(got, "", "")
				fmt.Printf("got: %v", string(b))
				t.Errorf("GetPipelineRunStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
