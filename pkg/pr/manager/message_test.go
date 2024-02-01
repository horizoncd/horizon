package manager

import (
	"context"
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/stretchr/testify/assert"

	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/pr/models"
)

func TestMessage(t *testing.T) {
	prMessageManager := NewPRMessageManager(db)

	prMessage := &models.PRMessage{
		PipelineRunID: 1,
		Content:       "message 1",
		MessageType:   models.MessageTypeSystem,
	}
	createdPRMessage, err := prMessageManager.Create(context.Background(), prMessage)
	assert.NoError(t, err)
	assert.NotNil(t, createdPRMessage)
	assert.Equal(t, prMessage.PipelineRunID, createdPRMessage.PipelineRunID)
	assert.Equal(t, prMessage.Content, createdPRMessage.Content)
	assert.Equal(t, prMessage.MessageType, createdPRMessage.MessageType)

	// Create some PRMessages to test with
	prMessages := []*models.PRMessage{
		{PipelineRunID: 2, Content: "user message 1", MessageType: models.MessageTypeUser},
		{PipelineRunID: 2, Content: "user message 2", MessageType: models.MessageTypeUser},
		{PipelineRunID: 2, Content: "system message 1", MessageType: models.MessageTypeSystem},
		{PipelineRunID: 3, Content: "user message 3", MessageType: models.MessageTypeUser},
	}
	for _, message := range prMessages {
		_, err := prMessageManager.Create(context.Background(), message)
		assert.NoError(t, err)
	}
	// Test listing PRMessages for a specific PipelineRunID
	totalCount, listedPRMessages, err := prMessageManager.List(context.Background(), 2, nil)
	assert.NoError(t, err)
	assert.Len(t, listedPRMessages, 3)
	assert.Equal(t, 3, totalCount)

	// Test listing PRMessages with page query
	query := &q.Query{PageSize: 1}
	totalCount, listedPRMessages, err = prMessageManager.List(context.Background(), 2, query)
	assert.NoError(t, err)
	assert.Len(t, listedPRMessages, 1)
	assert.Equal(t, 3, totalCount)

	// Test listing PRMessages with system query
	query = &q.Query{Keywords: map[string]interface{}{common.MessageQueryBySystem: true}}
	totalCount, listedPRMessages, err = prMessageManager.List(context.Background(), 2, query)
	assert.NoError(t, err)
	assert.Len(t, listedPRMessages, 1)
	for m := range listedPRMessages {
		assert.Equal(t, listedPRMessages[m].MessageType, uint(models.MessageTypeSystem))
	}
	assert.Equal(t, 1, totalCount)

	// Test listing PRMessages with system query
	query = &q.Query{Keywords: map[string]interface{}{common.MessageQueryBySystem: false}}
	totalCount, listedPRMessages, err = prMessageManager.List(context.Background(), 2, query)
	assert.NoError(t, err)
	assert.Len(t, listedPRMessages, 2)
	for m := range listedPRMessages {
		assert.Equal(t, listedPRMessages[m].MessageType, uint(models.MessageTypeUser))
	}
	assert.Equal(t, 2, totalCount)
}
