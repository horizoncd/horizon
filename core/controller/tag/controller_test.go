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

package tag

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/middleware/requestid"
	"github.com/horizoncd/horizon/lib/orm"
	clustergitrepomock "github.com/horizoncd/horizon/mock/pkg/cluster/gitrepo"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	templatemodels "github.com/horizoncd/horizon/pkg/template/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
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
	if err := db.AutoMigrate(&appmodels.Application{}, &models.Cluster{},
		&tagmodels.Tag{}, &membermodels.Member{}, &regionmodels.Region{},
		&trmodels.TemplateRelease{}, &templatemodels.Template{}); err != nil {
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
	mockCtl := gomock.NewController(t)
	clusterGitRepo := clustergitrepomock.NewMockClusterGitRepo(mockCtl)
	clusterGitRepo.EXPECT().UpdateTags(ctx, gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	appMgr := manager.ApplicationMgr
	clusterMgr := manager.ClusterMgr
	templateReleaseMgr := manager.TemplateReleaseMgr
	regionMgr := manager.RegionMgr

	// init data
	region, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name: "test",
	})
	assert.Nil(t, err)

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

	cluster, err := clusterMgr.Create(ctx, &models.Cluster{
		ApplicationID:   application.ID,
		Name:            "cluster",
		Template:        "javaapp",
		TemplateRelease: "v1.2.0",
		RegionName:      region.Name,
	}, nil, nil)
	assert.Nil(t, err)

	_, err = templateReleaseMgr.Create(ctx, &trmodels.TemplateRelease{
		Template:     1,
		TemplateName: "javaapp",
		ChartVersion: "v1.2.0",
		ChartName:    "0-javaapp",
	})
	assert.Nil(t, err)

	c = &controller{
		clusterMgr:     clusterMgr,
		tagMgr:         manager.TagMgr,
		clusterGitRepo: clusterGitRepo,
		applicationMgr: appMgr,
	}

	clusterID := cluster.ID
	err = c.Update(ctx, common.ResourceCluster, clusterID, &UpdateRequest{
		Tags: []*tagmodels.TagBasic{
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

	resp, err := c.List(ctx, common.ResourceCluster, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Tags))
	assert.Equal(t, "a", resp.Tags[0].Key)
	assert.Equal(t, "1", resp.Tags[0].Value)
	assert.Equal(t, "b", resp.Tags[1].Key)
	assert.Equal(t, "2", resp.Tags[1].Value)

	err = c.Update(ctx, common.ResourceCluster, clusterID, &UpdateRequest{
		Tags: []*tagmodels.TagBasic{
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

	resp, err = c.List(ctx, common.ResourceCluster, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Tags))
	assert.Equal(t, "a", resp.Tags[0].Key)
	assert.Equal(t, "1", resp.Tags[0].Value)
	assert.Equal(t, "c", resp.Tags[1].Key)
	assert.Equal(t, "3", resp.Tags[1].Value)

	err = c.Update(ctx, common.ResourceCluster, clusterID, &UpdateRequest{
		Tags: []*tagmodels.TagBasic{
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

	resp, err = c.List(ctx, common.ResourceCluster, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(resp.Tags))
	assert.Equal(t, "a", resp.Tags[0].Key)
	assert.Equal(t, "1", resp.Tags[0].Value)
	assert.Equal(t, "c", resp.Tags[1].Key)
	assert.Equal(t, "3", resp.Tags[1].Value)
	assert.Equal(t, "d", resp.Tags[2].Key)
	assert.Equal(t, "4", resp.Tags[2].Value)

	err = c.Update(ctx, common.ResourceCluster, clusterID, &UpdateRequest{
		Tags: []*tagmodels.TagBasic{
			{
				Key:   "d",
				Value: "4",
			},
		},
	})
	assert.Nil(t, err)

	resp, err = c.List(ctx, common.ResourceCluster, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp.Tags))
	assert.Equal(t, "d", resp.Tags[0].Key)
	assert.Equal(t, "4", resp.Tags[0].Value)

	err = c.Update(ctx, common.ResourceCluster, clusterID, &UpdateRequest{
		Tags: []*tagmodels.TagBasic{},
	})
	assert.Nil(t, err)

	resp, err = c.List(ctx, common.ResourceCluster, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp.Tags))

	tags := make([]*tagmodels.TagBasic, 0)
	for i := 0; i < 21; i++ {
		tags = append(tags, &tagmodels.TagBasic{
			Key:   strconv.Itoa(i),
			Value: strconv.Itoa(i),
		})
	}

	var request = &UpdateRequest{
		Tags: tags,
	}
	err = c.Update(ctx, common.ResourceCluster, clusterID, request)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())

	cluster2, err := clusterMgr.Create(ctx, &models.Cluster{
		ApplicationID: application.ID,
		Name:          "cluster2",
		RegionName:    region.Name,
	}, nil, nil)
	assert.Nil(t, err)

	err = c.Update(ctx, common.ResourceCluster, cluster2.ID, &UpdateRequest{
		Tags: []*tagmodels.TagBasic{
			{
				Key:   "d",
				Value: "4",
			},
			{
				Key:   "e",
				Value: "5",
			},
		},
	})
	assert.Nil(t, err)

	resp, err = c.ListSubResourceTags(ctx, common.ResourceApplication, cluster.ApplicationID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Tags))
}
