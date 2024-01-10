// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package badge

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	badgemodel "github.com/horizoncd/horizon/pkg/badge/models"
	clustermodel "github.com/horizoncd/horizon/pkg/cluster/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/server/global"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
)

var (
	ctx = context.Background()
	db  *gorm.DB
	mgr *managerparam.Manager
)

func createContext() {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&usermodels.User{},
		&clustermodel.Cluster{},
		&membermodels.Member{},
		&badgemodel.Badge{}); err != nil {
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
	ctrl := NewController(&param.Param{Manager: mgr})

	redirectLink := "https://https://horizoncd.github.io"
	_, err := ctrl.CreateBadge(ctx, common.ResourceCluster, 0, &Create{
		Name:         "horizon",
		SvgLink:      "https://horizoncd.io",
		RedirectLink: redirectLink,
	})

	c, err := mgr.ClusterMgr.Create(ctx,
		&clustermodel.Cluster{
			Model: global.Model{
				ID: 1,
			},
			Name: "test",
		}, nil, nil)

	assert.Nil(t, err)

	badge, err := ctrl.CreateBadge(ctx, common.ResourceCluster, c.ID, &Create{
		Name:         "horizon",
		SvgLink:      "https://github.com/horizoncd/horizon/svgs/horizon.svg",
		RedirectLink: redirectLink,
	})
	assert.Nil(t, err)

	badgeGot, err := ctrl.GetBadge(ctx, badge.ID)
	assert.Nil(t, err)
	assert.Equal(t, badge.ID, badgeGot.ID)
	assert.Equal(t, badge.Name, badgeGot.Name)

	updatedSvgLink := "https://github.com/horizoncd/horizon/svgs/horizon2.svg"
	badgeUpdated, err := ctrl.UpdateBadge(ctx, badge.ID, &Update{
		SvgLink: &updatedSvgLink,
	})

	assert.Equal(t, badgeUpdated.SvgLink, "https://github.com/horizoncd/horizon/svgs/horizon2.svg")
	assert.Equal(t, badgeUpdated.RedirectLink, badgeGot.RedirectLink)

	badgeGot, err = ctrl.GetBadgeByName(ctx, common.ResourceCluster, c.ID, badge.Name)
	assert.Nil(t, err)
	assert.Equal(t, badge.ID, badgeGot.ID)
	assert.Equal(t, badgeUpdated.SvgLink, "https://github.com/horizoncd/horizon/svgs/horizon2.svg")
	assert.Equal(t, badgeUpdated.RedirectLink, badgeGot.RedirectLink)

	_, err = ctrl.GetBadgeByName(ctx, common.ResourceCluster, c.ID, "not-exist")
	assert.NotNil(t, err)

	updatedRedirectLink := "https://horizoncd.github.io2"
	badge2, err := ctrl.CreateBadge(ctx, common.ResourceCluster, c.ID, &Create{
		Name:         "horizon2",
		SvgLink:      "https://github.com/horizoncd/horizon/svgs/horizon.svg2",
		RedirectLink: updatedRedirectLink,
	})
	assert.Nil(t, err)

	badges, err := ctrl.ListBadges(ctx, common.ResourceCluster, c.ID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(badges))
	for _, b := range badges {
		if b.ID == badge.ID {
			assert.Equal(t, b.Name, badgeUpdated.Name)
			assert.Equal(t, b.SvgLink, badgeUpdated.SvgLink)
			assert.Equal(t, b.RedirectLink, badgeUpdated.RedirectLink)
		} else {
			assert.Equal(t, b.Name, badge2.Name)
			assert.Equal(t, b.SvgLink, badge2.SvgLink)
			assert.Equal(t, b.RedirectLink, badge2.RedirectLink)
		}
	}

	err = ctrl.DeleteBadge(ctx, badge.ID)
	assert.Nil(t, err)

	badges, err = ctrl.ListBadges(ctx, common.ResourceCluster, c.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(badges))

	err = ctrl.DeleteBadgeByName(ctx, common.ResourceCluster, c.ID, badge2.Name)
	assert.Nil(t, err)

	badges, err = ctrl.ListBadges(ctx, common.ResourceCluster, c.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(badges))
}
