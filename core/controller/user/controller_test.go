package user

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	usermock "g.hz.netease.com/horizon/mock/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/server/global"
	"g.hz.netease.com/horizon/pkg/user/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	ctx    = context.Background()
	filter = "name"
)

// nolint
func Test(t *testing.T) {
	mockCtl := gomock.NewController(t)
	userMgr := usermock.NewMockManager(mockCtl)

	users := []*models.User{
		{
			Model: global.Model{
				ID: 1,
			},
			Name:     "name1",
			FullName: "Name1",
			Email:    "name1@example.com",
			Admin:    true,
		}, {
			Model: global.Model{
				ID: 2,
			},
			Name:     "name2",
			FullName: "Name2",
			Email:    "name2@example.com",
			Admin:    false,
		},
	}

	userMgr.EXPECT().List(ctx, &q.Query{Keywords: q.KeyWords{common.UserQueryName: filter}}).Return(
		int64(10), users, nil)
	userMgr.EXPECT().List(ctx, &q.Query{Keywords: q.KeyWords{common.UserQueryName: filter + "1"}}).Return(
		int64(0), nil, errors.New("err"))

	c := &controller{
		userMgr: userMgr,
	}
	count, res, err := c.List(ctx, &q.Query{Keywords: q.KeyWords{common.UserQueryName: filter}})
	assert.Nil(t, err)
	assert.Equal(t, int64(10), count)
	assert.Equal(t, 2, len(res))
	for _, u := range res {
		b, _ := json.Marshal(u)
		t.Logf("%v", string(b))
	}

	_, _, err = c.List(ctx, &q.Query{Keywords: q.KeyWords{common.UserQueryName: filter + "1"}})
	assert.NotNil(t, err)

	// test GetUserByEmail
	userMgr.EXPECT().GetUserByEmail(ctx, "name1@example.com").Return(users[0], nil)
	userMgr.EXPECT().GetUserByEmail(ctx, "name2@example.com").Return(
		nil, herrors.NewErrGetFailed(herrors.UserInDB, ""))
	user, err := c.GetUserByEmail(ctx, "name1@example.com")
	assert.Nil(t, err)
	assert.Equal(t, uint(1), user.ID)
	assert.Equal(t, "name1", user.Name)
	_, err = c.GetUserByEmail(ctx, "name2@example.com")
	assert.NotNil(t, err)
}
