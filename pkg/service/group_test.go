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

package service

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func createGroupCtx() (context.Context, *managerparam.Manager, *gorm.DB) {
	var (
		// use tmp sqlite
		db, _   = orm.NewSqliteDB("")
		ctx     = context.TODO()
		manager = managerparam.InitManager(db)
	)

	err := db.AutoMigrate(&models.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}

	return ctx, manager, db
}

func TestServiceGetChildByID(t *testing.T) {
	ctx, manager, db := createGroupCtx()
	group := &models.Group{
		Name:         "a",
		Path:         "1",
		TraversalIDs: "1",
	}
	db.Save(group)

	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Child
		wantErr bool
	}{
		{
			name: "GetChildByID",
			args: args{
				id: group.ID,
			},
			want: &models.Child{
				ID:           group.ID,
				Name:         "a",
				Path:         "1",
				TraversalIDs: "1",
				FullPath:     "/1",
				FullName:     "a",
				Type:         ChildTypeGroup,
				UpdatedAt:    group.UpdatedAt,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := groupService{
				groupManager: manager.GroupManager,
			}
			got, err := s.GetChildByID(ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChildByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got.UpdatedAt = tt.want.UpdatedAt
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChildByID() got = %v, want %v", got, tt.want)
				return
			}

			children, err := s.GetChildrenByIDs(ctx, []uint{tt.args.id})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChildrenByIDs() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if len(children) != 1 {
				t.Errorf("GetChildrenByIDs() returns %v child(ren), but only 1 want", len(children))
			}
			got = children[tt.args.id]
			got.UpdatedAt = tt.want.UpdatedAt
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChildByIDs() got = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func TestConvertApplicationToChild(t *testing.T) {
	app := &models.Application{
		Name: "test",
	}
	f := &models.Full{
		FullName: "full",
	}
	c := ConvertApplicationToChild(app, f)
	assert.Equal(t, "test", c.Name)
}

func TestConvertGroupOrApplicationToChild(t *testing.T) {
	app := &models.GroupOrApplication{
		Name: "test",
	}
	f := &models.Full{
		FullName: "full",
	}
	c := ConvertGroupOrApplicationToChild(app, f)
	assert.Equal(t, "test", c.Name)
}

func TestConvertClusterToChild(t *testing.T) {
	cluster := &models.Cluster{
		Name: "test",
	}
	f := &models.Full{
		FullName: "full",
	}
	c := ConvertClusterToChild(cluster, f)
	assert.Equal(t, "test", c.Name)
}

func TestConvertGroupToChild(t *testing.T) {
	gp := &models.Group{
		Name: "test",
	}
	f := &models.Full{
		FullName: "full",
	}
	c := ConvertGroupToChild(gp, f)
	assert.Equal(t, "test", c.Name)
}
