package tag

import (
	"context"
	"os"
	"strconv"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/requestid"
	"g.hz.netease.com/horizon/lib/orm"
	clustergitrepomock "g.hz.netease.com/horizon/mock/pkg/cluster/gitrepo"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
		&tagmodels.Tag{}, &membermodels.Member{},
		&trmodels.TemplateRelease{}); err != nil {
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

	appMgr := manager.ApplicationManager
	clusterMgr := manager.ClusterMgr
	templateReleaseMgr := manager.TemplateReleaseManager

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

	cluster, err := clusterMgr.Create(ctx, &models.Cluster{
		ApplicationID:   application.ID,
		Name:            "cluster",
		Template:        "javaapp",
		TemplateRelease: "v1.2.0",
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
		tagMgr:         manager.TagManager,
		clusterGitRepo: clusterGitRepo,
		applicationMgr: appMgr,
	}

	clusterID := cluster.ID
	err = c.Update(ctx, common.ResourceCluster, clusterID, &UpdateRequest{
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

	resp, err := c.List(ctx, common.ResourceCluster, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Tags))
	assert.Equal(t, "a", resp.Tags[0].Key)
	assert.Equal(t, "1", resp.Tags[0].Value)
	assert.Equal(t, "b", resp.Tags[1].Key)
	assert.Equal(t, "2", resp.Tags[1].Value)

	err = c.Update(ctx, common.ResourceCluster, clusterID, &UpdateRequest{
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

	resp, err = c.List(ctx, common.ResourceCluster, clusterID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Tags))
	assert.Equal(t, "a", resp.Tags[0].Key)
	assert.Equal(t, "1", resp.Tags[0].Value)
	assert.Equal(t, "c", resp.Tags[1].Key)
	assert.Equal(t, "3", resp.Tags[1].Value)

	err = c.Update(ctx, common.ResourceCluster, clusterID, &UpdateRequest{
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
		Tags: []*Tag{
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
		Tags: []*Tag{},
	})
	assert.Nil(t, err)

	resp, err = c.List(ctx, common.ResourceCluster, clusterID)
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
	err = c.Update(ctx, common.ResourceCluster, clusterID, request)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())

	cluster2, err := clusterMgr.Create(ctx, &models.Cluster{
		ApplicationID: application.ID,
		Name:          "cluster2",
	}, nil, nil)
	assert.Nil(t, err)

	err = c.Update(ctx, common.ResourceCluster, cluster2.ID, &UpdateRequest{
		Tags: []*Tag{
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
