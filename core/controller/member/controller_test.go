package member

import (
	"context"
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/controller/group"
	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
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

var (
	user1ID   uint   = 1
	user1Name string = contextUserName

	user2ID   uint   = 2
	user2Name string = "tom"

	user3Name string = "jerry"

	user4Name string = "alias"

	user5Name string = "henry"
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

func MemberSame(m1, m2 Member) bool {
	if m1.MemberName == m2.MemberName &&
		m1.MemberNameID == m2.MemberNameID &&
		m1.ResourceType == m2.ResourceType &&
		m1.ResourceID == m2.ResourceID &&
		m1.Role == m2.Role &&
		m1.GrantBy == m2.GrantBy {
		return true
	}
	return false
}

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
	newGroup := &group.NewGroup{
		Name:            "1",
		Path:            "1",
		VisibilityLevel: "private",
		Description:     "i am a private Group",
		ParentID:        0,
	}

	groupID, err := group.Ctl.CreateGroup(ctx, newGroup)
	assert.Nil(t, err)

	retMembers, err := Ctl.ListMember(ctx, membermodels.TypeGroupStr, groupID)
	expectMember := Member{
		MemberType:   membermodels.MemberUser,
		MemberName:   contextUserName,
		MemberNameID: contextUserID,
		ResourceType: membermodels.TypeGroup,
		ResourceID:   groupID,
		Role:         Owner,
		GrantBy:      contextUserID,
	}
	assert.Nil(t, err)
	assert.NotNil(t, retMembers)
	assert.True(t, MemberSame(retMembers[0], expectMember))

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&membermodels.Member{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&usermodel.User{})
}

func PostMemberAndMemberEqual(postMember PostMember, member2 Member) bool {
	return postMember.ResourceType == string(member2.ResourceType) &&
		postMember.ResourceID == member2.ResourceID &&
		postMember.MemberType == member2.MemberType &&
		postMember.MemberNameID == member2.MemberNameID &&
		postMember.Role == member2.Role
}

func TestCreateGetUpdateRemoveList(t *testing.T) {
	CreateUsers(t)

	// create group
	newGroup := &group.NewGroup{
		Name:            "1",
		Path:            "1",
		VisibilityLevel: "private",
		Description:     "i am a private Group",
		ParentID:        0,
	}

	groupID, err := group.Ctl.CreateGroup(ctx, newGroup)
	assert.Nil(t, err)

	// create member
	postMember2 := PostMember{
		ResourceType: membermodels.TypeGroupStr,
		ResourceID:   groupID,
		MemberNameID: user2ID,
		MemberType:   membermodels.MemberUser,
		Role:         membermodels.Owner,
	}
	retMember2, err := Ctl.CreateMember(ctx, &postMember2)
	assert.Nil(t, err)
	assert.True(t, PostMemberAndMemberEqual(postMember2, *retMember2))

	// update member
	retMember3, err := Ctl.UpdateMember(ctx, retMember2.ID, membermodels.Maitainer)
	assert.Nil(t, err)
	postMember2.Role = membermodels.Maitainer
	assert.True(t, PostMemberAndMemberEqual(postMember2, *retMember3))

	// remove the member
	err = Ctl.RemoveMember(ctx, retMember2.ID)
	assert.Nil(t, err)

	// list member (create post2 and then list)
	retMember2, err = Ctl.CreateMember(ctx, &postMember2)
	assert.Nil(t, err)
	assert.True(t, PostMemberAndMemberEqual(postMember2, *retMember2))

	postMemberOwner := PostMember{
		ResourceType: membermodels.TypeGroupStr,
		ResourceID:   groupID,
		MemberNameID: user1ID,
		MemberType:   membermodels.MemberUser,
		Role:         membermodels.Owner,
	}
	members, err := Ctl.ListMember(ctx, membermodels.TypeGroupStr, groupID)
	assert.Nil(t, err)
	assert.Equal(t, len(members), 2)
	assert.True(t, PostMemberAndMemberEqual(postMemberOwner, members[0]))
	assert.True(t, PostMemberAndMemberEqual(postMember2, members[1]))

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&membermodels.Member{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&usermodel.User{})
}
