package manager

import (
	"context"
	"fmt"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
	"g.hz.netease.com/horizon/pkg/pipeline/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"reflect"
	"testing"
	"time"
)

var (
	// use tmp sqlite
	db, _ = orm.NewSqliteDB("")
	ctx   = orm.NewContext(context.TODO(), db)
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
}

func Test_manager_Create(t *testing.T) {
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
					BusinessData: &metrics.PrBusinessData{
						Application:   "a",
						Cluster:       "c",
						Environment:   "dev",
						PipelinerunID: "1",
					},
					PrResult: &metrics.PrResult{
						DurationSeconds: 0,
						Result:          "failed",
						StartTime: &metav1.Time{
							Time: time.Now(),
						},
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
			m := New()
			if err := m.Create(ctx, tt.args.results); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := m.ListPipelineSLOsByEnvsAndTimeRange(ctx, []string{"dev"}, time.Now().Unix()-3600, time.Now().Unix())
			for _, item := range got {
				t.Logf("got: %v", item)
			}
			if err != nil {
				t.Errorf("ListPipelineSLOsByEnvsAndTimeRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_ListPipelineSLOsByEnvsAndTimeRange(t *testing.T) {
	type args struct {
		envs  []string
		start int64
		end   int64
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.PipelineSLO
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				envs:  []string{"dev"},
				start: time.Now().Unix() - 3600,
				end:   time.Now().Unix() + 1,
			},
			want: []*models.PipelineSLO{
				{
					Pipeline: "horizon-pipeline",
					Result:   "failed",
					Duration: 0,
					Tasks: map[string]*models.TaskSLO{
						"build": {
							Task:     "build",
							Result:   "failed",
							Duration: 0,
							Steps: map[string]*models.StepSLO{
								"git": {
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
			m := New()
			got, err := m.ListPipelineSLOsByEnvsAndTimeRange(ctx, tt.args.envs, tt.args.start, tt.args.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPipelineSLOsByEnvsAndTimeRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListPipelineSLOsByEnvsAndTimeRange() got = %v, want %v", got, tt.want)
			}
		})
	}
}
