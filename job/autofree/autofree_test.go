package autofree

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	usercommon "g.hz.netease.com/horizon/core/common"
	coreconfig "g.hz.netease.com/horizon/core/config"
	clusterctl "g.hz.netease.com/horizon/core/controller/cluster"
	environmentctl "g.hz.netease.com/horizon/core/controller/environment"
	prctl "g.hz.netease.com/horizon/core/controller/pipelinerun"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	cdmock "g.hz.netease.com/horizon/mock/pkg/cluster/cd"
	pipelinemockmanager "g.hz.netease.com/horizon/mock/pkg/pipelinerun/manager"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/config/autofree"
	emodel "g.hz.netease.com/horizon/pkg/environment/models"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	"g.hz.netease.com/horizon/pkg/errors"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	idpmodels "g.hz.netease.com/horizon/pkg/idp/models"
	"g.hz.netease.com/horizon/pkg/idp/utils"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	pipelinemodel "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	registrymodels "g.hz.netease.com/horizon/pkg/registry/models"
	"g.hz.netease.com/horizon/pkg/server/global"
	xrequestid "g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	tmodel "g.hz.netease.com/horizon/pkg/tag/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	tagmodel "g.hz.netease.com/horizon/pkg/templateschematag/models"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	linkmodels "g.hz.netease.com/horizon/pkg/userlink/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	db, _   = orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
	ctx     = context.Background()
)

const secondsInOneDay = 24 * 3600

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&appmodels.Application{}, &models.Cluster{}, &groupmodels.Group{},
		&trmodels.TemplateRelease{}, &membermodels.Member{}, &usermodels.User{},
		&registrymodels.Registry{}, &idpmodels.IdentityProvider{}, &linkmodels.UserLink{},
		&regionmodels.Region{}, &envregionmodels.EnvironmentRegion{},
		&prmodels.Pipelinerun{}, &tagmodel.ClusterTemplateSchemaTag{}, &tmodel.Tag{}, &emodel.Environment{}); err != nil {
		panic(err)
	}
	// nolint
	ctx = context.WithValue(ctx, usercommon.UserContextKey(), &userauth.DefaultInfo{
		Name: "horizon",
		ID:   uint(1),
	})
	// nolint
	ctx = context.WithValue(ctx, xrequestid.HeaderXRequestID, "requestid")
	os.Exit(m.Run())
}

func createUser(t *testing.T) {
	var (
		name  = "horizon"
		email = "horizon@noreply.com"
	)

	method := uint8(idpmodels.ClientSecretSentAsPost)
	idp, err := manager.IdpManager.Create(ctx, &idpmodels.IdentityProvider{
		Model:                   global.Model{ID: 1},
		Name:                    "netease",
		TokenEndpointAuthMethod: (*idpmodels.TokenEndpointAuthMethod)(&method),
	})
	assert.Nil(t, err)
	assert.Equal(t, uint(1), idp.ID)
	assert.Equal(t, "netease", idp.Name)

	mgr := manager.UserManager

	u, err := mgr.Create(ctx, &usermodels.User{
		Name:  name,
		Email: email,
	})
	assert.Nil(t, err)
	assert.NotNil(t, u)

	link, err := manager.UserLinksManager.CreateLink(ctx, u.ID, idp.ID, &utils.Claims{
		Sub:   "netease",
		Name:  name,
		Email: email,
	}, true)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), link.ID)
}

func TestAutoFreeExpiredCluster(t *testing.T) {
	mockCtl := gomock.NewController(t)
	cd := cdmock.NewMockCD(mockCtl)
	conf := &coreconfig.Config{}
	parameter := &param.Param{
		Manager: manager,
		Cd:      cd,
	}
	mockPipelineManager := pipelinemockmanager.NewMockManager(mockCtl)
	parameter.PipelinerunMgr = mockPipelineManager
	clrCtl := clusterctl.NewController(conf, parameter)
	envCtl := environmentctl.NewController(parameter)
	prCtl := prctl.NewController(parameter)

	// init data

	// User
	createUser(t)

	// environment
	devID, err := envCtl.Create(ctx, &environmentctl.CreateEnvironmentRequest{
		Name:        "dev",
		DisplayName: "DEV",
		AutoFree:    true,
	})
	assert.Nil(t, err)
	devEnv, err := envCtl.GetByID(ctx, devID)
	assert.Nil(t, err)
	assert.Equal(t, "DEV", devEnv.DisplayName)

	onlineID, err := envCtl.Create(ctx, &environmentctl.CreateEnvironmentRequest{
		Name:        "online",
		DisplayName: "ONLINE",
		AutoFree:    false,
	})
	assert.Nil(t, err)
	onlineEnv, err := envCtl.GetByID(ctx, onlineID)
	assert.Nil(t, err)
	assert.Equal(t, "ONLINE", onlineEnv.DisplayName)

	// ListClusterWithExpiry
	for i := 0; i < 7; i++ {
		name := "clusterWithExpiry" + strconv.Itoa(i)
		cluster := &clustermodels.Cluster{
			ApplicationID:   uint(1),
			Name:            name,
			EnvironmentName: devEnv.Name,
			RegionName:      "hzListClusterWithExpiry",
			GitURL:          "ssh://git@g.hz.netease.com:22222/music-cloud-native/horizon/horizon.git",
			Status:          "",
			ExpireSeconds:   uint((i + 1) * secondsInOneDay),
			Model: global.Model{
				UpdatedAt: time.Now().AddDate(0, 0, -i-2),
			},
		}
		if i == 3 {
			cluster.EnvironmentName = onlineEnv.Name
		}
		if i == 6 {
			cluster.UpdatedAt = time.Now()
		}
		cluster, err = manager.ClusterMgr.Create(ctx, cluster, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, cluster)

		// pipelinerun
		pipelineruns := make([]*pipelinemodel.Pipelinerun, 0)
		if i == 5 {
			pipelineruns = append(pipelineruns, &pipelinemodel.Pipelinerun{
				ID:        uint(i + 1),
				ClusterID: cluster.ID,
				CreatedBy: 1,
				UpdatedAt: time.Now(),
			})
		} else if i != 4 {
			pipelineruns = append(pipelineruns, &pipelinemodel.Pipelinerun{
				ID:        uint(i + 1),
				ClusterID: cluster.ID,
				CreatedBy: 1,
				UpdatedAt: time.Now().AddDate(0, 0, -i-2),
			})
		}
		if i != 4 {
			mockPipelineManager.EXPECT().GetByClusterID(gomock.Any(), cluster.ID, gomock.Any(), gomock.Any()).
				Return(1, pipelineruns, nil).AnyTimes()
		} else {
			mockPipelineManager.EXPECT().GetByClusterID(gomock.Any(), cluster.ID, gomock.Any(), gomock.Any()).
				Return(0, pipelineruns, nil).AnyTimes()
		}
		mockPipelineManager.EXPECT().GetFirstCanRollbackPipelinerun(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		num, pipelineBasics, err := prCtl.List(ctx, cluster.ID, false, q.Query{
			PageNumber: 1,
			PageSize:   10,
		})
		assert.Nil(t, err)
		t.Logf("%v", num)
		t.Logf("%v", pipelineBasics)
	}
	cd.EXPECT().DeleteCluster(gomock.Any(), gomock.Any()).Return(errors.New("test")).AnyTimes()
	ctx, cancelFunc := context.WithCancel(ctx)
	go func() {
		timer := time.NewTimer(time.Second * 5)
		<-timer.C
		cancelFunc()
	}()
	AutoReleaseExpiredClusterJob(ctx, &autofree.Config{
		AccountIDP:    "netease",
		Account:       "horizon@noreply.com",
		JobInterval:   1 * time.Second,
		BatchInterval: 0 * time.Second,
		BatchSize:     20,
		SupportedEnvs: []string{"dev"},
	}, manager.UserManager, clrCtl, prCtl, envCtl)
}
