package manager

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	harbordao "g.hz.netease.com/horizon/pkg/harbor/dao"
	harbormodels "g.hz.netease.com/horizon/pkg/harbor/models"
	k8sclustermanager "g.hz.netease.com/horizon/pkg/k8scluster/manager"
	k8sclustermodels "g.hz.netease.com/horizon/pkg/k8scluster/models"
	"g.hz.netease.com/horizon/pkg/region/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func Test(t *testing.T) {
	k8sClusterMgr := k8sclustermanager.Mgr
	hzCluster, err := k8sClusterMgr.Create(ctx, &k8sclustermodels.K8SCluster{
		Name:         "hz",
		Certificate:  "hz-cert",
		DomainSuffix: "hz.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, hzCluster)

	jdCluster, err := k8sClusterMgr.Create(ctx, &k8sclustermodels.K8SCluster{
		Name:         "jd",
		Certificate:  "jd-cert",
		DomainSuffix: "jd.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, jdCluster)

	harborDAO := harbordao.NewDAO()
	harbor, err := harborDAO.Create(ctx, &harbormodels.Harbor{
		Server:          "https://harbor1",
		Token:           "asdf",
		PreheatPolicyID: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, harbor)

	hzRegion, err := Mgr.Create(ctx, &models.Region{
		Name:         "hz",
		DisplayName:  "HZ",
		K8SClusterID: hzCluster.ID,
		HarborID:     harbor.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, hzRegion)

	jdRegion, err := Mgr.Create(ctx, &models.Region{
		Name:         "jd",
		DisplayName:  "JD",
		K8SClusterID: jdCluster.ID,
		HarborID:     harbor.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, jdRegion)

	regions, err := Mgr.ListAll(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, regions)
	assert.Equal(t, 2, len(regions))
	assert.Equal(t, "hz", regions[0].Name)
	assert.Equal(t, uint(1), regions[0].K8SClusterID)
	assert.Equal(t, "jd", regions[1].Name)
	assert.Equal(t, uint(2), regions[1].K8SClusterID)

	regionEntities, err := Mgr.ListRegionEntities(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(regionEntities))
	assert.Equal(t, hzCluster.Certificate, regionEntities[0].K8SCluster.Certificate)
	assert.Equal(t, jdCluster.Certificate, regionEntities[1].K8SCluster.Certificate)

	hzRegionEntity, err := Mgr.GetRegionEntity(ctx, "hz")
	assert.Nil(t, err)
	assert.NotNil(t, hzRegionEntity)
	assert.Equal(t, hzRegionEntity.K8SCluster.Certificate, hzCluster.Certificate)
	assert.Equal(t, hzRegionEntity.Harbor.Server, harbor.Server)
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&k8sclustermodels.K8SCluster{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&models.Region{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&harbormodels.Harbor{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
