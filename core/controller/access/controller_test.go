package access

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	applicationmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/pkg/member"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/rbac"
	roleservice "g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	"github.com/stretchr/testify/assert"
)

var (
	ctx         context.Context
	c           Controller
	group       *groupmodels.Group
	application *applicationmodels.Application
	cluster     *clustermodels.Cluster
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	if err := db.AutoMigrate(&membermodels.Member{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&groupmodels.Group{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&applicationmodels.Application{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&clustermodels.Cluster{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&usermodels.User{}); err != nil {
		panic(err)
	}

	roleService, err := roleservice.NewFileRole(context.Background(), strings.NewReader(roleConfig))
	if err != nil {
		panic(err)
	}
	memberService := memberservice.NewService(roleService)
	if err != nil {
		panic(err)
	}
	ctx = orm.NewContext(ctx, db)
	ctx = context.WithValue(ctx, common.Key(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   uint(110),
	})

	rbacAuthorizer := rbac.NewAuthorizer(roleService, memberService)
	skippers := middleware.MethodAndPathSkipper("*",
		regexp.MustCompile("(^/apis/front/.*)|(^/health)|(^/metrics)|(^/apis/login)|"+
			"(^/apis/core/v1/roles)|(^/apis/internal/.*)"))
	c = NewController(rbacAuthorizer, skippers)

	group, err = groupmanager.Mgr.Create(ctx, &groupmodels.Group{
		Name:            "group",
		Path:            "/group",
		VisibilityLevel: "private",
	})
	if err != nil {
		panic(err)
	}
	application, _ = applicationmanager.Mgr.Create(ctx, &applicationmodels.Application{
		Name:    "application",
		GroupID: group.ID,
	}, nil)

	cluster, _ = clustermanager.Mgr.Create(ctx, &clustermodels.Cluster{
		Name:          "cluster",
		ApplicationID: application.ID,
	}, nil, nil)

	os.Exit(m.Run())
}

// nolint
func TestController_GetAccesses_Guest(t *testing.T) {

	guest, err := usermanager.Mgr.Create(ctx, &usermodels.User{
		Name: "guest",
	})

	nonMemberCtx := context.WithValue(ctx, common.Key(), &userauth.DefaultInfo{
		ID: 2,
	})
	guestCtx := context.WithValue(ctx, common.Key(), &userauth.DefaultInfo{
		ID: guest.ID,
	})

	deniedAPIs := []API{
		{
			URL:    fmt.Sprintf("/apis/core/v1/clusters/%d/shell", cluster.ID),
			Method: "GET",
		},
		{
			URL:    fmt.Sprintf("/apis/core/v1/applications/%d/clusters?scope=dev/hz", application.ID),
			Method: "POST",
		},
	}

	allowAPIs := []API{
		{
			URL:    fmt.Sprintf("/apis/core/v1/clusters/%d/status", cluster.ID),
			Method: "GET",
		},
		{
			URL:    fmt.Sprintf("/apis/core/v1/applications/%d/clusters?scope=test/hz", application.ID),
			Method: "POST",
		},
	}

	apis := append(deniedAPIs, allowAPIs...)

	reviewResults, err := c.Review(nonMemberCtx, apis)
	assert.Nil(t, err)
	for _, api := range deniedAPIs {
		assert.Equal(t, false, reviewResults[api.URL][api.Method].Allowed)
	}
	for _, api := range allowAPIs {
		assert.Equal(t, true, reviewResults[api.URL][api.Method].Allowed)
	}

	reviewResults, err = c.Review(guestCtx, apis)
	assert.Nil(t, err)
	for _, api := range deniedAPIs {
		assert.Equal(t, false, reviewResults[api.URL][api.Method].Allowed)
	}
	for _, api := range allowAPIs {
		assert.Equal(t, true, reviewResults[api.URL][api.Method].Allowed)
	}
}

// nolint
func TestController_GetAccesses_Owner(t *testing.T) {
	owner, err := usermanager.Mgr.Create(ctx, &usermodels.User{
		Name: "owner",
	})

	ctx := context.WithValue(ctx, common.Key(), &userauth.DefaultInfo{
		ID: owner.ID,
	})

	_, err = member.Mgr.Create(ctx, &membermodels.Member{
		ResourceType: "groups",
		ResourceID:   group.ID,
		Role:         "owner",
		MemberType:   membermodels.MemberUser,
		MemberNameID: owner.ID,
	})
	assert.Nil(t, err)

	deniedAPIs := []API{
		{
			URL:    fmt.Sprintf("/apis/core/v1/clusters/%d/templateschematags", cluster.ID),
			Method: "POST",
		},
	}

	allowAPIs := []API{
		{
			URL:    fmt.Sprintf("/apis/core/v1/clusters/%d/shell", cluster.ID),
			Method: "GET",
		},
		{
			URL:    fmt.Sprintf("/apis/core/v1/clusters/%d", cluster.ID),
			Method: "DELETE",
		},
	}
	apis := append(deniedAPIs, allowAPIs...)

	reviewResults, err := c.Review(ctx, apis)
	assert.Nil(t, err)
	for _, api := range deniedAPIs {
		fmt.Println(reviewResults[api.URL][api.Method])
		assert.Equal(t, false, reviewResults[api.URL][api.Method].Allowed)
	}
	for _, api := range allowAPIs {
		fmt.Println(reviewResults[api.URL][api.Method])
		assert.Equal(t, true, reviewResults[api.URL][api.Method].Allowed)
	}
}

// nolint
func TestController_GetAccesses_Admin(t *testing.T) {
	admin, err := usermanager.Mgr.Create(ctx, &usermodels.User{
		Name: "admin",
	})

	ctx := context.WithValue(ctx, common.Key(), &userauth.DefaultInfo{
		ID:    admin.ID,
		Admin: true,
	})

	apis := []API{
		{
			URL:    fmt.Sprintf("/apis/core/v1/clusters/%d/shell", cluster.ID),
			Method: "GET",
		},
		{
			URL:    fmt.Sprintf("/apis/core/v1/clusters/%d", cluster.ID),
			Method: "DELETE",
		},
		{
			URL:    "/apis/core/v1/groups",
			Method: "POST",
		},
	}

	reviewResults, err := c.Review(ctx, apis)
	assert.Nil(t, err)
	for _, api := range apis {
		fmt.Println(reviewResults[api.URL][api.Method])
		assert.Equal(t, true, reviewResults[api.URL][api.Method].Allowed)
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
