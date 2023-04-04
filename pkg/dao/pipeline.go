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

package dao

import (
	"context"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/metrics"
	"github.com/horizoncd/horizon/pkg/errors"
	models2 "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	"gorm.io/gorm"
)

var (
	ErrInsertPipeline = errors.New("Insert pipeline error")
)

type PipelineDAO interface {
	// Create create a pipeline
	Create(ctx context.Context, results *metrics.PipelineResults, data *global.HorizonMetaData) error
	// ListPipelineStats list pipeline stats by query struct
	ListPipelineStats(ctx context.Context, query *q.Query) ([]*models2.PipelineStats, int64, error)
}

type pipelineDAO struct{ db *gorm.DB }

func (d pipelineDAO) Create(ctx context.Context, results *metrics.PipelineResults, data *global.HorizonMetaData) error {
	prMetadata := results.Metadata
	prResult := results.PrResult
	trResults, stepResults := results.TrResults, results.StepResults

	pipeline := prMetadata.Pipeline
	application, cluster, region := data.Application, data.Cluster, data.Region

	pipelinerunID := data.PipelinerunID
	err := d.db.Transaction(func(tx *gorm.DB) error {
		p := &models2.Pipeline{
			PipelinerunID: pipelinerunID,
			Application:   application,
			Cluster:       cluster,
			Region:        region,
			Pipeline:      pipeline,
			Result:        prResult.Result,
			StartedAt:     prResult.StartTime.Time,
			FinishedAt:    prResult.CompletionTime.Time,
			Duration:      uint(prResult.DurationSeconds),
		}
		result := tx.Create(p)
		if result.Error != nil {
			return errors.Wrap(ErrInsertPipeline, result.Error.Error())
		}

		for _, trResult := range trResults {
			if trResult.CompletionTime == nil {
				continue
			}
			t := &models2.Task{
				PipelinerunID: pipelinerunID,
				Application:   application,
				Cluster:       cluster,
				Region:        region,
				Pipeline:      pipeline,
				Task:          trResult.Task,
				Result:        trResult.Result,
				StartedAt:     trResult.StartTime.Time,
				FinishedAt:    trResult.CompletionTime.Time,
				Duration:      uint(trResult.DurationSeconds),
			}
			result = tx.Create(t)
			if result.Error != nil {
				return errors.Wrap(ErrInsertPipeline, result.Error.Error())
			}
		}

		for _, stepResult := range stepResults {
			if stepResult.CompletionTime == nil {
				continue
			}
			s := &models2.Step{
				PipelinerunID: pipelinerunID,
				Application:   application,
				Cluster:       cluster,
				Region:        region,
				Pipeline:      pipeline,
				Task:          stepResult.Task,
				Step:          stepResult.Step,
				Result:        stepResult.Result,
				StartedAt:     stepResult.StartTime.Time,
				FinishedAt:    stepResult.CompletionTime.Time,
				Duration:      uint(stepResult.DurationSeconds),
			}
			result = tx.Create(s)
			if result.Error != nil {
				return errors.Wrap(ErrInsertPipeline, result.Error.Error())
			}
		}

		return nil
	})

	return err
}

func (d *pipelineDAO) ListPipelineStats(ctx context.Context, query *q.Query) ([]*models2.PipelineStats, int64, error) {
	var pipelines []*models2.Pipeline

	sort := orm.FormatSortExp(query)
	offset := (query.PageNumber - 1) * query.PageSize
	var count int64
	result := d.db.Order(sort).Where(query.Keywords).Offset(offset).Limit(query.PageSize).Find(&pipelines).
		Offset(-1).Count(&count)
	if result.Error != nil {
		return nil, 0, herrors.NewErrListFailed(herrors.PipelineInDB, result.Error.Error())
	}

	var pipelinerunIDs []uint
	for _, pipeline := range pipelines {
		pipelinerunIDs = append(pipelinerunIDs, pipeline.PipelinerunID)
	}

	var tasks []*models2.Task
	result = d.db.Where(map[string]interface{}{"pipelinerun_id": pipelinerunIDs}).Find(&tasks)
	if result.Error != nil {
		return nil, 0, herrors.NewErrListFailed(herrors.TaskInDB, result.Error.Error())
	}

	var steps []*models2.Step
	result = d.db.Where(map[string]interface{}{"pipelinerun_id": pipelinerunIDs}).Find(&steps)
	if result.Error != nil {
		return nil, 0, herrors.NewErrListFailed(herrors.StepInDB, result.Error.Error())
	}

	return formatPipelineStats(pipelines, tasks, steps), count, nil
}

func formatPipelineStats(pipelines []*models2.Pipeline, tasks []*models2.Task,
	steps []*models2.Step) []*models2.PipelineStats {
	stepMap := make(map[uint]map[string][]*models2.StepStats)
	for _, step := range steps {
		if task2Step, ok := stepMap[step.PipelinerunID]; ok {
			task2Step[step.Task] = append(task2Step[step.Task], &models2.StepStats{
				Step:     step.Step,
				Result:   step.Result,
				Duration: step.Duration,
			})
		} else {
			stepMap[step.PipelinerunID] = make(map[string][]*models2.StepStats)
			stepMap[step.PipelinerunID][step.Task] = []*models2.StepStats{
				{
					Step:     step.Step,
					Result:   step.Result,
					Duration: step.Duration,
				},
			}
		}
	}

	taskMap := make(map[uint][]*models2.TaskStats)
	for _, task := range tasks {
		taskStats := &models2.TaskStats{
			Task:     task.Task,
			Result:   task.Result,
			Duration: task.Duration,
			Steps:    stepMap[task.PipelinerunID][task.Task],
		}
		if _, ok := taskMap[task.PipelinerunID]; !ok {
			taskMap[task.PipelinerunID] = []*models2.TaskStats{}
		}
		taskMap[task.PipelinerunID] = append(taskMap[task.PipelinerunID], taskStats)
	}

	var stats []*models2.PipelineStats
	for _, pipeline := range pipelines {
		stats = append(stats, &models2.PipelineStats{
			Application:   pipeline.Application,
			Cluster:       pipeline.Cluster,
			Pipeline:      pipeline.Pipeline,
			Result:        pipeline.Result,
			Duration:      pipeline.Duration,
			Tasks:         taskMap[pipeline.PipelinerunID],
			StartedAt:     pipeline.StartedAt,
			PipelinerunID: pipeline.PipelinerunID,
		})
	}

	return stats
}

func NewPipelineDAO(db *gorm.DB) PipelineDAO {
	return &pipelineDAO{db: db}
}
