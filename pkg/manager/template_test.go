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

package manager_test

import (
	"context"
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	applicationmodel "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func createTemplateCtx() (context.Context, *managerparam.Manager, *gorm.DB) {
	var (
		db     *gorm.DB
		ctx    context.Context
		params *managerparam.Manager
	)

	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&applicationmodel.Template{}, &applicationmodel.TemplateRelease{},
		&applicationmodel.Member{},
		&applicationmodel.Application{}, &applicationmodel.Cluster{}); err != nil {
		panic(err)
	}
	ctx = context.Background()
	// nolint
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		ID:   1,
		Name: "Jerry",
	})

	params = managerparam.InitManager(db)
	return ctx, params, db
}

func TestTemplate(t *testing.T) {
	ctx, params, _ := createTemplateCtx()
	template1 := &applicationmodel.Template{
		Name:        "javaapp1",
		Description: "java app for test",
		GroupID:     1,
	}
	template1InDB, err := params.TemplateMgr.Create(ctx, template1)
	assert.Nil(t, err)

	assert.Equal(t, template1.Name, template1InDB.Name)
	assert.Equal(t, template1.Description, template1InDB.Description)
	assert.Equal(t, 1, int(template1.ID))

	template2 := &applicationmodel.Template{
		Name:        "javaapp2",
		Description: "java app for test 2",
		GroupID:     2,
	}

	_, err = params.TemplateMgr.Create(ctx, template2)
	assert.Nil(t, err)

	template2InDB, err := params.TemplateMgr.GetByID(ctx, 2)
	assert.Nil(t, err)
	assert.Equal(t, template2.Name, template2InDB.Name)
	assert.Equal(t, template2.Description, template2InDB.Description)
	assert.Equal(t, template2.WithoutCI, false)

	template2InDB, err = params.TemplateMgr.GetByName(ctx, template2.Name)
	assert.Nil(t, err)
	assert.Equal(t, template2.Description, template2InDB.Description)

	templates, err := params.TemplateMgr.ListTemplate(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(templates))
	assert.Equal(t, template1.Name, templates[0].Name)
	assert.Equal(t, template1.Description, templates[0].Description)
	assert.Equal(t, 1, int(templates[0].ID))
	assert.Equal(t, template2.Name, templates[1].Name)
	assert.Equal(t, template2.Description, templates[1].Description)
	assert.Equal(t, 2, int(templates[1].ID))

	templates, err = params.TemplateMgr.ListByGroupID(ctx, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, template2.Name, templates[0].Name)
	assert.Equal(t, template2.Description, templates[0].Description)
	assert.Equal(t, 2, int(templates[0].ID))

	template2.Description = "changed description"
	err = params.TemplateMgr.UpdateByID(ctx, 2, template2)
	assert.Nil(t, err)

	template2InDB, err = params.TemplateMgr.GetByID(ctx, 2)
	assert.Nil(t, err)
	assert.Equal(t, template2.Name, template2InDB.Name)
	assert.Equal(t, template2.Description, template2InDB.Description)

	templates, err = params.TemplateMgr.ListByGroupIDs(ctx, []uint{1, 2})
	assert.Nil(t, err)
	assert.NotNil(t, templates)
	assert.Equal(t, 2, len(templates))

	templates, err = params.TemplateMgr.ListByGroupIDs(ctx, []uint{1})
	assert.Nil(t, err)
	assert.NotNil(t, templates)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, template1.Name, templates[0].Name)

	templates, err = params.TemplateMgr.ListByIDs(ctx, []uint{1, 2})
	assert.Nil(t, err)
	assert.NotNil(t, templates)
	assert.Equal(t, 2, len(templates))

	templates, err = params.TemplateMgr.ListByIDs(ctx, []uint{1})
	assert.Nil(t, err)
	assert.NotNil(t, templates)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, template1.Name, templates[0].Name)

	err = params.TemplateMgr.DeleteByID(ctx, 2)
	assert.Nil(t, err)

	template2InDB, err = params.TemplateMgr.GetByID(ctx, 2)
	assert.NotNil(t, err)
	assert.Nil(t, template2InDB)

	app := &applicationmodel.Application{
		Model:    global.Model{ID: 1},
		Template: template1.Name,
		Name:     "test",
	}
	_, err = params.ApplicationManager.Create(ctx, app, map[string]string{})
	assert.Nil(t, err)

	cluster := &applicationmodel.Cluster{
		Model:         global.Model{ID: 1},
		ApplicationID: 1,
		Name:          "testgroup",
		Template:      template1.Name,
	}
	_, err = params.ClusterMgr.Create(ctx, cluster, nil, nil)
	assert.Nil(t, err)

	apps, _, err := params.TemplateMgr.GetRefOfApplication(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, app.Name, apps[0].Name)

	clusters, _, err := params.TemplateMgr.GetRefOfCluster(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, cluster.Name, clusters[0].Name)
}
