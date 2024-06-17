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

package s3

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	backend := s3mem.New()
	_ = backend.CreateBucket("bucket")
	faker := gofakes3.New(backend)
	ts := httptest.NewServer(faker.Server())
	defer ts.Close()

	params := &Params{
		AccessKey:        "accessKey",
		SecretKey:        "secretKey",
		Region:           "us-east-1",
		Endpoint:         ts.URL,
		Bucket:           "bucket",
		ContentType:      "text/plain",
		SkipVerify:       true,
		S3ForcePathStyle: true,
		Prefix:           "test",
		// LogLevel:         func() *aws.LogLevelType { l := aws.LogDebugWithHTTPBody; return &l }(),
	}

	d, err := NewDriver(*params)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	content := "abcdefghigklmnopqrstuvwxyz"
	if err := d.PutObject(ctx, "pr-log/20210714/1", bytes.NewReader([]byte(content)), nil); err != nil {
		t.Fatal(err)
	}

	b, err := d.GetObject(ctx, "pr-log/20210714/1")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", string(b))
	assert.Equal(t, content, string(b))

	if err := d.CopyObject(ctx, "pr-log/20210714/1", "pr-log/20210714/2"); err != nil {
		t.Fatal(err)
	}

	b, err = d.GetObject(ctx, "pr-log/20210714/1")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", string(b))
	assert.Equal(t, content, string(b))

	b, err = d.GetObject(ctx, "pr-log/20210714/2")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", string(b))
	assert.Equal(t, content, string(b))

	url, err := d.GetSignedObjectURL("pr-log/20210714/2", time.Hour*24*30)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, content, string(res))

	if err := d.DeleteObjects(ctx, "pr-log/20210714"); err != nil {
		t.Fatal(err)
	}

	l, err := d.ListObjects(ctx, "pr-log/20210714", 10)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, len(l))
}
