package templateschematag

import (
	"os"
	"strconv"
	"testing"

	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	appmanager "g.hz.netease.com/horizon/pkg/application/manager"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	tagmanager "g.hz.netease.com/horizon/pkg/templateschematag/manager"
	templateschemamodels "g.hz.netease.com/horizon/pkg/templateschematag/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var (
	ctx context.Context
	c   Controller
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	if err := db.AutoMigrate(&appmodels.Application{}, &models.Cluster{},
		&templateschemamodels.ClusterTemplateSchemaTag{}, &membermodels.Member{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	ctx = context.WithValue(ctx, user.Key(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   uint(1),
	})
	ctx = context.WithValue(ctx, requestid.HeaderXRequestID, "requestid")

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	appMgr := appmanager.Mgr
	clusterMgr := clustermanager.Mgr

	// init data
	application, err := appMgr.Create(ctx, &appmodels.Application{
		GroupID:         uint(1),
		Name:            "app",
		Priority:        "P3",
		GitURL:          "ssh://git.com",
		GitSubfolder:    "/test",
		GitBranch:       "master",
		Template:        "javaapp",
		TemplateRelease: "v1.0.0",
	}, nil)
	assert.Nil(t, err)

	cluster, err := clusterMgr.Create(ctx, &models.Cluster{
		ApplicationID: application.ID,
		Name:          "cluster",
	}, nil, nil)
	assert.Nil(t, err)

	c = &controller{
		clusterMgr:          clusterMgr,
		clusterSchemaTagMgr: tagmanager.Mgr,
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
