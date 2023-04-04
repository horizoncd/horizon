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
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	prmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/horizoncd/horizon/lib/s3"
	tektonmock "github.com/horizoncd/horizon/mock/pkg/cluster/tekton"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/log"
)

func getPipelineRunLog(pr *v1beta1.PipelineRun) (<-chan log.Log, <-chan error, error) {
	logCh := make(chan log.Log)
	pipeline := pr.ObjectMeta.Labels["tekton.dev/pipeline"]
	task := "test-task"
	step := "test-step"
	go func() {
		defer close(logCh)
		for i := 0; i < 10; i++ {
			logCh <- log.Log{
				Pipeline: pipeline,
				Task:     task,
				Step:     step,
				Log:      "line" + strconv.Itoa(i),
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	return logCh, nil, nil
}

func TestNewS3Collector(t *testing.T) {
	s3Driver := &s3.Driver{}
	ctl := gomock.NewController(t)
	tek := tektonmock.NewMockInterface(ctl)
	s := NewS3Collector(s3Driver, tek)
	assert.NotNil(t, s)
}

func TestS3Collector_Collect(t *testing.T) {
	var pr *v1beta1.PipelineRun
	if err := json.Unmarshal([]byte(pipelineRunJSON), &pr); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	ctl := gomock.NewController(t)
	tek := tektonmock.NewMockInterface(ctl)
	tek.EXPECT().GetPipelineRunLog(ctx, pr).Return(getPipelineRunLog(pr))
	tek.EXPECT().DeletePipelineRun(ctx, pr).Return(nil)

	backend := s3mem.New()
	_ = backend.CreateBucket("bucket")
	faker := gofakes3.New(backend)
	ts := httptest.NewServer(faker.Server())
	defer ts.Close()

	params := &s3.Params{
		AccessKey:        "accessKey",
		SecretKey:        "secretKey",
		Region:           "us-east-1",
		Endpoint:         ts.URL,
		Bucket:           "bucket",
		ContentType:      "text/plain",
		SkipVerify:       true,
		S3ForcePathStyle: true,
	}

	d, err := s3.NewDriver(*params)
	assert.Nil(t, err)

	c := NewS3Collector(d, tek)

	businessDatas := &global.HorizonMetaData{
		Application: "app",
		Cluster:     "cluster",
		Environment: "test",
	}

	// collect
	collectResult, err := c.Collect(ctx, pr, businessDatas)
	assert.Nil(t, err)
	b, _ := json.Marshal(collectResult)
	t.Logf("%v", string(b))

	// 1. getLatestPipelineRunLog
	prModel := &prmodels.Pipelinerun{
		LogObject: collectResult.LogObject,
		PrObject:  collectResult.PrObject,
		S3Bucket:  collectResult.Bucket,
	}
	_, err = c.GetPipelineRunLog(ctx, prModel)
	assert.Nil(t, err)

	// 2. getLatestPipelineRunObject
	obj, err := c.GetPipelineRunObject(ctx, collectResult.PrObject)
	assert.Nil(t, err)
	assert.NotNil(t, obj)
	objectMeta := resolveObjMetadata(pr, businessDatas)
	if !reflect.DeepEqual(objectMeta, obj.Metadata) {
		t.Fatalf("pipelineRun objectMeta: expected %v, got %v", objectMeta, obj.Metadata)
	}

	// 3. getLatestPipelineRun
	tektonPR, err := c.GetPipelineRun(ctx, prModel)
	assert.Nil(t, err)
	if !reflect.DeepEqual(tektonPR, pr) {
		t.Fatalf("pipelineRun objectMeta: expected %v, got %v", objectMeta, obj.Metadata)
	}
}
