package manager

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
	"g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// use tmp sqlite
	db, _ = orm.NewSqliteDB("")
	ctx   = context.TODO()
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
						RegionID:      "1",
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
			m := New(db)
			if err := m.Create(ctx, tt.args.results); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
