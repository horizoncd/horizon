package member

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/member/models"
	"github.com/golang/mock/gomock"
	"gorm.io/gorm"
)
var (
	db  *gorm.DB
	ctx context.Context
	memberService Service
)


func TestList(t *testing.T) {
	// mock the groupManager
	ctrl := gomock.NewController(t)
	defer  ctrl.Finish()
	NewMock

	t.Fatal("123")
}

func TestMain(m *testing.M) {

	db, _ := orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Member{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)



	os.Exit(m.Run())
}