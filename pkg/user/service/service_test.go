package service

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/user/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.User{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}

func Test_service_CheckUsersExists(t *testing.T) {
	userManger := usermanager.Mgr
	_, err := userManger.Create(ctx, &models.User{
		Name:  "tony",
		Email: "tony@corp.com",
	})
	assert.Nil(t, err)

	_, err = userManger.Create(ctx, &models.User{
		Name:  "mary",
		Email: "mary@corp.com",
	})
	assert.Nil(t, err)

	svc := NewService()
	err = svc.CheckUsersExists(ctx, []string{"tony@corp.com"})
	assert.Nil(t, err)

	err = svc.CheckUsersExists(ctx, []string{"tony@corp.com", "mary@corp.com"})
	assert.Nil(t, err)

	err = svc.CheckUsersExists(ctx, []string{"tony@corp.com", "not-exist@corp.com"})
	assert.NotNil(t, err)
	t.Logf("%v", err)
}
