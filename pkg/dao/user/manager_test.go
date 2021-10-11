package user

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	orm2 "g.hz.netease.com/horizon/pkg/lib/orm"
	"g.hz.netease.com/horizon/pkg/lib/q"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
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
	u, err := Mgr.Create(ctx, &User{
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

func TestSearchUser(t *testing.T) {
	var (
		name1 = "jessy"
		name2 = "mike"

		err error
	)
	for i := 0; i < 10; i++ {
		_, err = Mgr.Create(ctx, &User{
			Name:     fmt.Sprintf("%s%d", name1, i),
			Email:    fmt.Sprintf("%s%d@163.com", name1, i),
			FullName: fmt.Sprintf("%s%d", strings.ToUpper(name1), i),
			OIDCId:   strconv.Itoa(i),
			OIDCType: "netease",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = Mgr.Create(ctx, &User{
			Name:     fmt.Sprintf("%s%d", name2, i),
			Email:    fmt.Sprintf("%s%d@163.com", name2, i),
			FullName: fmt.Sprintf("%s%d", strings.ToUpper(name2), i),
			OIDCId:   strconv.Itoa(i),
			OIDCType: "netease",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	count, users, err := Mgr.SearchUser(ctx, name1, &q.Query{
		PageNumber: 1,
		PageSize:   5,
	})
	assert.Nil(t, err)
	assert.Equal(t, 5, len(users))
	assert.Equal(t, 10, count)
	for _, u := range users {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}

	count, users, err = Mgr.SearchUser(ctx, name1, &q.Query{
		PageNumber: 2,
		PageSize:   5,
	})
	assert.Nil(t, err)
	assert.Equal(t, 5, len(users))
	assert.Equal(t, 10, count)
	for _, u := range users {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}

	count, users, err = Mgr.SearchUser(ctx, name2, &q.Query{
		PageNumber: 0,
		PageSize:   3,
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(users))
	assert.Equal(t, 10, count)
	for _, u := range users {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}

	count, users, err = Mgr.SearchUser(ctx, "5", &q.Query{
		PageNumber: 0,
		PageSize:   3,
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(users))
	assert.Equal(t, 2, count)
	for _, u := range users {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}

	count, users, err = Mgr.SearchUser(ctx, "e", nil)
	assert.Nil(t, err)
	assert.Equal(t, 20, len(users))
	assert.Equal(t, 20, count)
	for _, u := range users {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}
}

func TestMain(m *testing.M) {
	db, _ = orm2.NewSqliteDB("")
	if err := db.AutoMigrate(&User{}); err != nil {
		panic(err)
	}
	ctx = orm2.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
