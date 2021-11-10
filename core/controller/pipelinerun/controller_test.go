package pipelinerun

import (
	"context"
	"os"
	"strconv"
	"testing"

	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	tektonmock "g.hz.netease.com/horizon/mock/pkg/cluster/tekton"
	tektoncollectormock "g.hz.netease.com/horizon/mock/pkg/cluster/tekton/collector"
	tektonftymock "g.hz.netease.com/horizon/mock/pkg/cluster/tekton/factory"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/log"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	prmanager "g.hz.netease.com/horizon/pkg/pipelinerun/manager"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	ctx context.Context
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Cluster{}, &membermodels.Member{},
		&envmodels.EnvironmentRegion{}, &prmodels.Pipelinerun{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&groupmodels.Group{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	ctx = context.WithValue(ctx, user.Key(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   uint(1),
	})

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tektonFty := tektonftymock.NewMockFactory(mockCtl)
	tekton := tektonmock.NewMockInterface(mockCtl)
	tektonCollector := tektoncollectormock.NewMockInterface(mockCtl)
	tektonFty.EXPECT().GetTekton(gomock.Any()).Return(tekton, nil).AnyTimes()
	tektonFty.EXPECT().GetTektonCollector(gomock.Any()).Return(tektonCollector, nil).AnyTimes()

	envMgr := envmanager.Mgr
	er, err := envMgr.CreateEnvironmentRegion(ctx, &envmodels.EnvironmentRegion{
		EnvironmentName: "test",
		RegionName:      "hz",
	})
	assert.Nil(t, err)
	assert.NotNil(t, er)

	clusterMgr := clustermanager.Mgr
	cluster, err := clusterMgr.Create(ctx, &models.Cluster{
		Name:                "cluster",
		EnvironmentRegionID: er.ID,
		CreatedBy:           0,
		UpdatedBy:           0,
	})
	assert.Nil(t, err)
	assert.NotNil(t, cluster)

	pipelinerunMgr := prmanager.Mgr
	pipelinerun, err := pipelinerunMgr.Create(ctx, &prmodels.Pipelinerun{
		ClusterID: cluster.ID,
		Action:    "builddeploy",
		Status:    "ok",
		S3Bucket:  "bucket",
		LogObject: "logObject",
		PrObject:  "prObject",
		CreatedBy: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, pipelinerun)

	c := &controller{
		prMgr:      pipelinerunMgr,
		clusterMgr: clusterMgr,
		envMgr:     envMgr,
		tektonFty:  tektonFty,
	}

	logBytes := []byte("this is a log")
	tektonCollector.EXPECT().GetPipelineRunLog(ctx, gomock.Any()).Return(logBytes, nil).AnyTimes()

	logCh := make(chan log.Log)
	errCh := make(chan error)
	tekton.EXPECT().GetPipelineRunLogByID(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(logCh, errCh, nil)

	l, err := c.GetPrLog(ctx, pipelinerun.ID)
	assert.Nil(t, err)
	assert.Nil(t, l.LogChannel)
	assert.Nil(t, l.ErrChannel)
	assert.Equal(t, l.LogBytes, logBytes)
	t.Logf("logBytes: %v", string(l.LogBytes))

	pipelinerun, err = pipelinerunMgr.Create(ctx, &prmodels.Pipelinerun{
		ClusterID: cluster.ID,
		Action:    "builddeploy",
		Status:    "ok",
		S3Bucket:  "",
		LogObject: "",
		PrObject:  "",
		CreatedBy: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, pipelinerun)

	go func() {
		defer close(logCh)
		defer close(errCh)
		for i := 0; i < 10; i++ {
			logCh <- log.Log{
				Pipeline: "default",
				Task:     "task",
				Step:     "step",
				Log:      strconv.Itoa(i),
			}
		}
	}()

	l, err = c.GetClusterLatestLog(ctx, cluster.ID)
	assert.Nil(t, err)
	assert.Nil(t, l.LogBytes)
	logC := l.LogChannel
	errC := l.ErrChannel
	for logC != nil || errC != nil {
		select {
		case l, ok := <-logC:
			if !ok {
				logC = nil
				continue
			}
			if l.Log == "EOFLOG" {
				t.Logf("\n")
				continue
			}
			t.Logf("[%s : %s] %s\n", l.Task, l.Step, l.Log)
		case e, ok := <-errC:
			if !ok {
				errC = nil
				continue
			}
			t.Logf("%s\n", e)
		}
	}
}
