package accesstoken

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"g.hz.netease.com/horizon/core/common"
	herror "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	applicationmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	perror "g.hz.netease.com/horizon/pkg/errors"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/oauth/generate"
	oauthmanager "g.hz.netease.com/horizon/pkg/oauth/manager"
	oauthmodels "g.hz.netease.com/horizon/pkg/oauth/models"
	"g.hz.netease.com/horizon/pkg/oauth/store"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	roleservice "g.hz.netease.com/horizon/pkg/rbac/role"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	callbacks "g.hz.netease.com/horizon/pkg/util/ormcallbacks"
)

var (
	ctx context.Context
	c   Controller
)

// valid params
var (
	commonName      = "test"
	commonRole      = "owner"
	commonScopes    = []string{"applications:read-write"}
	commonExpiresAt = time.Now().Add(time.Hour * 24).Format(ExpiresAtFormat)
	commonResource  = Resource{
		ResourceType: "groups",
	}
)

// invalid params
const (
	expiredDate  = "2021-10-1"
	invalidDate  = "20222-11-2"
	invalidRole  = "joker"
	invalidScope = "test"
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
	oauthMgr := oauthmanager.NewManager(oauthAppStore, tokenStore, generate.NewAuthorizeGenerate(), authorizeCodeExpireIn, accessTokenExpireIn)

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
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: user.Name,
		ID:   user.ID,
	})

	group, err := manager.GroupManager.Create(ctx, &groupmodels.Group{
		Name: "test",
	})
	if err != nil {
		panic(err)
	}
	commonResource.ResourceID = group.ID

	c = NewController(param)

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	type SingleCase struct {
		req              CreateAccessTokenRequest
		expectedCauseErr error
		expectedResp     *CreateAccessTokenResponse
	}

	testCases := []SingleCase{
		{
			req: CreateAccessTokenRequest{
				Name:      commonName,
				Role:      commonRole,
				Scopes:    commonScopes,
				ExpiresAt: expiredDate,
			},
			expectedCauseErr: herror.ErrParamInvalid,
		},
		{
			req: CreateAccessTokenRequest{
				Name:      commonName,
				Role:      commonRole,
				Scopes:    commonScopes,
				ExpiresAt: invalidDate,
			},
			expectedCauseErr: herror.ErrParamInvalid,
		},
		{
			req: CreateAccessTokenRequest{
				Name:      commonName,
				Role:      commonRole,
				Scopes:    commonScopes,
				ExpiresAt: expiredDate,
			},
			expectedCauseErr: herror.ErrParamInvalid,
		},
		{
			req: CreateAccessTokenRequest{
				Name:      commonName,
				Role:      commonRole,
				Scopes:    commonScopes,
				ExpiresAt: NeverExpire,
			},
		},
		{
			req: CreateAccessTokenRequest{
				Name:      commonName,
				Role:      commonRole,
				Scopes:    commonScopes,
				ExpiresAt: commonExpiresAt,
				Resource:  &commonResource,
			},
		},
	}

	for _, testCase := range testCases {
		var (
			createTokenResp *CreateAccessTokenResponse
			err             error
		)

		createTokenResp, err = c.CreateAccessToken(ctx, testCase.req)
		assert.Equal(t, testCase.expectedCauseErr, perror.Cause(err))
		if testCase.expectedResp != nil {
			assert.Equal(t, testCase.expectedResp, createTokenResp)
		}

		if testCase.expectedCauseErr == nil {
			_, total, err := c.ListTokens(ctx, testCase.req.Resource, nil)
			assert.Equal(t, nil, err)
			assert.Equal(t, 1, total)

			c.RevokeAccessToken(ctx, createTokenResp.ID)
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
