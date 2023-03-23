package jobautofree

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	usercommon "github.com/horizoncd/horizon/core/common"
	coreconfig "github.com/horizoncd/horizon/core/config"
	clusterctl "github.com/horizoncd/horizon/core/controller/cluster"
	prctl "github.com/horizoncd/horizon/core/controller/pipelinerun"
	xrequestid "github.com/horizoncd/horizon/core/middleware/requestid"
	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	cdmock "github.com/horizoncd/horizon/mock/pkg/cluster/cd"
	pipelinemockmanager "github.com/horizoncd/horizon/mock/pkg/pipelinerun/manager"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
	"github.com/horizoncd/horizon/pkg/config/autofree"
	emodel "github.com/horizoncd/horizon/pkg/environment/models"
	"github.com/horizoncd/horizon/pkg/environment/service"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	"github.com/horizoncd/horizon/pkg/errors"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	idpmodels "github.com/horizoncd/horizon/pkg/idp/models"
	"github.com/horizoncd/horizon/pkg/idp/utils"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	pipelinemodel "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	prmodels "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	registrymodels "github.com/horizoncd/horizon/pkg/registry/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	tmodel "github.com/horizoncd/horizon/pkg/tag/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
	tagmodel "github.com/horizoncd/horizon/pkg/templateschematag/models"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	linkmodels "github.com/horizoncd/horizon/pkg/userlink/models"
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
		AutoFreeSvc: service.New([]string{"dev"}),
		Manager:     manager,
		Cd:          cd,
	}
	mockPipelineManager := pipelinemockmanager.NewMockManager(mockCtl)
	parameter.PipelinerunMgr = mockPipelineManager
	clrCtl := clusterctl.NewController(conf, parameter)
	prCtl := prctl.NewController(parameter)

	// init data

	// User
	createUser(t)

	// ListClusterWithExpiry
	for i := 0; i < 7; i++ {
		name := "clusterWithExpiry" + strconv.Itoa(i)
		cluster := &clustermodels.Cluster{
			ApplicationID:   uint(1),
			Name:            name,
			EnvironmentName: "dev",
			RegionName:      "hzListClusterWithExpiry",
			GitURL:          "ssh://git@cloudnative.com:22222/music-cloud-native/horizon/horizon.git",
			Status:          "",
			ExpireSeconds:   uint((i + 1) * secondsInOneDay),
			Model: global.Model{
				UpdatedAt: time.Now().AddDate(0, 0, -i-2),
			},
		}
		if i == 3 {
			cluster.EnvironmentName = "online"
		}
		if i == 6 {
			cluster.UpdatedAt = time.Now()
		}
		cluster, err := manager.ClusterMgr.Create(ctx, cluster, nil, nil)
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
	Run(ctx, &autofree.Config{
		AccountID:     1,
		JobInterval:   1 * time.Second,
		BatchInterval: 0 * time.Second,
		BatchSize:     20,
		SupportedEnvs: []string{"dev"},
	}, manager.UserManager, clrCtl, prCtl)
}
