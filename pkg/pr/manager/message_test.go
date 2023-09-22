package manager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/pr/models"
)

func TestMessage(t *testing.T) {
	prMessageManager := NewPRMessageManager(db)

	// Create some PRMessages to test with
	prMessages := []*models.PRMessage{
		{PipelineRunID: 1, Content: "message 2"},
		{PipelineRunID: 2, Content: "message 3"},
	}
	for _, prMessage := range prMessages {
		_, err := prMessageManager.Create(context.Background(), prMessage)
		assert.NoError(t, err)
	}
	prMessage := &models.PRMessage{PipelineRunID: 1, Content: "message 1"}

	createdPRMessage, err := prMessageManager.Create(context.Background(), prMessage)
	assert.NoError(t, err)
	assert.NotNil(t, createdPRMessage)
	assert.Equal(t, prMessage.PipelineRunID, createdPRMessage.PipelineRunID)
	assert.Equal(t, prMessage.Content, createdPRMessage.Content)

	// Test listing PRMessages for a specific PipelineRunID
	totalCount, listedPRMessages, err := prMessageManager.List(context.Background(), 1, nil)
	assert.NoError(t, err)
	assert.Len(t, listedPRMessages, 2)
	assert.Equal(t, 2, totalCount)

	// Test listing PRMessages with a query
	query := &q.Query{PageSize: 1}
	totalCount, listedPRMessages, err = prMessageManager.List(context.Background(), 1, query)
	assert.NoError(t, err)
	assert.Len(t, listedPRMessages, 1)
	assert.Equal(t, 2, totalCount)
}
