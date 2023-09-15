package manager

import (
	"context"
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/pr/models"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	checkManager := NewCheckManager(db)

	ctx := context.Background()

	// Test Create
	check := &models.Check{
		Resource: common.Resource{
			Type:       "clusters",
			ResourceID: 123,
		},
	}
	createdCheck, err := checkManager.Create(ctx, check)
	assert.NoError(t, err)
	assert.Equal(t, createdCheck.Type, check.Type)

	// Test UpdateByID
	checkRun := &models.CheckRun{
		PipelineRunID: 10,
		CheckID:       createdCheck.ResourceID,
		Status:        models.CheckStatusQueue,
	}
	_, err = checkManager.CreateCheckRun(ctx, checkRun)
	assert.NoError(t, err)

	newCheckRun := &models.CheckRun{
		Status: models.CheckStatusInProgress,
	}
	err = checkManager.UpdateByID(ctx, checkRun.ID, newCheckRun)
	assert.NoError(t, err)

	// Test GetByResource
	checks, err := checkManager.GetByResource(ctx, common.Resource{
		Type:       "clusters",
		ResourceID: 123,
	})
	assert.NoError(t, err)
	assert.Len(t, checks, 1)

	// Test ListCheckRuns
	keyWords := make(map[string]interface{})
	keyWords[common.CheckrunQueryByPipelinerunID] = checkRun.PipelineRunID
	query := q.New(keyWords)
	checkRuns, err := checkManager.ListCheckRuns(ctx, query)
	assert.NoError(t, err)
	assert.Len(t, checkRuns, 1)
	assert.Equal(t, checkRuns[0].Status, models.CheckStatusInProgress)
	assert.Equal(t, checkRuns[0].CheckID, createdCheck.ResourceID)
}
