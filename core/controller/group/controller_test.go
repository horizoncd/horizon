package group

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	membermock "g.hz.netease.com/horizon/mock/pkg/member/service"
	applicationdao "g.hz.netease.com/horizon/pkg/application/dao"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/group/manager"
	"g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/pkg/group/service"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	// use tmp sqlite
	db, _                    = orm.NewSqliteDB("")
	ctx                      = orm.NewContext(context.TODO(), db)
	contextUserID       uint = 1
	contextUserName          = "Tony"
	contextUserFullName      = "TonyWu"
	groupCtl                 = NewController(nil)
)

func GroupValueEqual(g1, g2 *models.Group) bool {
	if g1.Name == g2.Name && g1.Path == g2.Path &&
		g1.VisibilityLevel == g2.VisibilityLevel &&
		g1.Description == g2.Description &&
		g1.ParentID == g2.ParentID &&
		g1.TraversalIDs == g2.TraversalIDs &&
		g1.CreatedBy == g2.CreatedBy &&
		g1.UpdatedBy == g2.UpdatedBy {
		return true
	}

	return false
}

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
}

func TestGetAuthedGroups(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	memberMock := membermock.NewMockService(mockCtrl)
	myGroupCtl := NewController(memberMock)

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
			wantErr: false,
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
			wantErr: false,
		},
		{
			name: "createSubGroup",
			args: args{
				ctx: ctx,
				newGroup: &NewGroup{
					Name:            "2",
					Path:            "b",
					VisibilityLevel: "private",
					ParentID:        1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := myGroupCtl.CreateGroup(tt.args.ctx, tt.args.newGroup)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				group, _ := manager.Mgr.GetByID(ctx, got)
				var traversalIDs string
				if group.ParentID == 0 {
					traversalIDs = strconv.Itoa(int(got))
				} else {
					parent, _ := manager.Mgr.GetByID(ctx, tt.args.newGroup.ParentID)
					traversalIDs = fmt.Sprintf("%s,%d", parent.TraversalIDs, got)
				}

				assert.True(t, GroupValueEqual(group, &models.Group{
					Name:            tt.args.newGroup.Name,
					Path:            tt.args.newGroup.Path,
					Description:     tt.args.newGroup.Description,
					ParentID:        tt.args.newGroup.ParentID,
					VisibilityLevel: tt.args.newGroup.VisibilityLevel,
					TraversalIDs:    traversalIDs,
					CreatedBy:       1,
					UpdatedBy:       1,
				}))
			}
		})
	}
	// case admin get all the groups
	rootUserContext := context.WithValue(ctx, user.Key(), &userauth.DefaultInfo{ // nolint
		Name:     contextUserName,
		FullName: contextUserFullName,
		ID:       contextUserID,
		Admin:    true,
	})
	groups, err := groupCtl.ListAuthedGroup(rootUserContext)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(groups))

	// case normal user get same groups
	normalUserContext := context.WithValue(ctx, user.Key(), &userauth.DefaultInfo{ // nolint
		Name:     contextUserName,
		FullName: contextUserFullName,
		ID:       contextUserID,
		Admin:    false,
	})
	memberMock.EXPECT().GetMemberOfResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&membermodels.Member{
			Model:        gorm.Model{},
			ResourceType: "",
			ResourceID:   0,
			Role:         "",
			MemberType:   0,
			MemberNameID: 0,
			GrantedBy:    0,
			CreatedBy:    0,
		}, nil).Times(1)
	memberMock.EXPECT().GetMemberOfResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&membermodels.Member{
			Model:        gorm.Model{},
			ResourceType: "",
			ResourceID:   0,
			Role:         role.Owner,
			MemberType:   0,
			MemberNameID: 0,
			GrantedBy:    0,
			CreatedBy:    0,
		}, nil).Times(1)
	memberMock.EXPECT().GetMemberOfResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&membermodels.Member{
			Model:        gorm.Model{},
			ResourceType: "",
			ResourceID:   0,
			Role:         role.Maintainer,
			MemberType:   0,
			MemberNameID: 0,
			GrantedBy:    0,
			CreatedBy:    0,
		}, nil).Times(1)
	groups, err = myGroupCtl.ListAuthedGroup(normalUserContext)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(groups))

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := groupCtl.CreateGroup(tt.args.ctx, tt.args.newGroup)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				group, _ := manager.Mgr.GetByID(ctx, got)
				var traversalIDs string
				if group.ParentID == 0 {
					traversalIDs = strconv.Itoa(int(got))
				} else {
					parent, _ := manager.Mgr.GetByID(ctx, tt.args.newGroup.ParentID)
					traversalIDs = fmt.Sprintf("%s,%d", parent.TraversalIDs, got)
				}

				assert.True(t, GroupValueEqual(group, &models.Group{
					Name:            tt.args.newGroup.Name,
					Path:            tt.args.newGroup.Path,
					Description:     tt.args.newGroup.Description,
					ParentID:        tt.args.newGroup.ParentID,
					VisibilityLevel: tt.args.newGroup.VisibilityLevel,
					TraversalIDs:    traversalIDs,
					CreatedBy:       1,
					UpdatedBy:       1,
				}))
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

	id, err := groupCtl.CreateGroup(ctx, newRootGroup)
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
			if err := groupCtl.Delete(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
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

	id, err := groupCtl.CreateGroup(ctx, newRootGroup)
	assert.Nil(t, err)

	group, err := manager.Mgr.GetByID(ctx, id)

	assert.Nil(t, err)

	type args struct {
		ctx context.Context
		id  uint
	}
	tests := []struct {
		name    string
		args    args
		want    *service.Child
		wantErr bool
	}{
		{
			name: "getExist",
			args: args{
				ctx: ctx,
				id:  id,
			},
			want: &service.Child{
				ID:              id,
				Name:            "1",
				Path:            "a",
				VisibilityLevel: "private",
				UpdatedAt:       group.UpdatedAt,
				ParentID:        0,
				FullPath:        "/a",
				FullName:        "1",
				TraversalIDs:    strconv.Itoa(int(id)),
				Type:            service.ChildTypeGroup,
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
			got, err := groupCtl.GetByID(tt.args.ctx, tt.args.id)
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

	id, err := groupCtl.CreateGroup(ctx, newRootGroup)
	assert.Nil(t, err)
	child, err := groupCtl.GetByID(ctx, id)
	assert.Nil(t, err)
	applicationDAO := applicationdao.NewDAO()
	app, err := applicationDAO.Create(ctx, &appmodels.Application{
		GroupID:     id,
		Name:        "app",
		Description: "this is a description",
		Priority:    "P0",
	}, nil)
	assert.Nil(t, err)

	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *service.Child
		wantErr bool
	}{
		{
			name: "getExist",
			args: args{
				ctx:  ctx,
				path: "/a",
			},
			want: &service.Child{
				ID:              id,
				Name:            "1",
				Path:            "a",
				VisibilityLevel: "private",
				ParentID:        0,
				TraversalIDs:    strconv.Itoa(int(id)),
				FullPath:        "/a",
				FullName:        "1",
				Type:            service.ChildTypeGroup,
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
			want: &service.Child{
				ID:          app.ID,
				Name:        "app",
				Path:        "app",
				Description: "this is a description",
				ParentID:    id,
				FullName:    "1/app",
				FullPath:    "/a/app",
				Type:        service.ChildTypeApplication,
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
			got, err := groupCtl.GetByFullPath(tt.args.ctx, tt.args.path)
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
	id, _ := groupCtl.CreateGroup(ctx, newRootGroup)

	newGroup := &NewGroup{
		Name:     "2",
		Path:     "b",
		ParentID: id,
	}
	id2, _ := groupCtl.CreateGroup(ctx, newGroup)
	group2, _ := groupCtl.GetByID(ctx, id2)

	applicationDAO := applicationdao.NewDAO()
	app, err := applicationDAO.Create(ctx, &appmodels.Application{
		GroupID: id,
		Name:    "c",
	}, nil)
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
		want    []*service.Child
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
			want: []*service.Child{
				{
					ID:        id2,
					Name:      "2",
					Path:      "b",
					FullPath:  "/a/b",
					FullName:  "1/2",
					Type:      service.ChildTypeGroup,
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
			want: []*service.Child{
				{
					ID:        app.ID,
					Name:      "c",
					Path:      "c",
					FullPath:  "/a/c",
					FullName:  "1/c",
					Type:      service.ChildTypeApplication,
					UpdatedAt: app.UpdatedAt,
				},
			},
			want1: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := groupCtl.GetChildren(tt.args.ctx, tt.args.id, tt.args.pageNumber, tt.args.pageSize)
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
	id, _ := groupCtl.CreateGroup(ctx, newRootGroup)
	group1, _ := groupCtl.GetByID(ctx, id)

	newGroup := &NewGroup{
		Name:            "2",
		Path:            "b",
		VisibilityLevel: "private",
		ParentID:        id,
	}
	id2, _ := groupCtl.CreateGroup(ctx, newGroup)
	group2, _ := groupCtl.GetByID(ctx, id2)

	type args struct {
		ctx        context.Context
		id         uint
		pageNumber int
		pageSize   int
	}
	tests := []struct {
		name    string
		args    args
		want    []*service.Child
		want1   int64
		wantErr bool
	}{
		{
			name: "getRootSubGroups",
			args: args{
				ctx: ctx,
			},
			want: []*service.Child{
				{
					ID:              id,
					Name:            "1",
					Path:            "a",
					VisibilityLevel: "private",
					ParentID:        0,
					TraversalIDs:    strconv.Itoa(int(id)),
					FullPath:        "/a",
					FullName:        "1",
					Type:            service.ChildTypeGroup,
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
			want: []*service.Child{
				{
					ID:              id2,
					Name:            "2",
					Path:            "b",
					VisibilityLevel: "private",
					ParentID:        id,
					TraversalIDs:    strconv.Itoa(int(id)) + "," + strconv.Itoa(int(id2)),
					FullPath:        "/a/b",
					FullName:        "1/2",
					Type:            service.ChildTypeGroup,
					ChildrenCount:   0,
					UpdatedAt:       group2.UpdatedAt,
				},
			},
			want1: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := groupCtl.GetSubGroups(tt.args.ctx, tt.args.id, tt.args.pageNumber, tt.args.pageSize)
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
	id, _ := groupCtl.CreateGroup(ctx, newRootGroup)

	newGroup := &NewGroup{
		Name:     "2",
		Path:     "b",
		ParentID: id,
	}
	id2, _ := groupCtl.CreateGroup(ctx, newGroup)
	_, _ = groupCtl.GetByID(ctx, id2)

	applicationDAO := applicationdao.NewDAO()
	app, err := applicationDAO.Create(ctx, &appmodels.Application{
		GroupID: id,
		Name:    "c",
	}, nil)
	assert.Nil(t, err)

	type args struct {
		ctx    context.Context
		id     uint
		filter string
	}
	tests := []struct {
		name    string
		args    args
		want    []*service.Child
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
			want: []*service.Child{
				{
					ID:       id2,
					Name:     "2",
					Path:     "b",
					FullPath: "/a/b",
					FullName: "1/2",
					Type:     service.ChildTypeGroup,
				},
				{
					ID:       app.ID,
					Name:     app.Name,
					Path:     app.Name,
					FullPath: "/a/c",
					FullName: "1/c",
					Type:     service.ChildTypeApplication,
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
			want: []*service.Child{
				{
					ID:       app.ID,
					Name:     app.Name,
					Path:     app.Name,
					FullPath: "/a/c",
					FullName: "1/c",
					Type:     service.ChildTypeApplication,
					ParentID: id,
				},
			},
			want1: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := groupCtl.SearchChildren(tt.args.ctx, &SearchParams{
				GroupID:    tt.args.id,
				Filter:     tt.args.filter,
				PageNumber: 1,
				PageSize:   10,
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
	id, _ := groupCtl.CreateGroup(ctx, newRootGroup)
	group1, _ := groupCtl.GetByID(ctx, id)

	newGroup := &NewGroup{
		Name:            "2",
		Path:            "b",
		VisibilityLevel: "private",
		ParentID:        id,
	}
	id2, _ := groupCtl.CreateGroup(ctx, newGroup)
	group2, _ := groupCtl.GetByID(ctx, id2)

	type args struct {
		ctx    context.Context
		id     uint
		filter string
	}
	tests := []struct {
		name    string
		args    args
		want    []*service.Child
		want1   int64
		wantErr bool
	}{
		{
			name: "blankFilter",
			args: args{
				ctx:    ctx,
				filter: "",
			},
			want: []*service.Child{
				{
					ID:              id,
					Name:            "1",
					Path:            "a",
					VisibilityLevel: "private",
					ParentID:        0,
					TraversalIDs:    strconv.Itoa(int(id)),
					FullPath:        "/a",
					FullName:        "1",
					Type:            service.ChildTypeGroup,
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
			want:  []*service.Child{},
			want1: 0,
		},
		{
			name: "matchFirstLevel",
			args: args{
				ctx:    ctx,
				filter: "1",
			},
			want: []*service.Child{
				{
					ID:              id,
					Name:            "1",
					Path:            "a",
					VisibilityLevel: "private",
					ParentID:        0,
					TraversalIDs:    strconv.Itoa(int(id)),
					FullPath:        "/a",
					FullName:        "1",
					Type:            service.ChildTypeGroup,
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
			want: []*service.Child{
				{
					ID:              id,
					Name:            "1",
					Path:            "a",
					VisibilityLevel: "private",
					ParentID:        0,
					TraversalIDs:    strconv.Itoa(int(id)),
					FullPath:        "/a",
					FullName:        "1",
					Type:            service.ChildTypeGroup,
					ChildrenCount:   1,
					UpdatedAt:       group1.UpdatedAt,
					Children: []*service.Child{
						{
							ID:              id2,
							Name:            "2",
							Path:            "b",
							VisibilityLevel: "private",
							ParentID:        id,
							TraversalIDs:    strconv.Itoa(int(id)) + "," + strconv.Itoa(int(id2)),
							FullPath:        "/a/b",
							FullName:        "1/2",
							Type:            service.ChildTypeGroup,
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
			got, got1, err := groupCtl.SearchGroups(tt.args.ctx, &SearchParams{
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
	id, _ := groupCtl.CreateGroup(ctx, newRootGroup)
	newGroup := &NewGroup{
		Name:            "2",
		Path:            "b",
		VisibilityLevel: "private",
		ParentID:        id,
	}
	id2, _ := groupCtl.CreateGroup(ctx, newGroup)

	newRootGroup2 := &NewGroup{
		Name:            "3",
		Path:            "c",
		VisibilityLevel: "private",
	}
	id3, _ := groupCtl.CreateGroup(ctx, newRootGroup2)

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
				// nolint
				ctx: context.WithValue(ctx, user.Key(), &userauth.DefaultInfo{
					ID: 2,
				}),
				id:          id,
				newParentID: id3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := groupCtl.Transfer(tt.args.ctx, tt.args.id, tt.args.newParentID); (err != nil) != tt.wantErr {
				t.Errorf("Transfer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// check transfer success
	g, err := manager.Mgr.GetByID(ctx, id)

	assert.Nil(t, err)
	assert.Equal(t, id3, g.ParentID)
	assert.Equal(t, strconv.Itoa(int(id3))+","+strconv.Itoa(int(id)), g.TraversalIDs)
	assert.True(t, g.UpdatedBy == 2)

	g, err = manager.Mgr.GetByID(ctx, id2)
	assert.Nil(t, err)
	assert.Equal(t, strconv.Itoa(int(id3))+","+strconv.Itoa(int(id))+","+strconv.Itoa(int(id2)), g.TraversalIDs)
	assert.True(t, g.UpdatedBy == 2)

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
}

func TestControllerUpdateBasic(t *testing.T) {
	newRootGroup := &NewGroup{
		Name:            "1",
		Path:            "a",
		VisibilityLevel: "private",
	}
	id, _ := groupCtl.CreateGroup(ctx, newRootGroup)

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
				// nolint
				ctx: context.WithValue(ctx, user.Key(), &userauth.DefaultInfo{
					ID: 2,
				}),
				id: id,
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
			if err := groupCtl.UpdateBasic(tt.args.ctx, tt.args.id, tt.args.updateGroup); (err != nil) != tt.wantErr {
				t.Errorf("UpdateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
			group, _ := manager.Mgr.GetByID(ctx, tt.args.id)

			if group.ID > 0 {
				assert.True(t, GroupValueEqual(group, &models.Group{
					Name:            tt.args.updateGroup.Name,
					Path:            tt.args.updateGroup.Path,
					Description:     tt.args.updateGroup.Description,
					ParentID:        group.ParentID,
					VisibilityLevel: tt.args.updateGroup.VisibilityLevel,
					TraversalIDs:    group.TraversalIDs,
					CreatedBy:       1,
					UpdatedBy:       2,
				}))
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
		want []*service.Child
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
			want: []*service.Child{},
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
			want: []*service.Child{
				{
					ID:            2,
					Name:          "2",
					Path:          "b",
					TraversalIDs:  "1,2",
					ParentID:      1,
					FullPath:      "/a/b",
					FullName:      "1/2",
					ChildrenCount: 2,
					Type:          service.ChildTypeGroup,
					Children: []*service.Child{
						{
							ID:           3,
							Name:         "3",
							Path:         "c",
							TraversalIDs: "1,2,3",
							ParentID:     2,
							FullPath:     "/a/b/c",
							FullName:     "1/2/3",
							Type:         service.ChildTypeGroup,
						},
						{
							ID:       6,
							Name:     "f",
							Path:     "f",
							ParentID: 2,
							FullPath: "/a/b/f",
							FullName: "1/2/f",
							Type:     service.ChildTypeApplication,
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
					FullName:      "1/4",
					ChildrenCount: 1,
					Type:          service.ChildTypeGroup,
					Children: []*service.Child{
						{
							ID:           5,
							Name:         "5",
							Path:         "e",
							TraversalIDs: "1,4,5",
							ParentID:     4,
							FullPath:     "/a/d/e",
							FullName:     "1/4/5",
							Type:         service.ChildTypeGroup,
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
}
