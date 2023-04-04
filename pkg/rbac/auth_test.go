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

package rbac

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/horizoncd/horizon/pkg/models"
	"github.com/stretchr/testify/assert"

	"github.com/horizoncd/horizon/core/common"
	servicemock "github.com/horizoncd/horizon/mock/pkg/member/service"
	rolemock "github.com/horizoncd/horizon/mock/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/auth"
	"github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/rbac/types"
)

// members and pipelineruns are allowed
var (
	defaultUser = &user.DefaultInfo{
		Name:     "tom",
		FullName: "tomsun",
		ID:       1,
	}
	ctx = common.WithContext(context.Background(), defaultUser)
)

// nolint
func TestAuthMember(t *testing.T) {
	mockCtl := gomock.NewController(t)
	memberServiceMock := servicemock.NewMockMemberService(mockCtl)
	roleServiceMock := rolemock.NewMockService(mockCtl)
	testAuthorizer := Authorizer(&authorizer{
		roleService:   roleServiceMock,
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

	ctx = context.WithValue(ctx, common.UserContextKey(), defaultUser)
	decision, reason, err := testAuthorizer.Authorize(ctx, authRecord)
	assert.Nil(t, err)
	assert.Equal(t, auth.DecisionAllow, decision)
	assert.Equal(t, NotChecked, reason)

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
	roleServiceMock.EXPECT().GetDefaultRole(ctx).Return(nil).Times(1)
	decision, reason, err = testAuthorizer.Authorize(ctx, authRecord)
	assert.Equal(t, auth.DecisionDeny, decision)
	assert.Equal(t, MemberNotExist, reason)
	assert.Nil(t, err)
}

// nolint
func TestAuthRole(t *testing.T) {
	mockCtl := gomock.NewController(t)
	memberServiceMock := servicemock.NewMockMemberService(mockCtl)
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

	// get role ok and denied
	roleServieMock.EXPECT().GetRole(ctx,
		gomock.Any()).Return(&types.Role{
		Name:        "owner",
		PolicyRules: nil,
	}, nil).Times(1)
	decision, reason, err = testAuthorizer.Authorize(ctx, authRecord)
	assert.Equal(t, auth.DecisionDeny, decision)
	assert.Nil(t, err)
}
