package group

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

	"g.hz.netease.com/horizon/core/controller/member"
	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	applicationdao "g.hz.netease.com/horizon/pkg/application/dao"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/group/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	usermodel "g.hz.netease.com/horizon/pkg/user/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	// use tmp sqlite
	db, _                      = orm.NewSqliteDB("")
	ctx                        = orm.NewContext(context.TODO(), db)
	contextUserID       uint   = 1
	contextUserName     string = "Tony"
	contextUserFullName string = "TonyWu"
)

// nolint
func init() {
	ctx = context.WithValue(ctx, user.Key(), &userauth.DefaultInfo{
		Name:     contextUserName,
		FullName: contextUserFullName,
		ID:       contextUserID,
	})
	// create table
	err := db.AutoMigrate(&models.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	err = db.AutoMigrate(&appmodels.Application{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	err = db.AutoMigrate(&membermodels.Member{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}

	err = db.AutoMigrate(&usermodel.User{})
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
				Type:            ChildTypeGroup,
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
	applicationDAO := applicationdao.NewDAO()
	app, err := applicationDAO.Create(ctx, &appmodels.Application{
		GroupID:     id,
		Name:        "app",
		Description: "this is a description",
		Priority:    "P0",
	})
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
				Type:            ChildTypeGroup,
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
		}, {
			name: "applicationExist",
			args: args{
				ctx:  ctx,
				path: "/a/app",
			},
			want: &Child{
				ID:          app.ID,
				Name:        "app",
				Path:        "app",
				Description: "this is a description",
				ParentID:    id,
				FullName:    "1 / app",
				FullPath:    "/a/app",
				Type:        ChildTypeApplication,
			},
		}, {
			name: "applicationNotExist",
			args: args{
				ctx:  ctx,
				path: "/a/app-not-exists",
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
	newRootGroup := &NewGroup{
		Name: "1",
		Path: "a",
	}
	id, _ := Ctl.CreateGroup(ctx, newRootGroup)

	newGroup := &NewGroup{
		Name:     "2",
		Path:     "b",
		ParentID: id,
	}
	id2, _ := Ctl.CreateGroup(ctx, newGroup)
	group2, _ := Ctl.GetByID(ctx, id2)

	applicationDAO := applicationdao.NewDAO()
	app, err := applicationDAO.Create(ctx, &appmodels.Application{
		GroupID: id,
		Name:    "c",
	})
	assert.Nil(t, err)

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
			name: "firstPage",
			args: args{
				ctx:        ctx,
				id:         id,
				pageSize:   1,
				pageNumber: 1,
			},
			want: []*Child{
				{
					ID:        id2,
					Name:      "2",
					Path:      "b",
					FullPath:  "/a/b",
					FullName:  "1 / 2",
					Type:      ChildTypeGroup,
					UpdatedAt: group2.UpdatedAt,
				},
			},
			want1: 2,
		},
		{
			name: "secondPage",
			args: args{
				ctx:        ctx,
				id:         id,
				pageSize:   1,
				pageNumber: 2,
			},
			want: []*Child{
				{
					ID:        app.ID,
					Name:      "c",
					Path:      "c",
					FullPath:  "/a/c",
					FullName:  "1 / c",
					Type:      ChildTypeApplication,
					UpdatedAt: app.UpdatedAt,
				},
			},
			want1: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Ctl.GetChildren(tt.args.ctx, tt.args.id, tt.args.pageNumber, tt.args.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChildren() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, val := range got {
				val.UpdatedAt = tt.want[i].UpdatedAt
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChildren() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetChildren() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&appmodels.Application{})
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
					Type:            ChildTypeGroup,
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
					Type:            ChildTypeGroup,
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
	newRootGroup := &NewGroup{
		Name: "1",
		Path: "a",
	}
	id, _ := Ctl.CreateGroup(ctx, newRootGroup)

	newGroup := &NewGroup{
		Name:     "2",
		Path:     "b",
		ParentID: id,
	}
	id2, _ := Ctl.CreateGroup(ctx, newGroup)
	_, _ = Ctl.GetByID(ctx, id2)

	applicationDAO := applicationdao.NewDAO()
	app, err := applicationDAO.Create(ctx, &appmodels.Application{
		GroupID: id,
		Name:    "c",
	})
	assert.Nil(t, err)

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
				id:     id,
			},
			want: []*Child{
				{
					ID:       id2,
					Name:     "2",
					Path:     "b",
					FullPath: "/a/b",
					FullName: "1 / 2",
					Type:     ChildTypeGroup,
				},
				{
					ID:       app.ID,
					Name:     app.Name,
					Path:     app.Name,
					FullPath: "/a/c",
					FullName: "1 / c",
					Type:     ChildTypeApplication,
				},
			},
			want1: 2,
		},
		{
			name: "matchApp",
			args: args{
				ctx:    ctx,
				filter: "c",
				id:     id,
			},
			want: []*Child{
				{
					ID:       app.ID,
					Name:     app.Name,
					Path:     app.Name,
					FullPath: "/a/c",
					FullName: "1 / c",
					Type:     ChildTypeApplication,
					ParentID: id,
				},
			},
			want1: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Ctl.SearchChildren(tt.args.ctx, &SearchParams{
				GroupID: tt.args.id,
				Filter:  tt.args.filter,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchChildren() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, val := range got {
				val.UpdatedAt = tt.want[i].UpdatedAt
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SearchChildren() got = %+v, want %+v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("SearchChildren() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
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
					Type:            ChildTypeGroup,
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
					Type:            ChildTypeGroup,
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
					Type:            ChildTypeGroup,
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
							Type:            ChildTypeGroup,
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
				GroupID: tt.args.id,
				Filter:  tt.args.filter,
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
		groupID      uint
		groups       []*models.Group
		applications []*appmodels.Application
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
				applications: []*appmodels.Application{},
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
				applications: []*appmodels.Application{
					{
						Model: gorm.Model{
							ID: 6,
						},
						Name:    "f",
						GroupID: 2,
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
					ChildrenCount: 2,
					Type:          ChildTypeGroup,
					Children: []*Child{
						{
							ID:           3,
							Name:         "3",
							Path:         "c",
							TraversalIDs: "1,2,3",
							ParentID:     2,
							FullPath:     "/a/b/c",
							FullName:     "1 / 2 / 3",
							Type:         ChildTypeGroup,
						},
						{
							ID:       6,
							Name:     "f",
							Path:     "f",
							ParentID: 2,
							FullPath: "/a/b/f",
							FullName: "1 / 2 / f",
							Type:     ChildTypeApplication,
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
					Type:          ChildTypeGroup,
					Children: []*Child{
						{
							ID:           5,
							Name:         "5",
							Path:         "e",
							TraversalIDs: "1,4,5",
							ParentID:     4,
							FullPath:     "/a/d/e",
							FullName:     "1 / 4 / 5",
							Type:         ChildTypeGroup,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateChildrenWithLevelStruct(tt.args.groupID, tt.args.groups,
				tt.args.applications); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateChildrenWithLevelStruct() = %+v, want %+v", got, tt.want)
			}
		})
	}
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&membermodels.Member{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&usermodel.User{})
}

func MemberSame(m1, m2 member.Member) bool {
	if m1.MemberInfo == m2.MemberInfo &&
		m1.MemberID == m2.MemberID &&
		m1.ResourceType == m2.ResourceType &&
		m1.ResourceID == m2.ResourceID &&
		m1.Role == m2.Role &&
		m1.GrantBy == m2.GrantBy {
		return true
	}
	return false
}

var (
	user1ID   uint   = 1
	user1Name string = contextUserName

	user2ID   uint   = 2
	user2Name string = "tom"

	user3Name string = "jerry"

	user4Name string = "alias"

	user5Name string = "henry"
)

func CreateUsers(t *testing.T) {
	// create user
	user1 := usermodel.User{
		Model: gorm.Model{},
		Name:  user1Name,
	}

	user2 := user1
	user2.Name = user2Name

	user3 := user1
	user3.Name = user3Name

	user4 := user1
	user4.Name = user4Name

	user5 := user1
	user5.Name = user5Name

	for _, user := range []usermodel.User{user1, user2, user3, user4, user5} {
		_, err := usermanager.Mgr.Create(ctx, &user)
		assert.Nil(t, err)
	}
}

func TestCreateGroupWithOwner(t *testing.T) {
	CreateUsers(t)
	// create group
	newGroup := &NewGroup{
		Name:            "1",
		Path:            "1",
		VisibilityLevel: "private",
		Description:     "i am a private Group",
		ParentID:        0,
	}

	groupID, err := Ctl.CreateGroup(ctx, newGroup)
	assert.Nil(t, err)

	retMembers, err := Ctl.ListMember(ctx, groupID)
	expectMember := member.Member{
		MemberType:   membermodels.MemberUser,
		MemberInfo:   contextUserName,
		MemberID:     contextUserID,
		ResourceType: membermodels.TypeGroup,
		ResourceID:   groupID,
		Role:         member.Owner,
		GrantBy:      contextUserID,
	}
	assert.Nil(t, err)
	assert.NotNil(t, retMembers)
	assert.True(t, MemberSame(retMembers[0], expectMember))

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&membermodels.Member{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&usermodel.User{})
}

func PostMemberAndMemberEqual(postMember member.PostMember, member2 member.Member) bool {
	return postMember.ResourceType == string(member2.ResourceType) &&
		postMember.ResourceID == member2.ResourceID &&
		postMember.MemberType == member2.MemberType &&
		postMember.MemberInfo == member2.MemberID &&
		postMember.Role == member2.Role
}

func TestCreateGetUpdateRemoveList(t *testing.T) {
	CreateUsers(t)

	// create group
	newGroup := &NewGroup{
		Name:            "1",
		Path:            "1",
		VisibilityLevel: "private",
		Description:     "i am a private Group",
		ParentID:        0,
	}

	groupID, err := Ctl.CreateGroup(ctx, newGroup)
	assert.Nil(t, err)

	// create member
	postMember2 := member.PostMember{
		ResourceType: membermodels.TypeGroupStr,
		ResourceID:   groupID,
		MemberInfo:   user2ID,
		MemberType:   membermodels.MemberUser,
		Role:         membermodels.Owner,
	}
	retMember2, err := Ctl.CreateMember(ctx, &postMember2)
	assert.Nil(t, err)
	assert.True(t, PostMemberAndMemberEqual(postMember2, *retMember2))

	// remove the member
	err = Ctl.RemoveMember(ctx, retMember2.ID)
	assert.Nil(t, err)

	// list member (create post2 and then list)
	retMember2, err = Ctl.CreateMember(ctx, &postMember2)
	assert.Nil(t, err)
	assert.True(t, PostMemberAndMemberEqual(postMember2, *retMember2))

	// create member already exist
	postMemberOwner := member.PostMember{
		ResourceType: membermodels.TypeGroupStr,
		ResourceID:   groupID,
		MemberInfo:   user1ID,
		MemberType:   membermodels.MemberUser,
		Role:         membermodels.Owner,
	}
	members, err := Ctl.ListMember(ctx, groupID)
	assert.Nil(t, err)
	assert.Equal(t, len(members), 2)
	assert.True(t, PostMemberAndMemberEqual(postMemberOwner, members[0]))
	assert.True(t, PostMemberAndMemberEqual(postMember2, members[1]))

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&membermodels.Member{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&usermodel.User{})
}
