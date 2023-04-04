package collector

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	prmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/stretchr/testify/assert"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	tektonmock "github.com/horizoncd/horizon/mock/pkg/cluster/tekton"
)

func TestDummyCollector(t *testing.T) {
	var pr *v1beta1.PipelineRun
	if err := json.Unmarshal([]byte(pipelineRunJSON), &pr); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	ctl := gomock.NewController(t)
	tek := tektonmock.NewMockInterface(ctl)
	tek.EXPECT().GetPipelineRunLogByID(ctx, gomock.Any()).Return(getPipelineRunLog(pr))
	tek.EXPECT().GetPipelineRunByID(ctx, gomock.Any()).Return(pr, nil)

	c := NewDummyCollector(tek)

	businessData := &global.HorizonMetaData{
		Application: "app",
		Cluster:     "cluster",
		Environment: "test",
	}

	// collect
	collectResult, err := c.Collect(ctx, pr, businessData)
	assert.Nil(t, err)
	b, _ := json.Marshal(collectResult)
	t.Logf("%v", string(b))

	// 1. getLatestPipelineRunLog
	prModel := &prmodels.Pipelinerun{
		CIEventID: "cttzw",
	}
	_, err = c.GetPipelineRunLog(ctx, prModel)
	assert.Nil(t, err)

	// 2. getLatestPipelineRunObject
	obj, err := c.GetPipelineRunObject(ctx, collectResult.PrObject)
	assert.Nil(t, err)
	assert.Nil(t, obj)

	// 3. getLatestPipelineRun
	_, err = c.GetPipelineRun(ctx, prModel)
	assert.Nil(t, err)
}
