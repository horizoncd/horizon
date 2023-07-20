//nolint:lll
package metrics

import (
	"context"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	clustermodel "github.com/horizoncd/horizon/pkg/cluster/models"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	templatemodels "github.com/horizoncd/horizon/pkg/template/models"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
)

func createContext(t *testing.T) (context.Context, *gorm.DB, *managerparam.Manager) {
	var (
		ctx     = context.Background()
		db, err = orm.NewSqliteDB("")
	)
	assert.Nil(t, err)

	err = db.AutoMigrate(&clustermodel.Cluster{},
		&appmodels.Application{}, &groupmodels.Group{},
		&usermodels.User{}, &regionmodels.Region{},
		&membermodels.Member{}, &templatemodels.Template{}, &tagmodels.Tag{})
	assert.Nil(t, err)

	return ctx, db, managerparam.InitManager(db)
}

func Test(t *testing.T) {
	ctx, _, mgr := createContext(t)
	user, err := mgr.UserMgr.Create(ctx, &usermodels.User{
		Name:  "tony",
		Email: "tony@dummy.org",
	})
	assert.Nil(t, err)
	assert.NotNil(t, user)
	// nolint
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name:  "tony",
		Email: "tony@dummy.org",
	})

	group, err := mgr.GroupMgr.Create(ctx, &groupmodels.Group{
		Name: "group1",
	})
	assert.Nil(t, err)

	app, err := mgr.ApplicationMgr.Create(ctx, &appmodels.Application{
		Name:    "app1",
		GroupID: group.ID,
	}, nil)
	assert.Nil(t, err)

	region, err := mgr.RegionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	cluster1 := &clustermodel.Cluster{
		ApplicationID:   app.ID,
		EnvironmentName: "dev",
		RegionName:      region.Name,
		Name:            "cluster1",
		Description:     "",
		GitURL:          "ssh://git@github.com",
		GitSubfolder:    "/",
		GitRef:          "develop",
		Image:           "ubuntu:latest",
		Template:        "javaapp",
		TemplateRelease: "v1.1.0",
		CreatedBy:       user.ID,
		UpdatedBy:       user.ID,
	}

	tmp := *cluster1
	cluster2 := &tmp
	cluster2.Name = "cluster2"

	cluster1Tags := []*tagmodels.Tag{
		{Key: "name", Value: "cluster1"},
	}
	_, err = mgr.ClusterMgr.Create(ctx, cluster1, cluster1Tags, nil)
	assert.Nil(t, err)

	cluster2Tags := []*tagmodels.Tag{
		{Key: "name", Value: "cluster2"},
		{Key: "hello_world1", Value: "cluster2"},
		{Key: "hello/world2", Value: "cluster2"},
		{Key: "&(*&)(*", Value: "cluster2"},
	}
	_, err = mgr.ClusterMgr.Create(ctx, cluster2, cluster2Tags, nil)
	assert.Nil(t, err)

	collector := &Collector{mgr}

	strReader := strings.NewReader(`# HELP horizon_cluster_info A metric with a constant '1' value labeled by cluster, application, group, etc.
# TYPE horizon_cluster_info gauge
horizon_cluster_info{application="app1",cluster="cluster1",environment="dev",group="group1",region="hz",template="javaapp"} 1
horizon_cluster_info{application="app1",cluster="cluster2",environment="dev",group="group1",region="hz",template="javaapp"} 1
# HELP horizon_cluster_labels A metric with a constant '1' value labeled by cluster and tags
# TYPE horizon_cluster_labels gauge
horizon_cluster_labels{application="app1",cluster="cluster1",label_name="cluster1"} 1
horizon_cluster_labels{application="app1",cluster="cluster2",label_hello_world1="cluster2",label_hello_world2="cluster2",label_name="cluster2"} 1
`)
	err = testutil.CollectAndCompare(collector, strReader)
	assert.Nil(t, err)
}
