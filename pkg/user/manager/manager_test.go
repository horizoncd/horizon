package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"g.hz.netease.com/horizon/core/common"
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
		name  = "Tony"
		email = "tony@163.com"
	)
	u, err := mgr.Create(ctx, &models.User{
		Name:  name,
		Email: email,
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(string(b))

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
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = mgr.Create(ctx, &models.User{
			Name:     fmt.Sprintf("%s%d", name2, i),
			Email:    fmt.Sprintf("%s%d@163.com", name2, i),
			FullName: fmt.Sprintf("%s%d", strings.ToUpper(name2), i),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	count, users, err := mgr.List(ctx, &q.Query{
		Keywords:   q.KeyWords{common.UserQueryName: name1},
		PageNumber: 1,
		PageSize:   5,
	})
	assert.Nil(t, err)
	assert.Equal(t, 5, len(users))
	assert.Equal(t, int64(10), count)
	for _, u := range users {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}

	count, users, err = mgr.List(ctx, &q.Query{
		Keywords:   q.KeyWords{common.UserQueryName: name1},
		PageNumber: 2,
		PageSize:   5,
	})
	assert.Nil(t, err)
	assert.Equal(t, 5, len(users))
	assert.Equal(t, int64(10), count)
	for _, u := range users {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}

	count, users, err = mgr.List(ctx, &q.Query{
		Keywords:   q.KeyWords{common.UserQueryName: name2},
		PageNumber: 0,
		PageSize:   3,
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(users))
	assert.Equal(t, int64(10), count)
	for _, u := range users {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}

	count, users, err = mgr.List(ctx, &q.Query{
		Keywords:   q.KeyWords{common.UserQueryName: "5"},
		PageNumber: 0,
		PageSize:   3,
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(users))
	assert.Equal(t, int64(2), count)
	for _, u := range users {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}

	count, users, err = mgr.List(ctx, &q.Query{
		Keywords: q.KeyWords{common.UserQueryName: "e"},
	})
	assert.Nil(t, err)
	assert.Equal(t, 20, len(users))
	assert.Equal(t, int64(20), count)
	for _, u := range users {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}
}

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&models.User{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	os.Exit(m.Run())
}
