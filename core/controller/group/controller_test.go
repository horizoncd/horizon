package group

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/group/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	// use tmp sqlite
	db, _ = orm.NewSqliteDB("")
	ctx   = orm.NewContext(context.TODO(), db)
)

func init() {
	// create table
	err := db.AutoMigrate(&models.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}

func TestControllerCreateGroup(t *testing.T) {
	type args struct {
		ctx      context.Context
		newGroup *NewGroup
	}
	tests := []struct {
		name    string
		args    args
		want    uint
		wantErr bool
	}{
		{
			name: "createRootGroup",
			args: args{
				ctx: ctx,
				newGroup: &NewGroup{
					Name:            "1",
					Path:            "a",
					VisibilityLevel: "private",
				},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "nameConflict",
			args: args{
				ctx: ctx,
				newGroup: &NewGroup{
					Name:            "1",
					Path:            "aa",
					VisibilityLevel: "private",
				},
			},
			wantErr: true,
		},
		{
			name: "pathConflict",
			args: args{
				ctx: ctx,
				newGroup: &NewGroup{
					Name:            "11",
					Path:            "a",
					VisibilityLevel: "private",
				},
			},
			wantErr: true,
		},
		{
			name: "createSubGroup",
			args: args{
				ctx: ctx,
				newGroup: &NewGroup{
					Name:            "1",
					Path:            "a",
					VisibilityLevel: "private",
					ParentID:        1,
				},
			},
			want:    2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Ctl.CreateGroup(tt.args.ctx, tt.args.newGroup)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateGroup() got = %v, want %v", got, tt.want)
			}
		})
	}

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
}

func TestControllerDelete(t *testing.T) {
	newRootGroup := &NewGroup{
		Name:            "1",
		Path:            "a",
		VisibilityLevel: "private",
	}

	id, err := Ctl.CreateGroup(ctx, newRootGroup)
	assert.Nil(t, err)
	assert.Greater(t, id, uint(0))

	type args struct {
		ctx context.Context
		id  uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "deleteNotExist",
			args: args{
				ctx: ctx,
				id:  0,
			},
			wantErr: true,
		},
		{
			name: "deleteExist",
			args: args{
				ctx: ctx,
				id:  id,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Ctl.Delete(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
}

func TestControllerGetByID(t *testing.T) {
	newRootGroup := &NewGroup{
		Name:            "1",
		Path:            "a",
		VisibilityLevel: "private",
	}

	id, err := Ctl.CreateGroup(ctx, newRootGroup)
	assert.Nil(t, err)

	child, err := Ctl.GetByID(ctx, id)
	assert.Nil(t, err)

	type args struct {
		ctx context.Context
		id  uint
	}
	tests := []struct {
		name    string
		args    args
		want    *Child
		wantErr bool
	}{
		{
			name: "getExist",
			args: args{
				ctx: ctx,
				id:  id,
			},
			want: &Child{
				ID:              id,
				Name:            "1",
				Path:            "a",
				VisibilityLevel: "private",
				ParentID:        0,
				TraversalIDs:    strconv.Itoa(int(id)),
				FullPath:        "/a",
				FullName:        "1",
				Type:            ChildType,
				UpdatedAt:       child.UpdatedAt,
			},
		},
		{
			name: "getNotExist",
			args: args{
				ctx: ctx,
				id:  0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Ctl.GetByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByID() got = %v, want %v", got, tt.want)
			}
		})
	}

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
}

func TestControllerGetByPath(t *testing.T) {
	newRootGroup := &NewGroup{
		Name:            "1",
		Path:            "a",
		VisibilityLevel: "private",
	}

	id, err := Ctl.CreateGroup(ctx, newRootGroup)
	assert.Nil(t, err)
	child, err := Ctl.GetByID(ctx, id)
	assert.Nil(t, err)

	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Child
		wantErr bool
	}{
		{
			name: "getExist",
			args: args{
				ctx:  ctx,
				path: "/a",
			},
			want: &Child{
				ID:              id,
				Name:            "1",
				Path:            "a",
				VisibilityLevel: "private",
				ParentID:        0,
				TraversalIDs:    strconv.Itoa(int(id)),
				FullPath:        "/a",
				FullName:        "1",
				Type:            ChildType,
				UpdatedAt:       child.UpdatedAt,
			},
		},
		{
			name: "getNotExist",
			args: args{
				ctx:  ctx,
				path: "b",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Ctl.GetByFullPath(tt.args.ctx, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByFullPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByFullPath() got = %v, want %v", got, tt.want)
			}
		})
	}

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
}

func TestControllerGetChildren(t *testing.T) {
	// todo including application
}

func TestControllerGetSubGroups(t *testing.T) {
	newRootGroup := &NewGroup{
		Name:            "1",
		Path:            "a",
		VisibilityLevel: "private",
	}
	id, _ := Ctl.CreateGroup(ctx, newRootGroup)
	group1, _ := Ctl.GetByID(ctx, id)

	newGroup := &NewGroup{
		Name:            "2",
		Path:            "b",
		VisibilityLevel: "private",
		ParentID:        id,
	}
	id2, _ := Ctl.CreateGroup(ctx, newGroup)
	group2, _ := Ctl.GetByID(ctx, id2)

	type args struct {
		ctx        context.Context
		id         uint
		pageNumber int
		pageSize   int
	}
	tests := []struct {
		name    string
		args    args
		want    []*Child
		want1   int64
		wantErr bool
	}{
		{
			name: "getRootSubGroups",
			args: args{
				ctx: ctx,
			},
			want: []*Child{
				{
					ID:              id,
					Name:            "1",
					Path:            "a",
					VisibilityLevel: "private",
					ParentID:        0,
					TraversalIDs:    strconv.Itoa(int(id)),
					FullPath:        "/a",
					FullName:        "1",
					Type:            ChildType,
					ChildrenCount:   1,
					UpdatedAt:       group1.UpdatedAt,
				},
			},
			want1: 1,
		},
		{
			name: "getSubGroups",
			args: args{
				ctx: ctx,
				id:  id,
			},
			want: []*Child{
				{
					ID:              id2,
					Name:            "2",
					Path:            "b",
					VisibilityLevel: "private",
					ParentID:        id,
					TraversalIDs:    strconv.Itoa(int(id)) + "," + strconv.Itoa(int(id2)),
					FullPath:        "/a/b",
					FullName:        "1 / 2",
					Type:            ChildType,
					ChildrenCount:   0,
					UpdatedAt:       group2.UpdatedAt,
				},
			},
			want1: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Ctl.GetSubGroups(tt.args.ctx, tt.args.id, tt.args.pageNumber, tt.args.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSubGroups() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetSubGroups() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
}

func TestControllerSearchChildren(t *testing.T) {
	// todo including application
}

func TestControllerSearchGroups(t *testing.T) {
	newRootGroup := &NewGroup{
		Name:            "1",
		Path:            "a",
		VisibilityLevel: "private",
	}
	id, _ := Ctl.CreateGroup(ctx, newRootGroup)
	group1, _ := Ctl.GetByID(ctx, id)

	newGroup := &NewGroup{
		Name:            "2",
		Path:            "b",
		VisibilityLevel: "private",
		ParentID:        id,
	}
	id2, _ := Ctl.CreateGroup(ctx, newGroup)
	group2, _ := Ctl.GetByID(ctx, id2)

	type args struct {
		ctx    context.Context
		id     uint
		filter string
	}
	tests := []struct {
		name    string
		args    args
		want    []*Child
		want1   int64
		wantErr bool
	}{
		{
			name: "blankFilter",
			args: args{
				ctx:    ctx,
				filter: "",
			},
			want: []*Child{
				{
					ID:              id,
					Name:            "1",
					Path:            "a",
					VisibilityLevel: "private",
					ParentID:        0,
					TraversalIDs:    strconv.Itoa(int(id)),
					FullPath:        "/a",
					FullName:        "1",
					Type:            ChildType,
					ChildrenCount:   1,
					UpdatedAt:       group1.UpdatedAt,
				},
			},
			want1: 1,
		},
		{
			name: "noMatch",
			args: args{
				ctx:    ctx,
				filter: "3",
			},
			want:  []*Child{},
			want1: 0,
		},
		{
			name: "matchFirstLevel",
			args: args{
				ctx:    ctx,
				filter: "1",
			},
			want: []*Child{
				{
					ID:              id,
					Name:            "1",
					Path:            "a",
					VisibilityLevel: "private",
					ParentID:        0,
					TraversalIDs:    strconv.Itoa(int(id)),
					FullPath:        "/a",
					FullName:        "1",
					Type:            ChildType,
					UpdatedAt:       group1.UpdatedAt,
				},
			},
			want1: 1,
		},
		{
			name: "matchSecondLevel",
			args: args{
				ctx:    ctx,
				filter: "2",
			},
			want: []*Child{
				{
					ID:              id,
					Name:            "1",
					Path:            "a",
					VisibilityLevel: "private",
					ParentID:        0,
					TraversalIDs:    strconv.Itoa(int(id)),
					FullPath:        "/a",
					FullName:        "1",
					Type:            ChildType,
					ChildrenCount:   1,
					UpdatedAt:       group1.UpdatedAt,
					Children: []*Child{
						{
							ID:              id2,
							Name:            "2",
							Path:            "b",
							VisibilityLevel: "private",
							ParentID:        id,
							TraversalIDs:    strconv.Itoa(int(id)) + "," + strconv.Itoa(int(id2)),
							FullPath:        "/a/b",
							FullName:        "1 / 2",
							Type:            ChildType,
							ChildrenCount:   0,
							UpdatedAt:       group2.UpdatedAt,
						},
					},
				},
			},
			want1: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Ctl.SearchGroups(tt.args.ctx, &SearchParams{
				GroupID:    tt.args.id,
				Filter:     tt.args.filter,
				PageNumber: common.DefaultPageNumber,
				PageSize:   common.MaxPageSize,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SearchGroups() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("SearchGroups() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
}

func TestControllerTransfer(t *testing.T) {
	newRootGroup := &NewGroup{
		Name:            "1",
		Path:            "a",
		VisibilityLevel: "private",
	}
	id, _ := Ctl.CreateGroup(ctx, newRootGroup)
	newGroup := &NewGroup{
		Name:            "2",
		Path:            "b",
		VisibilityLevel: "private",
		ParentID:        id,
	}
	id2, _ := Ctl.CreateGroup(ctx, newGroup)

	newRootGroup2 := &NewGroup{
		Name:            "3",
		Path:            "c",
		VisibilityLevel: "private",
	}
	id3, _ := Ctl.CreateGroup(ctx, newRootGroup2)

	type args struct {
		ctx         context.Context
		id          uint
		newParentID uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "twoRecordsTransfer",
			args: args{
				ctx:         ctx,
				id:          id,
				newParentID: id3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Ctl.Transfer(tt.args.ctx, tt.args.id, tt.args.newParentID); (err != nil) != tt.wantErr {
				t.Errorf("Transfer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// check transfer success
	g, err := Ctl.GetByID(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, id3, g.ParentID)
	assert.Equal(t, strconv.Itoa(int(id3))+","+strconv.Itoa(int(id)), g.TraversalIDs)

	g, err = Ctl.GetByID(ctx, id2)
	assert.Nil(t, err)
	assert.Equal(t, strconv.Itoa(int(id3))+","+strconv.Itoa(int(id))+","+strconv.Itoa(int(id2)), g.TraversalIDs)

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
}

func TestControllerUpdateBasic(t *testing.T) {
	newRootGroup := &NewGroup{
		Name:            "1",
		Path:            "a",
		VisibilityLevel: "private",
	}
	id, _ := Ctl.CreateGroup(ctx, newRootGroup)

	type args struct {
		ctx         context.Context
		id          uint
		updateGroup *UpdateGroup
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "updateNotExist",
			args: args{
				ctx: ctx,
				id:  0,
				updateGroup: &UpdateGroup{
					Name:            "2",
					Path:            "b",
					VisibilityLevel: "public",
				},
			},
			wantErr: true,
		},
		{
			name: "updateExist",
			args: args{
				ctx: ctx,
				id:  id,
				updateGroup: &UpdateGroup{
					Name:            "2",
					Path:            "b",
					VisibilityLevel: "public",
					Description:     "111",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Ctl.UpdateBasic(tt.args.ctx, tt.args.id, tt.args.updateGroup); (err != nil) != tt.wantErr {
				t.Errorf("UpdateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateChildrenWithLevelStruct(t *testing.T) {
	type args struct {
		groupID uint
		groups  []*models.Group
	}
	tests := []struct {
		name string
		args args
		want []*Child
	}{
		{
			name: "noMatch",
			args: args{
				groupID: 10,
				groups: []*models.Group{
					{
						Model: gorm.Model{
							ID: 1,
						},
						Name:         "1",
						Path:         "a",
						TraversalIDs: "1",
						ParentID:     0,
					},
				},
			},
			want: []*Child{},
		},
		{
			name: "match",
			args: args{
				groupID: 1,
				groups: []*models.Group{
					{
						Model: gorm.Model{
							ID: 1,
						},
						Name:         "1",
						Path:         "a",
						TraversalIDs: "1",
						ParentID:     0,
					},
					{
						Model: gorm.Model{
							ID: 2,
						},
						Name:         "2",
						Path:         "b",
						TraversalIDs: "1,2",
						ParentID:     1,
					},
					{
						Model: gorm.Model{
							ID: 3,
						},
						Name:         "3",
						Path:         "c",
						TraversalIDs: "1,2,3",
						ParentID:     2,
					},
					{
						Model: gorm.Model{
							ID: 4,
						},
						Name:         "4",
						Path:         "d",
						TraversalIDs: "1,4",
						ParentID:     1,
					},
					{
						Model: gorm.Model{
							ID: 5,
						},
						Name:         "5",
						Path:         "e",
						TraversalIDs: "1,4,5",
						ParentID:     4,
					},
				},
			},
			want: []*Child{
				{
					ID:            2,
					Name:          "2",
					Path:          "b",
					TraversalIDs:  "1,2",
					ParentID:      1,
					FullPath:      "/a/b",
					FullName:      "1 / 2",
					ChildrenCount: 1,
					Type:          ChildType,
					Children: []*Child{
						{
							ID:           3,
							Name:         "3",
							Path:         "c",
							TraversalIDs: "1,2,3",
							ParentID:     2,
							FullPath:     "/a/b/c",
							FullName:     "1 / 2 / 3",
							Type:         ChildType,
						},
					},
				},
				{
					ID:            4,
					Name:          "4",
					Path:          "d",
					TraversalIDs:  "1,4",
					ParentID:      1,
					FullPath:      "/a/d",
					FullName:      "1 / 4",
					ChildrenCount: 1,
					Type:          ChildType,
					Children: []*Child{
						{
							ID:           5,
							Name:         "5",
							Path:         "e",
							TraversalIDs: "1,4,5",
							ParentID:     4,
							FullPath:     "/a/d/e",
							FullName:     "1 / 4 / 5",
							Type:         ChildType,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateChildrenWithLevelStruct(tt.args.groupID, tt.args.groups); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateChildrenWithLevelStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}
