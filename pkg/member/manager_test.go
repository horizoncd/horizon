package member

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/member/models"
	"github.com/stretchr/testify/assert"
)

var (
	db, _ = orm.NewSqliteDB("")
	ctx   context.Context
	mgr   = New(db)
)

func MemberValueEqual(member1, member2 *models.Member) bool {
	if member2.ResourceType == member1.ResourceType &&
		member1.ResourceID == member2.ResourceID &&
		member1.Role == member2.Role &&
		member1.MemberType == member2.MemberType &&
		member1.MemberNameID == member2.MemberNameID &&
		member1.GrantedBy == member2.GrantedBy {
		return true
	}
	return false
}

// nolint
func TestBasic(t *testing.T) {
	var grantedByAdmin uint = 0
	member1 := &models.Member{
		ResourceType: "group",
		ResourceID:   1234324,
		Role:         "owner",
		MemberType:   models.MemberUser,
		MemberNameID: 1,
		GrantedBy:    grantedByAdmin,
	}

	// test create
	member, err := mgr.Create(ctx, member1)
	assert.Nil(t, err)

	b, err := json.Marshal(member)
	assert.Nil(t, err)

	t.Logf(string(b))

	retMember, err := mgr.GetByID(ctx, member.ID)
	assert.Nil(t, err)
	assert.True(t, MemberValueEqual(retMember, member))

	// test update
	var grantedByCat uint = 3
	member1.Role = "maintainer"
	member1.GrantedBy = grantedByAdmin
	var grandUser userauth.User = &userauth.DefaultInfo{
		Name:     "cat",
		FullName: "cat",
		ID:       grantedByCat,
	}
	ctx = context.WithValue(ctx, common.UserContextKey(), grandUser)

	retMember2, err := mgr.UpdateByID(ctx, retMember.ID, member1.Role)
	assert.Nil(t, err)

	member1.GrantedBy = grantedByCat
	assert.True(t, MemberValueEqual(retMember2, member1))

	retMember3, err := mgr.Get(ctx, member1.ResourceType, member1.ResourceID, models.MemberUser, member1.MemberNameID)
	assert.Nil(t, err)
	assert.True(t, MemberValueEqual(retMember2, retMember3))

	// test delete
	assert.Nil(t, mgr.DeleteMember(ctx, retMember3.ID))
	retMember4, err := mgr.Get(ctx, member1.ResourceType, member1.ResourceID, models.MemberUser, member1.MemberNameID)
	assert.Nil(t, err)
	assert.Nil(t, retMember4)
}

func TestList(t *testing.T) {
	var grantedByAdmin uint

	member1 := &models.Member{
		ResourceType: "group",
		ResourceID:   123456,
		Role:         "owner",
		MemberType:   models.MemberUser,
		MemberNameID: 1,
		GrantedBy:    grantedByAdmin,
	}

	// create 1
	retMember1, err := mgr.Create(ctx, member1)
	assert.Nil(t, err)
	assert.True(t, MemberValueEqual(member1, retMember1))

	// create 2
	member2 := &models.Member{
		ResourceType: "group",
		ResourceID:   123456,
		Role:         "owner",
		MemberType:   models.MemberUser,
		MemberNameID: 2,
		GrantedBy:    grantedByAdmin,
	}
	retMember2, err := mgr.Create(ctx, member2)
	assert.Nil(t, err)
	assert.True(t, MemberValueEqual(member2, retMember2))

	members, err := mgr.ListDirectMember(ctx, member1.ResourceType, member1.ResourceID)
	assert.Nil(t, err)
	assert.Equal(t, len(members), 2)
	assert.True(t, MemberValueEqual(&members[0], retMember1))
	assert.True(t, MemberValueEqual(&members[1], retMember2))
}

func TestListResourceOfMemberInfo(t *testing.T) {
	var grantedByAdmin uint

	member1 := &models.Member{
		ResourceType: models.TypeGroup,
		ResourceID:   11,
		Role:         "owner",
		MemberType:   models.MemberUser,
		MemberNameID: 1,
		GrantedBy:    grantedByAdmin,
	}

	// create 1
	retMember1, err := mgr.Create(ctx, member1)
	assert.Nil(t, err)
	assert.True(t, MemberValueEqual(member1, retMember1))

	// create 2
	member2 := &models.Member{
		ResourceType: models.TypeGroup,
		ResourceID:   22,
		Role:         "owner",
		MemberType:   models.MemberUser,
		MemberNameID: 1,
		GrantedBy:    grantedByAdmin,
	}
	retMember2, err := mgr.Create(ctx, member2)
	assert.Nil(t, err)
	assert.True(t, MemberValueEqual(member2, retMember2))

	resourceIDs, err := mgr.ListResourceOfMemberInfo(ctx, models.TypeGroup, 1)
	assert.Nil(t, err)
	t.Logf("%v", resourceIDs)
	assert.Equal(t, 2, len(resourceIDs))
	assert.Equal(t, uint(11), resourceIDs[0])
	assert.Equal(t, uint(22), resourceIDs[1])
}

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&models.Member{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	os.Exit(m.Run())
}
