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

package templateschematag

import (
	"os"
	"strconv"
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/middleware/requestid"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	appmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var (
	ctx     context.Context
	c       Controller
	manager *managerparam.Manager
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
	if err := db.AutoMigrate(&appmodels.Application{}, &appmodels.Cluster{},
		&appmodels.ClusterTemplateSchemaTag{}, &appmodels.Member{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   uint(1),
	})
	ctx = context.WithValue(ctx, requestid.HeaderXRequestID, "requestid")

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	appMgr := manager.ApplicationManager
	clusterMgr := manager.ClusterMgr

	// init data
	application, err := appMgr.Create(ctx, &appmodels.Application{
		GroupID:         uint(1),
		Name:            "app",
		Priority:        "P3",
		GitURL:          "ssh://git.com",
		GitSubfolder:    "/test",
		GitRef:          "master",
		Template:        "javaapp",
		TemplateRelease: "v1.0.0",
	}, nil)
	assert.Nil(t, err)

	cluster, err := clusterMgr.Create(ctx, &appmodels.Cluster{
		ApplicationID: application.ID,
		Name:          "cluster",
	}, nil, nil)
	assert.Nil(t, err)

	c = &controller{
		clusterMgr:          clusterMgr,
		clusterSchemaTagMgr: manager.ClusterSchemaTagMgr,
	}

	clusterID := cluster.ID
	err = c.Update(ctx, clusterID, &UpdateRequest{
		Tags: []*Tag{
			{
				Key:   "a",
				Value: "1",
			}, {
				Key:   "b",
				Value: "2",
			},
		},
	})
	assert.Nil(t, err)

	resp, err := c.List(ctx, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Tags))
	assert.Equal(t, "a", resp.Tags[0].Key)
	assert.Equal(t, "1", resp.Tags[0].Value)
	assert.Equal(t, "b", resp.Tags[1].Key)
	assert.Equal(t, "2", resp.Tags[1].Value)

	err = c.Update(ctx, clusterID, &UpdateRequest{
		Tags: []*Tag{
			{
				Key:   "a",
				Value: "1",
			}, {
				Key:   "c",
				Value: "3",
			},
		},
	})
	assert.Nil(t, err)

	resp, err = c.List(ctx, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Tags))
	assert.Equal(t, "a", resp.Tags[0].Key)
	assert.Equal(t, "1", resp.Tags[0].Value)
	assert.Equal(t, "c", resp.Tags[1].Key)
	assert.Equal(t, "3", resp.Tags[1].Value)

	err = c.Update(ctx, clusterID, &UpdateRequest{
		Tags: []*Tag{
			{
				Key:   "a",
				Value: "1",
			}, {
				Key:   "c",
				Value: "3",
			}, {
				Key:   "d",
				Value: "4",
			},
		},
	})
	assert.Nil(t, err)

	resp, err = c.List(ctx, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(resp.Tags))
	assert.Equal(t, "a", resp.Tags[0].Key)
	assert.Equal(t, "1", resp.Tags[0].Value)
	assert.Equal(t, "c", resp.Tags[1].Key)
	assert.Equal(t, "3", resp.Tags[1].Value)
	assert.Equal(t, "d", resp.Tags[2].Key)
	assert.Equal(t, "4", resp.Tags[2].Value)

	err = c.Update(ctx, clusterID, &UpdateRequest{
		Tags: []*Tag{
			{
				Key:   "d",
				Value: "4",
			},
		},
	})
	assert.Nil(t, err)

	resp, err = c.List(ctx, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp.Tags))
	assert.Equal(t, "d", resp.Tags[0].Key)
	assert.Equal(t, "4", resp.Tags[0].Value)

	err = c.Update(ctx, clusterID, &UpdateRequest{
		Tags: []*Tag{},
	})
	assert.Nil(t, err)

	resp, err = c.List(ctx, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp.Tags))

	tags := make([]*Tag, 0)
	for i := 0; i < 21; i++ {
		tags = append(tags, &Tag{
			Key:   strconv.Itoa(i),
			Value: strconv.Itoa(i),
		})
	}

	var request = &UpdateRequest{
		Tags: tags,
	}
	err = c.Update(ctx, clusterID, request)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())
}
