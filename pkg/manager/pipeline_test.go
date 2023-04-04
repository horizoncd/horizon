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
	"reflect"
	"testing"
	"time"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/metrics"
	models2 "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var tm = time.Now()

func createPipelineCtx() (context.Context, *gorm.DB, *metav1.Time) {
	db, _ := orm.NewSqliteDB("")
	var err = db.AutoMigrate(&models2.Pipeline{})
	_ = db.AutoMigrate(&models2.Task{})
	_ = db.AutoMigrate(&models2.Step{})

	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}

	startTime := &metav1.Time{
		Time: tm,
	}
	return context.TODO(), db, startTime
}

func Test_manager_Create(t *testing.T) {
	ctx, db, startTime := createPipelineCtx()
	data := &global.HorizonMetaData{
		Application:   "a",
		Cluster:       "c",
		Environment:   "dev",
		PipelinerunID: 1,
	}
	type createArgs struct {
		results *metrics.PipelineResults
	}
	tests := []struct {
		name    string
		args    createArgs
		wantErr bool
	}{
		{
			name:    "1",
			wantErr: false,
			args: createArgs{
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
			m := NewPipelineManager(db)
			if err := m.Create(ctx, tt.args.results, data); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	type listArgs struct {
		application string
		cluster     string
		pageNumber  int
		pageSize    int
	}
	listTests := []struct {
		name    string
		args    listArgs
		want    []*models2.PipelineStats
		wantErr bool
	}{
		{
			name: "1",
			args: listArgs{
				application: "a",
				cluster:     "c",
				pageNumber:  1,
				pageSize:    50,
			},
			want: []*models2.PipelineStats{
				{
					Pipeline:      "horizon-pipeline",
					Application:   "a",
					Cluster:       "c",
					Result:        "failed",
					Duration:      0,
					PipelinerunID: 1,
					StartedAt:     startTime.Time,
					Tasks: []*models2.TaskStats{
						{
							Task:     "build",
							Result:   "failed",
							Duration: 0,
							Steps: []*models2.StepStats{
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
	for _, tt := range listTests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewPipelineManager(db)
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
