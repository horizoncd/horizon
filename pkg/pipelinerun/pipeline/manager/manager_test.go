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

package manager

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/metrics"
	"github.com/horizoncd/horizon/pkg/pipelinerun/pipeline/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// use tmp sqlite
	db, _     = orm.NewSqliteDB("")
	ctx       = context.TODO()
	startTime *metav1.Time
)

func init() {
	// create table
	var err = db.AutoMigrate(&models.Pipeline{})
	_ = db.AutoMigrate(&models.Task{})
	_ = db.AutoMigrate(&models.Step{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}

	startTime = &metav1.Time{
		Time: time.Now(),
	}
}

func Test_manager_Create(t *testing.T) {
	data := &global.HorizonMetaData{
		Application:   "a",
		Cluster:       "c",
		Environment:   "dev",
		PipelinerunID: 1,
	}
	type args struct {
		results *metrics.PipelineResults
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "1",
			wantErr: false,
			args: args{
				results: &metrics.PipelineResults{
					Metadata: &metrics.PrMetadata{
						Pipeline: "horizon-pipeline",
					},
					PrResult: &metrics.PrResult{
						DurationSeconds: 0,
						Result:          "failed",
						StartTime:       startTime,
						CompletionTime: &metav1.Time{
							Time: time.Now(),
						},
					},
					TrResults: metrics.TrResults{
						{
							Task: "build",
							StartTime: &metav1.Time{
								Time: time.Now(),
							},
							CompletionTime: &metav1.Time{
								Time: time.Now(),
							},
							DurationSeconds: 0,
							Result:          "failed",
						},
					},
					StepResults: metrics.StepResults{
						{
							Step: "git",
							Task: "build",
							StartTime: &metav1.Time{
								Time: time.Now(),
							},
							CompletionTime: &metav1.Time{
								Time: time.Now(),
							},
							DurationSeconds: 0,
							Result:          "failed",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(db)
			if err := m.Create(ctx, tt.args.results, data); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_ListPipelineStats(t *testing.T) {
	type args struct {
		application string
		cluster     string
		pageNumber  int
		pageSize    int
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.PipelineStats
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				application: "a",
				cluster:     "c",
				pageNumber:  1,
				pageSize:    50,
			},
			want: []*models.PipelineStats{
				{
					Pipeline:      "horizon-pipeline",
					Application:   "a",
					Cluster:       "c",
					Result:        "failed",
					Duration:      0,
					PipelinerunID: 1,
					StartedAt:     startTime.Time,
					Tasks: []*models.TaskStats{
						{
							Task:     "build",
							Result:   "failed",
							Duration: 0,
							Steps: []*models.StepStats{
								{
									Step:     "git",
									Result:   "failed",
									Duration: 0,
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(db)
			got, count, err := m.ListPipelineStats(ctx, tt.args.application,
				tt.args.cluster, tt.args.pageNumber, tt.args.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPipelineStats() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			got[0].StartedAt = tt.want[0].StartedAt
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListPipelineStats() got = %+v, want %+v", got, tt.want)
			}
			assert.Equal(t, int(count), 1)
		})
	}
}
