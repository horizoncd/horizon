package service

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	"github.com/stretchr/testify/assert"
)

var (
	// use tmp sqlite
	db, _   = orm.NewSqliteDB("")
	ctx     = context.TODO()
	manager = managerparam.InitManager(db)
)

// nolint
func init() {
	// create table
	err := db.AutoMigrate(&models.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}

func TestServiceGetChildByID(t *testing.T) {
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
		want    *Child
		wantErr bool
	}{
		{
			name: "GetChildByID",
			args: args{
				id: group.ID,
			},
			want: &Child{
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
			s := service{
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
			}
		})
	}
}

func TestConvertApplicationToChild(t *testing.T) {
	app := &appmodels.Application{
		Name: "test",
	}
	f := &Full{
		FullName: "full",
	}
	c := ConvertApplicationToChild(app, f)
	assert.Equal(t, "test", c.Name)
}

func TestConvertGroupOrApplicationToChild(t *testing.T) {
	app := &models.GroupOrApplication{
		Name: "test",
	}
	f := &Full{
		FullName: "full",
	}
	c := ConvertGroupOrApplicationToChild(app, f)
	assert.Equal(t, "test", c.Name)
}

func TestConvertClusterToChild(t *testing.T) {
	cluster := &clustermodels.Cluster{
		Name: "test",
	}
	f := &Full{
		FullName: "full",
	}
	c := ConvertClusterToChild(cluster, f)
	assert.Equal(t, "test", c.Name)
}

func TestConvertGroupToChild(t *testing.T) {
	gp := &models.Group{
		Name: "test",
	}
	f := &Full{
		FullName: "full",
	}
	c := ConvertGroupToChild(gp, f)
	assert.Equal(t, "test", c.Name)
}
