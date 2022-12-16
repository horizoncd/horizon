package service

import (
	"context"
	"os"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/user/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db      *gorm.DB
	ctx     context.Context
	manager *managerparam.Manager
)

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
	if err := db.AutoMigrate(&models.User{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	os.Exit(m.Run())
}

func Test_service_CheckUsersExists(t *testing.T) {
	userManger := manager.UserManager
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

	svc := NewService(manager)
	err = svc.CheckUsersExists(ctx, []string{"tony@corp.com"})
	assert.Nil(t, err)

	err = svc.CheckUsersExists(ctx, []string{"tony@corp.com", "mary@corp.com"})
	assert.Nil(t, err)

	err = svc.CheckUsersExists(ctx, []string{"tony@corp.com", "not-exist@corp.com"})
	assert.NotNil(t, err)
	t.Logf("%v", err)
}
