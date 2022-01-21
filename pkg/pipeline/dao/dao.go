package dao

import (
	"context"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/pipeline/models"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type DAO interface {
	// Create create a pipeline
	Create(ctx context.Context, results *metrics.PipelineResults) error

	ListPipelineSLOsByEnvsAndTimeRange(ctx context.Context, envs []string, start, end int64) ([]*models.PipelineSLO, error)
}

type dao struct{}

func (d dao) ListPipelineSLOsByEnvsAndTimeRange(ctx context.Context, envs []string,
	start, end int64) ([]*models.PipelineSLO, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	startTime := time.Unix(start, 0)
	endTime := time.Unix(end, 0)
	// search pipelines in given time range
	var pipelines []*models.Pipeline
	result := db.Raw(common.ListPipelinesByEnvsAndTimeRange, envs, startTime, endTime).Scan(&pipelines)
	if result.Error != nil {
		return nil, result.Error
	}
	// 指定时间段内没有相关流水线
	if len(pipelines) == 0 {
		return nil, nil
	}

	var pipelinerunIDs []uint
	for _, pipeline := range pipelines {
		pipelinerunIDs = append(pipelinerunIDs, pipeline.PipelinerunID)
	}

	// search tasks
	var tasks []*models.Task
	result = db.Raw(common.ListTasksByPipelinerunIDs, pipelinerunIDs).Scan(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}

	// search steps
	var steps []*models.Step
	result = db.Raw(common.ListStepsByPipelinerunIDs, pipelinerunIDs).Scan(&steps)
	if result.Error != nil {
		return nil, result.Error
	}

	return formatPipelineSLO(pipelines, tasks, steps), nil
}

func (d dao) Create(ctx context.Context, results *metrics.PipelineResults) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	prMetadata := results.Metadata
	prBusinessData := results.BusinessData
	prResult := results.PrResult
	trResults, stepResults := results.TrResults, results.StepResults

	pipeline := prMetadata.Pipeline
	pipelinerunIDStr := prBusinessData.PipelinerunID
	pipelinerunID, err := strconv.ParseUint(pipelinerunIDStr, 10, 0)
	if err != nil {
		return err
	}
	application, cluster, environment := prBusinessData.Application, prBusinessData.Cluster, prBusinessData.Environment

	err = db.Transaction(func(tx *gorm.DB) error {
		p := &models.Pipeline{
			PipelinerunID: uint(pipelinerunID),
			Application:   application,
			Cluster:       cluster,
			Environment:   environment,
			Pipeline:      pipeline,
			Result:        prResult.Result,
			StartedAt:     prResult.StartTime.Time,
			FinishedAt:    prResult.CompletionTime.Time,
			Duration:      uint(prResult.DurationSeconds),
		}
		result := tx.Create(p)
		if result.Error != nil {
			return result.Error
		}

		for _, trResult := range trResults {
			t := &models.Task{
				PipelinerunID: uint(pipelinerunID),
				Application:   application,
				Cluster:       cluster,
				Environment:   environment,
				Pipeline:      pipeline,
				Task:          trResult.Task,
				Result:        trResult.Result,
				StartedAt:     trResult.StartTime.Time,
				FinishedAt:    trResult.CompletionTime.Time,
				Duration:      uint(trResult.DurationSeconds),
			}
			result = tx.Create(t)
			if result.Error != nil {
				return result.Error
			}
		}

		for _, stepResult := range stepResults {
			s := &models.Step{
				PipelinerunID: uint(pipelinerunID),
				Application:   application,
				Cluster:       cluster,
				Environment:   environment,
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
				return result.Error
			}
		}

		return nil
	})

	return err
}

func formatPipelineSLO(pipelines []*models.Pipeline, tasks []*models.Task, steps []*models.Step) []*models.PipelineSLO {
	// format PipelineSLO structure
	stepMap := make(map[uint]map[string]map[string]*models.StepSLO)
	for _, step := range steps {
		if task2Step, ok := stepMap[step.PipelinerunID]; ok {
			task2Step[step.Task][step.Step] = &models.StepSLO{
				Step:     step.Step,
				Result:   step.Result,
				Duration: step.Duration,
			}
		} else {
			stepMap[step.PipelinerunID] = make(map[string]map[string]*models.StepSLO)
			stepMap[step.PipelinerunID][step.Task] = map[string]*models.StepSLO{
				step.Step: {
					Step:     step.Step,
					Result:   step.Result,
					Duration: step.Duration,
				},
			}
		}
	}
	taskMap := make(map[uint]map[string]*models.TaskSLO)
	for _, task := range tasks {
		taskSlo := &models.TaskSLO{
			Task:     task.Task,
			Result:   task.Result,
			Duration: task.Duration,
			Steps:    stepMap[task.PipelinerunID][task.Task],
		}
		if item, ok := taskMap[task.PipelinerunID]; ok {
			item[task.Task] = taskSlo
		} else {
			taskMap[task.PipelinerunID] = make(map[string]*models.TaskSLO)
			taskMap[task.PipelinerunID][task.Task] = taskSlo
		}
	}
	var slos []*models.PipelineSLO
	for _, pipeline := range pipelines {
		slos = append(slos, &models.PipelineSLO{
			Pipeline: pipeline.Pipeline,
			Result:   pipeline.Result,
			Duration: pipeline.Duration,
			Tasks:    taskMap[pipeline.PipelinerunID],
		})
	}

	return slos
}

func NewDAO() DAO {
	return &dao{}
}
