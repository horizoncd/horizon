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

package collector

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	prmodels "github.com/horizoncd/horizon/pkg/pr/models"
	"github.com/horizoncd/horizon/pkg/server/global"

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
