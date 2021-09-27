package user

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"g.hz.netease.com/horizon/core/middleware/user"
	usermock "g.hz.netease.com/horizon/mock/pkg/user"
	"g.hz.netease.com/horizon/pkg/user/models"

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

	users := []models.User{
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

	_, err = c.GetName(ctx)
	assert.NotNil(t, err)

	ctx = context.WithValue(ctx, user.Key(), &models.User{
		Model: gorm.Model{
			ID: 1,
		},
		Name: "tony",
	})

	name, err := c.GetName(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "tony", name)

	id, err := c.GetID(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, id)
}
