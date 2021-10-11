package user

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	usermock "g.hz.netease.com/horizon/mock/pkg/dao/user"
	"g.hz.netease.com/horizon/pkg/dao/user"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	ctx    = context.Background()
	filter = "name"
)

// nolint
func Test(t *testing.T) {
	mockCtl := gomock.NewController(t)
	userMgr := usermock.NewMockManager(mockCtl)

	users := []user.User{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Name:     "name1",
			FullName: "Name1",
			Email:    "name1@example.com",
		}, {
			Model: gorm.Model{
				ID: 2,
			},
			Name:     "name2",
			FullName: "Name2",
			Email:    "name2@example.com",
		},
	}

	userMgr.EXPECT().SearchUser(ctx, filter, nil).Return(
		10, users, nil)
	userMgr.EXPECT().SearchUser(ctx, filter+"1", nil).Return(
		0, nil, errors.New("err"))

	c := &controller{
		userMgr: userMgr,
	}
	count, res, err := c.SearchUser(ctx, filter, nil)
	assert.Nil(t, err)
	assert.Equal(t, 10, count)
	assert.Equal(t, 2, len(res))
	for _, u := range res {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}

	_, _, err = c.SearchUser(ctx, filter+"1", nil)
	assert.NotNil(t, err)
}

func TestT(t *testing.T) {
	readFile := func(b *[]byte, err *error) {
		bytes, e := []byte("asdf"), errors.New("asdf")
		*b = bytes
		*err = e
	}
	b := []byte("123")
	e := errors.New("123")
	t.Logf("b: %v", string(b))
	t.Logf("e: %v", e)

	readFile(&b, &e)
	t.Logf("b: %v", string(b))
	t.Logf("e: %v", e)

	unmarshal := func(b []byte, m *map[string]interface{}, err *error) {
		if e := json.Unmarshal(b, m); e != nil {
			*err = e
		}
	}
	bytes := []byte(`{"1": "2"}`)
	var m map[string]interface{}
	var err1 error
	t.Logf("m: %v", m)
	t.Logf("err1: %v", err1)
	unmarshal(bytes, &m, &err1)
	t.Logf("m: %v", m)
	t.Logf("err1: %v", err1)
}
