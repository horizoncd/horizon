package dao

import (
	"context"
	"strconv"

	"g.hz.netease.com/horizon/lib/orm"
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
}

type dao struct{}

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
	application, cluster, region := prBusinessData.Application, prBusinessData.Cluster, prBusinessData.Region

	err = db.Transaction(func(tx *gorm.DB) error {
		p := &models.Pipeline{
			PipelinerunID: uint(pipelinerunID),
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
				PipelinerunID: uint(pipelinerunID),
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
				PipelinerunID: uint(pipelinerunID),
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

func NewDAO() DAO {
	return &dao{}
}
