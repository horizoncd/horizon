package dao

import (
	"context"
	"strconv"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
	"g.hz.netease.com/horizon/pkg/errors"
	models "g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/models"
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
	application, cluster, regionIDStr := prBusinessData.Application, prBusinessData.Cluster, prBusinessData.RegionID
	regionID, err := strconv.ParseUint(regionIDStr, 10, 0)
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		p := &models.Pipeline{
			PipelinerunID: uint(pipelinerunID),
			Application:   application,
			Cluster:       cluster,
			RegionID:      uint(regionID),
			Pipeline:      pipeline,
			Result:        prResult.Result,
			StartedAt:     prResult.StartTime.Time,
			FinishedAt:    prResult.CompletionTime.Time,
			Duration:      uint(prResult.DurationSeconds),
		}
		result := tx.Create(p)
		if result.Error != nil {
			return errors.Wrap(ErrInsertPipeline, err.Error())
		}

		for _, trResult := range trResults {
			t := &models.Task{
				PipelinerunID: uint(pipelinerunID),
				Application:   application,
				Cluster:       cluster,
				RegionID:      uint(regionID),
				Pipeline:      pipeline,
				Task:          trResult.Task,
				Result:        trResult.Result,
				StartedAt:     trResult.StartTime.Time,
				FinishedAt:    trResult.CompletionTime.Time,
				Duration:      uint(trResult.DurationSeconds),
			}
			result = tx.Create(t)
			if result.Error != nil {
				return errors.Wrap(ErrInsertPipeline, err.Error())
			}
		}

		for _, stepResult := range stepResults {
			s := &models.Step{
				PipelinerunID: uint(pipelinerunID),
				Application:   application,
				Cluster:       cluster,
				RegionID:      uint(regionID),
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
				return errors.Wrap(ErrInsertPipeline, err.Error())
			}
		}

		return nil
	})

	return err
}

func NewDAO() DAO {
	return &dao{}
}
