package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/user/models"

	"github.com/stretchr/testify/assert"
)

var (
	db, _ = orm.NewSqliteDB("")
	ctx   context.Context
	mgr   = New(db)
)

func Test(t *testing.T) {
	var (
		name     = "Tony"
		email    = "tony@163.com"
		oidcID   = "H12323"
		oidcType = "ne"
	)
	u, err := mgr.Create(ctx, &models.User{
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

	u2, err := mgr.GetByOIDCMeta(ctx, oidcType, email)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, name, u2.Name)
	assert.Equal(t, email, u2.Email)

	// user not exits
	u3, err := mgr.GetByOIDCMeta(ctx, "not-exist", "not-exist")
	assert.Nil(t, err)
	assert.Nil(t, u3)

	u4, err := mgr.GetUserByEmail(ctx, email)
	assert.Nil(t, err)
	assert.NotNil(t, u4)
	assert.Equal(t, u4.Name, name)

	users, err := mgr.GetUserByIDs(ctx, []uint{u4.ID})
	assert.Nil(t, err)
	assert.NotNil(t, users)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, u4.Name, users[0].Name)

	userMap, err := mgr.GetUserMapByIDs(ctx, []uint{u4.ID})
	assert.Nil(t, err)
	assert.NotNil(t, userMap)
	assert.Equal(t, 1, len(userMap))
	assert.Equal(t, u4.Name, userMap[u4.ID].Name)
}

func TestSearchUser(t *testing.T) {
	var (
		name1 = "jessy"
		name2 = "mike"

		err error
	)
	for i := 0; i < 10; i++ {
		_, err = mgr.Create(ctx, &models.User{
			Name:     fmt.Sprintf("%s%d", name1, i),
			Email:    fmt.Sprintf("%s%d@163.com", name1, i),
			FullName: fmt.Sprintf("%s%d", strings.ToUpper(name1), i),
			OIDCId:   strconv.Itoa(i),
			OIDCType: "netease",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = mgr.Create(ctx, &models.User{
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

	count, users, err := mgr.SearchUser(ctx, name1, &q.Query{
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

	count, users, err = mgr.SearchUser(ctx, name1, &q.Query{
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

	count, users, err = mgr.SearchUser(ctx, name2, &q.Query{
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

	count, users, err = mgr.SearchUser(ctx, "5", &q.Query{
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

	count, users, err = mgr.SearchUser(ctx, "e", nil)
	assert.Nil(t, err)
	assert.Equal(t, 20, len(users))
	assert.Equal(t, 20, count)
	for _, u := range users {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}
}

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&models.User{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
