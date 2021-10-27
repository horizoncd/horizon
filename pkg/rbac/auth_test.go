package rbac

import (
	"context"
	"errors"
	"testing"

	muser "g.hz.netease.com/horizon/core/middleware/user"
	servicemock "g.hz.netease.com/horizon/mock/pkg/member/service"
	rolemock "g.hz.netease.com/horizon/mock/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/auth"
	"g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// members and pipelineruns are allowed
var (
	defaultUser = &user.DefaultInfo{
		Name:     "tom",
		FullName: "tomsun",
		ID:       1,
	}
	ctx = context.TODO()
)

// nolint
func TestAuthMember(t *testing.T) {
	mockCtl := gomock.NewController(t)
	memberServiceMock := servicemock.NewMockService(mockCtl)
	roleServieMock := rolemock.NewMockService(mockCtl)
	testAuthorizer := Authorizer(&authorizer{
		roleService:   roleServieMock,
		memberService: memberServiceMock,
	})

	authRecord := auth.AttributesRecord{
		User:            defaultUser,
		Verb:            "delete",
		APIGroup:        "/apis/core",
		APIVersion:      "v1",
		Resource:        "members",
		ResourceRequest: true,
		Path:            "",
	}

	ctx = context.WithValue(ctx, muser.Key(), defaultUser)
	decision, reason, err := testAuthorizer.Authorize(ctx, authRecord)
	assert.Nil(t, err)
	assert.Equal(t, auth.DecisionAllow, decision)
	assert.Equal(t, NotChecked, reason)

	// phaseResourceError
	authRecord = auth.AttributesRecord{
		User:            defaultUser,
		Verb:            "delete",
		APIGroup:        "/apis/core",
		APIVersion:      "v1",
		Resource:        "groups",
		Name:            "groupID",
		ResourceRequest: true,
		Path:            "",
	}
	decision, reason, err = testAuthorizer.Authorize(ctx, authRecord)
	assert.Equal(t, auth.DecisionDeny, decision)
	assert.Equal(t, ResourceFormatErr, reason)
	assert.NotNil(t, err)

	authRecord = auth.AttributesRecord{
		User:            defaultUser,
		Verb:            "delete",
		APIGroup:        "/apis/core",
		APIVersion:      "v1",
		Resource:        "groups",
		Name:            "123",
		ResourceRequest: true,
		Path:            "",
	}
	// getMember error
	memberServiceMock.EXPECT().GetMemberOfResource(ctx, gomock.Any(),
		gomock.Any()).Return(nil, errors.New("error")).Times(1)
	decision, reason, err = testAuthorizer.Authorize(ctx, authRecord)
	assert.Equal(t, auth.DecisionDeny, decision)
	assert.Equal(t, InternalError, reason)
	assert.NotNil(t, err)

	// member not exist
	memberServiceMock.EXPECT().GetMemberOfResource(ctx, gomock.Any(),
		gomock.Any()).Return(nil, nil).Times(1)
	decision, reason, err = testAuthorizer.Authorize(ctx, authRecord)
	assert.Equal(t, auth.DecisionDeny, decision)
	assert.Equal(t, MemberNotExist, reason)
	assert.Nil(t, err)
}

// nolint
func TestAuthRole(t *testing.T) {
	mockCtl := gomock.NewController(t)
	memberServiceMock := servicemock.NewMockService(mockCtl)
	roleServieMock := rolemock.NewMockService(mockCtl)
	testAuthorizer := Authorizer(&authorizer{
		roleService:   roleServieMock,
		memberService: memberServiceMock,
	})
	authRecord := auth.AttributesRecord{
		User:            defaultUser,
		Verb:            "delete",
		APIGroup:        "/apis/core",
		APIVersion:      "v1",
		Resource:        "groups",
		Name:            "123",
		ResourceRequest: true,
		Path:            "",
	}

	// test getRole error
	memberServiceMock.EXPECT().GetMemberOfResource(ctx, gomock.Any(),
		gomock.Any()).Return(&models.Member{
		Role: "owner",
	}, nil).Times(3)

	roleServieMock.EXPECT().GetRole(ctx,
		gomock.Any()).Return(nil, errors.New("error")).Times(1)
	decision, reason, err := testAuthorizer.Authorize(ctx, authRecord)
	assert.Equal(t, auth.DecisionDeny, decision)
	assert.Equal(t, InternalError, reason)
	assert.NotNil(t, err)

	// test role not exist
	roleServieMock.EXPECT().GetRole(ctx,
		gomock.Any()).Return(nil, nil).Times(1)
	decision, reason, err = testAuthorizer.Authorize(ctx, authRecord)
	assert.Equal(t, auth.DecisionDeny, decision)
	assert.Equal(t, RoleNotExist, reason)
	assert.Nil(t, err)

	// get role ok and denyed
	roleServieMock.EXPECT().GetRole(ctx,
		gomock.Any()).Return(&types.Role{
		Name:        "owner",
		PolicyRules: nil,
	}, nil).Times(1)
	decision, reason, err = testAuthorizer.Authorize(ctx, authRecord)
	assert.Equal(t, auth.DecisionDeny, decision)
	assert.Nil(t, err)
}
