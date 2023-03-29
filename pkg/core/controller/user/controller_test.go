package user

import (
	"context"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/core/common"
	idpmodels "github.com/horizoncd/horizon/pkg/idp/models"
	"github.com/horizoncd/horizon/pkg/idp/utils"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/horizoncd/horizon/pkg/user/models"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	linkmodels "github.com/horizoncd/horizon/pkg/userlink/models"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

var (
	ctx    = context.Background()
	db     *gorm.DB
	mgr    *managerparam.Manager
	filter = "name"
)

func createContext() {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&usermodels.User{},
		&linkmodels.UserLink{}, &idpmodels.IdentityProvider{}); err != nil {
		panic(err)
	}
	mgr = managerparam.InitManager(db)
	ctx = context.Background()
	// nolint
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:  "Jerry",
		ID:    1,
		Admin: true,
	})
}

// nolint
func Test(t *testing.T) {
	createContext()

	userMgr := mgr.UserManager
	linkMgr := mgr.UserLinksManager
	ctrl := NewController(&param.Param{Manager: mgr})

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
	for _, user := range users {
		_, err := userMgr.Create(ctx, user)
		assert.Nil(t, err)
	}

	method := uint8(idpmodels.ClientSecretSentAsPost)
	err := db.Table("tb_identity_provider").Create(&idpmodels.IdentityProvider{
		Model:                   global.Model{ID: 1},
		DisplayName:             "netease",
		Name:                    "netease",
		TokenEndpointAuthMethod: (*idpmodels.TokenEndpointAuthMethod)(&method),
	}).Error
	assert.Nil(t, err)

	_, err = linkMgr.CreateLink(ctx, 1, 1, &utils.Claims{
		Sub:   "name1",
		Name:  "name1",
		Email: "name1@example.com",
	}, true)
	assert.Nil(t, err)

	_, err = linkMgr.CreateLink(ctx, 2, 1, &utils.Claims{
		Sub:   "name2",
		Name:  "name2",
		Email: "name2@example.com",
	}, false)
	assert.Nil(t, err)

	count, res, err := ctrl.List(ctx, &q.Query{Keywords: q.KeyWords{common.UserQueryName: filter}, PageSize: 1})
	assert.Nil(t, err)
	assert.Equal(t, int64(2), count)
	assert.Equal(t, 1, len(res))

	links, err := ctrl.ListUserLinks(ctx, 1)
	assert.Nil(t, err)
	assert.NotNil(t, links)
	assert.Equal(t, 1, len(links))

	err = ctrl.DeleteLinksByID(ctx, 1)
	assert.Nil(t, err)

	err = ctrl.DeleteLinksByID(ctx, 2)
	assert.NotNil(t, err)

	user, err := ctrl.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, uint(1), user.ID)
	assert.Equal(t, "name1", user.Name)
	assert.Equal(t, true, user.IsAdmin)
	assert.Equal(t, false, user.IsBanned)

	resT, resF := true, false
	_, err = ctrl.UpdateByID(ctx, 1, &UpdateUserRequest{
		IsAdmin:  &resF,
		IsBanned: &resT,
	})
	assert.Nil(t, err)

	user, err = ctrl.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, uint(1), user.ID)
	assert.Equal(t, "name1", user.Name)
	assert.Equal(t, false, user.IsAdmin)
	assert.Equal(t, true, user.IsBanned)

	user, err = ctrl.UpdateByID(ctx, 2, &UpdateUserRequest{
		IsAdmin:  &resF,
		IsBanned: &resT,
	})
	assert.Nil(t, err)
	assert.Equal(t, uint(2), user.ID)
	assert.Equal(t, "name2", user.Name)
	assert.Equal(t, false, user.IsAdmin)
	assert.Equal(t, true, user.IsBanned)

	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Admin: false,
	})

	_, err = ctrl.UpdateByID(ctx, 2, &UpdateUserRequest{
		IsAdmin:  &resF,
		IsBanned: &resT,
	})
	assert.NotNil(t, err)
}
