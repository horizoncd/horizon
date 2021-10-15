package member

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/member/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func MemberValueEqual(member1, member2 *models.Member) bool {
	if member2.ResourceType == member1.ResourceType &&
		member1.ResourceID == member2.ResourceID &&
		member1.Role == member2.Role &&
		member1.MemberType == member2.MemberType &&
		member1.MemberInfo == member2.MemberInfo &&
		member1.GrantBy == member2.GrantBy {
		return true
	}
	return false
}

// nolint
func TestBasic(t *testing.T) {
	var grandByadmin uint = 0
	member1 := &models.Member{
		ResourceType: "group",
		ResourceID:   1234324,
		Role:         "owner",
		MemberType:   models.MemberUser,
		MemberInfo:   1,
		GrantBy:      grandByadmin,
	}

	// test create
	member, err := Mgr.Create(ctx, member1)
	assert.Nil(t, err)

	b, err := json.Marshal(member)
	assert.Nil(t, err)

	t.Logf(string(b))

	retMember, err := Mgr.GetByID(ctx, member.ID)
	assert.Nil(t, err)
	assert.True(t, MemberValueEqual(retMember, member))

	// test update
	var grandByCat uint = 3
	member1.Role = "maintainer"
	member1.GrantBy = grandByadmin
	var grandUser userauth.User = &userauth.DefaultInfo{
		Name:     "cat",
		FullName: "cat",
		ID:       grandByCat,
	}
	ctx = context.WithValue(ctx, user.Key(), grandUser)

	retMember2, err := Mgr.UpdateByID(ctx, retMember.ID, member1.Role)
	assert.Nil(t, err)

	member1.GrantBy = grandByCat
	assert.True(t, MemberValueEqual(retMember2, member1))

	retMember3, err := Mgr.Get(ctx, member1.ResourceType, member1.ResourceID, models.MemberUser, member1.MemberInfo)
	assert.Nil(t, err)
	assert.True(t, MemberValueEqual(retMember2, retMember3))

	// test delete
	assert.Nil(t, Mgr.DeleteMember(ctx, retMember3.ID))
	retMember4, err := Mgr.Get(ctx, member1.ResourceType, member1.ResourceID, models.MemberUser, member1.MemberInfo)
	assert.Nil(t, err)
	assert.Nil(t, retMember4)
}

func TestList(t *testing.T) {
	var grandByadmin uint = 0

	member1 := &models.Member{
		ResourceType: "group",
		ResourceID:   123456,
		Role:         "owner",
		MemberType:   models.MemberUser,
		MemberInfo:   1,
		GrantBy:      grandByadmin,
	}

	// create 1
	retMember1, err := Mgr.Create(ctx, member1)
	assert.Nil(t, err)
	assert.True(t, MemberValueEqual(member1, retMember1))

	// create 2
	member2 := &models.Member{
		ResourceType: "group",
		ResourceID:   123456,
		Role:         "owner",
		MemberType:   models.MemberUser,
		MemberInfo:   2,
		GrantBy:      grandByadmin,
	}
	retMember2, err := Mgr.Create(ctx, member2)
	assert.Nil(t, err)
	assert.True(t, MemberValueEqual(member2, retMember2))

	members, err := Mgr.ListDirectMember(ctx, member1.ResourceType, member1.ResourceID)
	assert.Nil(t, err)
	assert.Equal(t, len(members), 2)
	assert.True(t, MemberValueEqual(&members[0], retMember1))
	assert.True(t, MemberValueEqual(&members[1], retMember2))
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Member{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
