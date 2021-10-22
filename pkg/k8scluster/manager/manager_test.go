package manager

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/k8scluster/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func Test(t *testing.T) {
	k8sClusters, err := Mgr.ListAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, len(k8sClusters), 0)

	k8sCluster, err := Mgr.Create(ctx, &models.K8SCluster{
		Name:          "hz",
		Certificate:   "i am a Certificate",
		IngressDomain: "hz.com",
	})
	assert.Nil(t, err)
	t.Logf("%v", k8sCluster)

	k8sClusters, err = Mgr.ListAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, len(k8sClusters), 1)
	t.Logf("%v", k8sClusters[0])
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.K8SCluster{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
