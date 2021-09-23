package user

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/user/models"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func Test(t *testing.T) {
	var (
		name     = "Tony"
		email    = "tony@163.com"
		oidcID   = "H12323"
		oidcType = "ne"
	)
	u, err := Mgr.Create(ctx, &models.User{
		Name:     name,
		Email:    email,
		OIDCId:   oidcID,
		OIDCType: oidcType,
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(string(b))

	u2, err := Mgr.GetByOIDCMeta(ctx, oidcID, oidcType)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, name, u2.Name)
	assert.Equal(t, email, u2.Email)

	// user not exits
	u3, err := Mgr.GetByOIDCMeta(ctx, "not-exist", "not-exist")
	assert.Nil(t, err)
	assert.Nil(t, u3)
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.User{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
