package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
	"g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/models"
	"gorm.io/gorm"
)

var (
	ErrInsertPipeline = errors.New("Insert pipeline error")
)

type DAO interface {
	// Create create a pipeline
	Create(ctx context.Context, results *metrics.PipelineResults) error
	// ListPipelineStats list pipeline stats by query struct
	ListPipelineStats(ctx context.Context, query *q.Query) ([]*models.PipelineStats, int64, error)
}

type dao struct{ db *gorm.DB }

func (d dao) Create(ctx context.Context, results *metrics.PipelineResults) error {
	prMetadata := results.Metadata
	prBusinessData := results.BusinessData
	prResult := results.PrResult
	trResults, stepResults := results.TrResults, results.StepResults

	pipeline := prMetadata.Pipeline
	application, cluster, region := prBusinessData.Application, prBusinessData.Cluster, prBusinessData.Region

	pipelinerunID := prBusinessData.PipelinerunID
	err := d.db.Transaction(func(tx *gorm.DB) error {
		p := &models.Pipeline{
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
			t := &models.Task{
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
			s := &models.Step{
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

func (d *dao) ListPipelineStats(ctx context.Context, query *q.Query) ([]*models.PipelineStats, int64, error) {
	var pipelines []*models.Pipeline

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

	var tasks []*models.Task
	result = d.db.Where(map[string]interface{}{"pipelinerun_id": pipelinerunIDs}).Find(&tasks)
	if result.Error != nil {
		return nil, 0, herrors.NewErrListFailed(herrors.TaskInDB, result.Error.Error())
	}

	var steps []*models.Step
	result = d.db.Where(map[string]interface{}{"pipelinerun_id": pipelinerunIDs}).Find(&steps)
	if result.Error != nil {
		return nil, 0, herrors.NewErrListFailed(herrors.StepInDB, result.Error.Error())
	}

	return formatPipelineStats(pipelines, tasks, steps), count, nil
}

func formatPipelineStats(pipelines []*models.Pipeline, tasks []*models.Task,
	steps []*models.Step) []*models.PipelineStats {
	stepMap := make(map[uint]map[string][]*models.StepStats)
	for _, step := range steps {
		if task2Step, ok := stepMap[step.PipelinerunID]; ok {
			task2Step[step.Task] = append(task2Step[step.Task], &models.StepStats{
				Step:     step.Step,
				Result:   step.Result,
				Duration: step.Duration,
			})
		} else {
			stepMap[step.PipelinerunID] = make(map[string][]*models.StepStats)
			stepMap[step.PipelinerunID][step.Task] = []*models.StepStats{
				{
					Step:     step.Step,
					Result:   step.Result,
					Duration: step.Duration,
				},
			}
		}
	}

	taskMap := make(map[uint][]*models.TaskStats)
	for _, task := range tasks {
		taskStats := &models.TaskStats{
			Task:     task.Task,
			Result:   task.Result,
			Duration: task.Duration,
			Steps:    stepMap[task.PipelinerunID][task.Task],
		}
		if _, ok := taskMap[task.PipelinerunID]; !ok {
			taskMap[task.PipelinerunID] = []*models.TaskStats{}
		}
		taskMap[task.PipelinerunID] = append(taskMap[task.PipelinerunID], taskStats)
	}

	var stats []*models.PipelineStats
	for _, pipeline := range pipelines {
		stats = append(stats, &models.PipelineStats{
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

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}
