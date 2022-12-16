package accesstoken

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/horizoncd/horizon/core/common"
	herror "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/orm"
	applicationmodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	perror "github.com/horizoncd/horizon/pkg/errors"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	memberservice "github.com/horizoncd/horizon/pkg/member/service"
	"github.com/horizoncd/horizon/pkg/oauth/generate"
	oauthmanager "github.com/horizoncd/horizon/pkg/oauth/manager"
	oauthmodels "github.com/horizoncd/horizon/pkg/oauth/models"
	"github.com/horizoncd/horizon/pkg/oauth/store"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	roleservice "github.com/horizoncd/horizon/pkg/rbac/role"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	callbacks "github.com/horizoncd/horizon/pkg/util/ormcallbacks"
)

var (
	ctx context.Context
	c   Controller
)

// valid params
var (
	commonName         = "test"
	commonRole         = "owner"
	commonScopes       = []string{"applications:read-write"}
	commonExpiresAt    = time.Now().Add(time.Hour * 24).Format(ExpiresAtFormat)
	commonResourceType = "groups"
	commonResourceID   = uint(0)
)

// invalid params
const (
	expiredDate = "2021-10-1"
	invalidDate = "20222-11-2"
	invalidRole = "joker"
)

func TestMain(m *testing.M) {
	db, err := orm.NewSqliteDB("")
	if err != nil {
		panic(err)
	}
	callbacks.RegisterCustomCallbacks(db)

	manager := managerparam.InitManager(db)
	if err := db.AutoMigrate(
		&usermodels.User{},
		&membermodels.Member{},
		&oauthmodels.Token{},
		&groupmodels.Group{},
		&applicationmodels.Application{},
	); err != nil {
		panic(err)
	}

	roleSvc, err := roleservice.NewFileRole(context.Background(), strings.NewReader(roleConfig))
	if err != nil {
		panic(err)
	}

	authorizeCodeExpireIn := time.Second * 3
	accessTokenExpireIn := time.Hour * 24

	tokenStore := store.NewTokenStore(db)
	oauthAppStore := store.NewOauthAppStore(db)
	oauthMgr := oauthmanager.NewManager(oauthAppStore, tokenStore,
		generate.NewAuthorizeGenerate(), authorizeCodeExpireIn, accessTokenExpireIn)

	param := &param.Param{
		Manager:       manager,
		MemberService: memberservice.NewService(roleSvc, oauthMgr, manager),
	}

	ctx = context.TODO()

	user, err := manager.UserManager.Create(ctx, &usermodels.User{
		Name: "test",
	})
	if err != nil {
		panic(err)
	}
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{ // nolint
		Name: user.Name,
		ID:   user.ID,
	})

	group, err := manager.GroupManager.Create(ctx, &groupmodels.Group{
		Name: "test",
	})
	if err != nil {
		panic(err)
	}
	commonResourceID = group.ID

	c = NewController(param)

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	type PatCase struct {
		req              CreatePersonalAccessTokenRequest
		expectedCauseErr error
		expectedResp     *CreatePersonalAccessTokenResponse
	}
	type RatCase struct {
		req              CreateResourceAccessTokenRequest
		expectedCauseErr error
		expectedResp     *CreateResourceAccessTokenResponse
		resourceType     string
		resourceID       uint
	}

	patTestCases := []PatCase{
		{
			req: CreatePersonalAccessTokenRequest{
				Name:      commonName,
				Scopes:    commonScopes,
				ExpiresAt: expiredDate,
			},
			expectedCauseErr: herror.ErrParamInvalid,
		},
		{
			req: CreatePersonalAccessTokenRequest{
				Name:      commonName,
				Scopes:    commonScopes,
				ExpiresAt: invalidDate,
			},
			expectedCauseErr: herror.ErrParamInvalid,
		},
		{
			req: CreatePersonalAccessTokenRequest{
				Name:      commonName,
				Scopes:    commonScopes,
				ExpiresAt: NeverExpire,
			},
		},
	}

	ratTestCases := []RatCase{
		{
			req: CreateResourceAccessTokenRequest{
				CreatePersonalAccessTokenRequest: CreatePersonalAccessTokenRequest{
					Name:      commonName,
					Scopes:    commonScopes,
					ExpiresAt: expiredDate,
				},
				Role: commonRole,
			},
			resourceType:     commonResourceType,
			resourceID:       commonResourceID,
			expectedCauseErr: herror.ErrParamInvalid,
		},
		{
			req: CreateResourceAccessTokenRequest{
				CreatePersonalAccessTokenRequest: CreatePersonalAccessTokenRequest{
					Name:      commonName,
					Scopes:    commonScopes,
					ExpiresAt: NeverExpire,
				},
				Role: invalidRole,
			},
			resourceType:     commonResourceType,
			resourceID:       commonResourceID,
			expectedCauseErr: roleservice.ErrorRoleNotFound,
		},
		{
			req: CreateResourceAccessTokenRequest{
				CreatePersonalAccessTokenRequest: CreatePersonalAccessTokenRequest{
					Name:      commonName,
					Scopes:    commonScopes,
					ExpiresAt: commonExpiresAt,
				},
				Role: commonRole,
			},
			resourceType: commonResourceType,
			resourceID:   commonResourceID,
		},
	}

	for _, testCase := range patTestCases {
		var (
			createTokenResp *CreatePersonalAccessTokenResponse
			err             error
		)

		createTokenResp, err = c.CreatePersonalAccessToken(ctx, testCase.req)
		assert.Equal(t, testCase.expectedCauseErr, perror.Cause(err))
		if testCase.expectedResp != nil {
			assert.Equal(t, testCase.expectedResp, createTokenResp)
		}

		if testCase.expectedCauseErr == nil {
			_, total, err := c.ListPersonalAccessTokens(ctx, nil)
			assert.Equal(t, nil, err)
			assert.Equal(t, 1, total)

			err = c.RevokePersonalAccessToken(ctx, createTokenResp.ID)
			assert.Equal(t, nil, err)
		}
	}

	for _, testCase := range ratTestCases {
		var (
			createTokenResp *CreateResourceAccessTokenResponse
			err             error
		)

		createTokenResp, err = c.CreateResourceAccessToken(ctx, testCase.req, testCase.resourceType, testCase.resourceID)
		assert.Equal(t, testCase.expectedCauseErr, perror.Cause(err))
		if testCase.expectedResp != nil {
			assert.Equal(t, testCase.expectedResp, createTokenResp)
		}

		if testCase.expectedCauseErr == nil {
			_, total, err := c.ListResourceAccessTokens(ctx, testCase.resourceType, testCase.resourceID, nil)
			assert.Equal(t, nil, err)
			assert.Equal(t, 1, total)

			err = c.RevokeResourceAccessToken(ctx, createTokenResp.ID)
			assert.Equal(t, nil, err)
		}
	}
}

const roleConfig = `RolePriorityRankDesc:
  - pe
  - owner
  - maintainer
  - guest
DefaultRole: guest
Roles:
  - name: owner
    desc: 'owner为组/应用/集群的拥有者,拥有最高权限'
    rules:
      - apiGroups:
          - core
        resources:
          - applications
          - groups/applications
          - applications/members
          - applications/envtemplates
        verbs:
          - '*'
        scopes:
          - '*'
        nonResourceURLs:
          - '*'
      - apiGroups:
          - core
        resources:
          - groups
          - groups/members
          - groups/groups
        verbs:
          - '*'
        scopes:
          - '*'
      - apiGroups:
          - core
        resources:
          - applications/clusters
          - clusters
          - clusters/builddeploy
          - clusters/deploy
          - clusters/diffs
          - clusters/next
          - clusters/restart
          - clusters/rollback
          - clusters/status
          - clusters/members
          - clusters/pipelineruns
          - clusters/terminal
          - clusters/containerlog
          - clusters/online
          - clusters/offline
          - clusters/tags
          - pipelineruns
          - pipelineruns/stop
          - pipelineruns/log
          - pipelineruns/diffs
          - clusters/dashboards
          - clusters/pods
          - clusters/free
          - clusters/events
          - clusters/outputs
          - clusters/promote
          - clusters/shell
        verbs:
          - '*'
        scopes:
          - '*'
        nonResourceURLs:
          - '*'
      - apiGroups:
          - core
        resources:
          - clusters/templateschematags
        verbs:
          - get
        scopes:
          - '*'
        nonResourceURLs:
          - '*'
  - name: maintainer
    desc: maintainer为组/应用/集群的管理者，拥有除删除资源之外的其他权限，并且也可以进行成员管理
    rules:
      - apiGroups:
          - core
        resources:
          - applications
          - groups/applications
          - applications/members
          - applications/envtemplates
        verbs:
          - create
          - get
          - update
        scopes:
          - '*'
        nonResourceURLs:
          - '*'
      - apiGroups:
          - core
        resources:
          - groups
          - groups/members
          - groups/groups
        verbs:
          - get
          - create
          - update
        scopes:
          - '*'
      - apiGroups:
          - core
        resources:
          - applications/clusters
          - clusters
          - clusters/builddeploy
          - clusters/deploy
          - clusters/diffs
          - clusters/next
          - clusters/restart
          - clusters/rollback
          - clusters/status
          - clusters/members
          - clusters/pipelineruns
          - clusters/terminal
          - clusters/containerlog
          - clusters/online
          - clusters/offline
          - clusters/tags
          - pipelineruns
          - pipelineruns/stop
          - pipelineruns/log
          - pipelineruns/diffs
          - clusters/dashboards
          - clusters/pods
          - clusters/free
          - clusters/events
          - clusters/outputs
          - clusters/promote
          - clusters/shell
        verbs:
          - create
          - get
          - update
        scopes:
          - '*'
        nonResourceURLs:
          - '*'
      - apiGroups:
          - core
        resources:
          - clusters/templateschematags
        verbs:
          - get
        scopes:
          - '*'
        nonResourceURLs:
          - '*'
  - name: pe
    desc: pe为应用/集群的管理者，拥有除删除资源之外的其他权限，并且也可以进行成员管理。破格修改资源上限等
    rules:
      - apiGroups:
          - core
        resources:
          - applications
          - groups/applications
          - applications/members
          - applications/envtemplates
        verbs:
          - create
          - get
          - update
        scopes:
          - '*'
        nonResourceURLs:
          - '*'
      - apiGroups:
          - core
        resources:
          - groups
          - groups/members
          - groups/groups
        verbs:
          - get
          - create
          - update
        scopes:
          - '*'
      - apiGroups:
          - core
        resources:
          - applications/clusters
          - clusters
          - clusters/builddeploy
          - clusters/deploy
          - clusters/diffs
          - clusters/next
          - clusters/restart
          - clusters/rollback
          - clusters/status
          - clusters/members
          - clusters/pipelineruns
          - clusters/terminal
          - clusters/containerlog
          - clusters/online
          - clusters/offline
          - clusters/tags
          - pipelineruns
          - pipelineruns/stop
          - pipelineruns/log
          - pipelineruns/diffs
          - clusters/dashboards
          - clusters/pods
          - clusters/free
          - clusters/templateschematags
          - clusters/events
          - clusters/outputs
          - clusters/promote
          - clusters/shell
        verbs:
          - create
          - get
          - update
        scopes:
          - '*'
        nonResourceURLs:
          - '*'
  - name: guest
    desc: guest为只读人员，拥有组/应用/项目的只读权限，以及测试环境集群创建的权限
    rules:
      - apiGroups:
          - core
        resources:
          - groups
          - groups/members
          - groups/groups
          - applications
          - groups/applications
          - applications/clusters
          - applications/members
          - applications/envtemplates
          - clusters
          - clusters/diffs
          - clusters/status
          - clusters/members
          - clusters/pipelineruns
          - clusters/containerlog
          - clusters/tags
          - pipelineruns
          - pipelineruns/log
          - pipelineruns/diffs
          - clusters/dashboards
          - clusters/pods
          - clusters/events
          - clusters/outputs
          - clusters/templateschematags
        verbs:
          - get
        scopes:
          - '*'
      - apiGroups:
          - core
        resources:
          - applications/clusters
        verbs:
          - create
          - get
          - update
        scopes:
          - test/*
          - reg/*
          - perf/*
          - pre/*`
