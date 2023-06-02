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

package manager

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/pipelinerun/models"

	"github.com/stretchr/testify/assert"
)

var (
	db, _ = orm.NewSqliteDB("")
	ctx   context.Context
	mgr   = New(db)
)

func Test(t *testing.T) {
	pr, err := mgr.Create(ctx, &models.Pipelinerun{
		ID:               0,
		ClusterID:        1,
		Action:           models.ActionBuildDeploy,
		Status:           "created",
		Title:            "title",
		Description:      "description",
		GitURL:           "",
		GitRefType:       codemodels.GitRefTypeBranch,
		GitRef:           "",
		GitCommit:        "",
		ImageURL:         "",
		LastConfigCommit: "",
		ConfigCommit:     "1",
		S3Bucket:         "",
		LogObject:        "",
		PrObject:         "",
		CreatedBy:        0,
	})
	assert.Nil(t, err)
	t.Logf("%v", pr)

	prGet, err := mgr.GetByID(ctx, pr.ID)
	assert.Nil(t, err)
	assert.Equal(t, "title", prGet.Title)
	assert.Equal(t, "1", prGet.ConfigCommit)

	err = mgr.UpdateConfigCommitByID(ctx, prGet.ID, "2")
	assert.Nil(t, err)

	prGet, err = mgr.GetByID(ctx, pr.ID)
	assert.Nil(t, err)
	assert.Equal(t, "2", prGet.ConfigCommit)

	prGet, err = mgr.GetLatestByClusterIDAndActions(ctx, pr.ClusterID, models.ActionBuildDeploy)
	assert.Nil(t, err)
	assert.Equal(t, "2", prGet.ConfigCommit)

	err = mgr.UpdateStatusByID(ctx, pr.ID, models.StatusMerged)
	assert.Nil(t, err)
	prGet, err = mgr.GetByID(ctx, pr.ID)
	assert.Nil(t, err)
	assert.Equal(t, prGet.Status, string(models.StatusMerged))

	err = mgr.UpdateResultByID(ctx, pr.ID, &models.Result{
		S3Bucket:   "bucket",
		LogObject:  "log-obj",
		PrObject:   "pr-obj",
		Result:     "ok",
		StartedAt:  func() *time.Time { t := time.Now(); return &t }(),
		FinishedAt: func() *time.Time { t := time.Now(); return &t }(),
	})
	assert.Nil(t, err)

	prGet, err = mgr.GetByID(ctx, pr.ID)
	assert.Nil(t, err)
	assert.Equal(t, prGet.S3Bucket, "bucket")
	assert.Equal(t, prGet.LogObject, "log-obj")
	assert.Equal(t, prGet.PrObject, "pr-obj")
	assert.Equal(t, prGet.Status, "ok")
	b, _ := json.Marshal(prGet)
	t.Logf("%v", string(b))

	err = mgr.DeleteByID(ctx, pr.ID)
	assert.Nil(t, err)

	prGet, err = mgr.GetByID(ctx, pr.ID)
	assert.Nil(t, err)
	assert.Nil(t, prGet)
}
func TestGetByClusterID(t *testing.T) {
	var clusterID uint = 1
	pr := &models.Pipelinerun{
		ID:          0,
		ClusterID:   clusterID,
		Action:      models.ActionBuildDeploy,
		Status:      "created",
		Title:       "title",
		Description: "description",
		CreatedBy:   0,
	}
	_, err := mgr.Create(ctx, pr)
	assert.Nil(t, err)

	pr.ID = 2
	_, err = mgr.Create(ctx, pr)
	assert.Nil(t, err)

	pr.ID = 3
	_, err = mgr.Create(ctx, pr)
	assert.Nil(t, err)

	var PageSize = 2
	var PageNumber = 1
	query := q.Query{
		PageNumber: PageNumber,
		PageSize:   PageSize,
	}
	totalCount, pipelineruns, err := mgr.GetByClusterID(ctx, clusterID, false, query)
	assert.Nil(t, err)
	assert.Equal(t, totalCount, 3)
	assert.Equal(t, len(pipelineruns), PageSize)
	body, _ := json.MarshalIndent(pipelineruns, "", " ")
	t.Logf("%s", string(body))

	pr.ID = 4
	pr.Status = "ok"
	_, err = mgr.Create(ctx, pr)
	assert.Nil(t, err)

	totalCount, pipelineruns, err = mgr.GetByClusterID(ctx, clusterID, true, query)
	assert.Nil(t, err)
	assert.Equal(t, 0, totalCount)
	assert.Equal(t, 0, len(pipelineruns))
}

// nolint
func TestGetLatestSuccessByClusterID(t *testing.T) {
	var clusterID uint = 1
	pr := &models.Pipelinerun{
		ID:          5,
		ClusterID:   clusterID,
		Action:      models.ActionBuildDeploy,
		Status:      "ok",
		Title:       "title",
		Description: "description",
		CreatedBy:   0,
		GitCommit:   "xxxxxx",
		UpdatedAt:   time.Now(),
	}
	_, err := mgr.Create(ctx, pr)
	assert.Nil(t, err)

	pr.ID = 6
	pr.UpdatedAt = time.Now()
	_, err = mgr.Create(ctx, pr)
	assert.Nil(t, err)

	pr.ID = 7
	pr.UpdatedAt = time.Now()
	pr.Action = models.ActionRollback
	_, err = mgr.Create(ctx, pr)
	assert.Nil(t, err)

	pr.ID = 8
	pr.UpdatedAt = time.Now()
	pr.Action = models.ActionBuildDeploy
	pr.Status = "created"
	_, err = mgr.Create(ctx, pr)
	assert.Nil(t, err)

	pipelinerun, err := mgr.GetLatestSuccessByClusterID(ctx, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, uint(7), pipelinerun.ID)
}

func TestGetFirstCanRollbackPipelinerun(t *testing.T) {
	var clusterID uint = 1
	pr := &models.Pipelinerun{
		ID:          10,
		ClusterID:   clusterID,
		Action:      models.ActionBuildDeploy,
		Status:      "ok",
		Title:       "title",
		Description: "description",
		CreatedBy:   0,
		GitCommit:   "xxxxxx",
		UpdatedAt:   time.Now(),
		CreatedAt:   time.Now(),
	}
	_, err := mgr.Create(ctx, pr)
	assert.Nil(t, err)

	pr = &models.Pipelinerun{
		ID:          11,
		ClusterID:   clusterID,
		Action:      models.ActionBuildDeploy,
		Status:      "ok",
		Title:       "title",
		Description: "description",
		CreatedBy:   0,
		GitCommit:   "xxxxxx",
		UpdatedAt:   time.Now(),
	}
	_, err = mgr.Create(ctx, pr)
	assert.Nil(t, err)

	pipelinerun, err := mgr.GetFirstCanRollbackPipelinerun(ctx, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, pipelinerun)
	assert.Equal(t, 11, int(pipelinerun.ID))
	t.Logf("%v", pipelinerun)

	pipelinerun, err = mgr.GetFirstCanRollbackPipelinerun(ctx, 10000)
	assert.Nil(t, err)
	assert.Nil(t, pipelinerun)
}

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&models.Pipelinerun{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	os.Exit(m.Run())
}
