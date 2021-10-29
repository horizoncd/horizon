package role

import (
	"context"
	"strings"
	"testing"

	"g.hz.netease.com/horizon/pkg/rbac/types"
	"github.com/stretchr/testify/assert"
)

var (
	ctx = context.Background()
)

var (
	roleForTestOk = `
RolePriorityRankDesc: [owner,maintainer]
DefaultRole: maintainer
Roles:
  - name: owner
    rules:
    - apiGroups:
      - /api/core/v1/*
      resources:
      - groups
      verbs:
      - "*"
      scopes:
      - "*"
      nonResourceURLs:
      - "*"
  - name: maintainer
    rules:
      - apiGroups:
          - /api/core/v1/*
        resources:
          - group
          - group/member
          - group/applications
          - group/groups
          - applications
          - applications/members
          - applications/clusters
          - clusters
        verbs:
          - create
          - get
          - patch
        scopes:
          - "*"
        nonResourceURLs:
          - "*"
      - apiGroups:
          - /api/core/v1/*
        resources:
          - clusters
          - clusters/members
          - clusters/pipelines
          - clusters/builddeploy
          - clusters/deploy
          - clusters/diff
          - clusters/next
          - clusters/restart
          - clusters/rollback
          - clusters/status
        verbs:
          - create
          - get
          - patch
        scopes:
          - "*"
        nonResourceURLs:
          - "*"
`
	roleForTestErr1 = `
RolePriorityRankDesc: [owner]
Roles:
  - name: owner
    rules:
      - apiGroups:
          - /api/core/v1/*
        resources:
          - groups
          - groups/*
          - application
          - application/*
          - clusters
          - clusters/*
        verbs:
          - "*"
        scopes:
          - "*"
        nonResourceURLs:
          - "*"
  - name: maintainer
    rules:
      - apiGroups:
          - /api/core/v1/*
        resources:
          - group
          - group/member
          - group/applications
          - group/groups
          - applications
          - applications/members
          - applications/clusters
          - clusters
        verbs:
          - create
          - get
          - patch
        scopes:
          - "*"
        nonResourceURLs:
          - "*"
      - apiGroups:
          - /api/core/v1/*
        resources:
          - clusters
          - clusters/members
          - clusters/pipelines
          - clusters/builddeploy
          - clusters/deploy
          - clusters/diff
          - clusters/next
          - clusters/restart
          - clusters/rollback
          - clusters/status
        verbs:
          - create
          - get
          - patch
        scopes:
          - "*"
        nonResourceURLs:
          - "*"
`
)

func TestNewFileRole(t *testing.T) {
	tests := []struct {
		name          string
		roleYamlInStr string
		err           error
	}{
		{
			name:          "role number not equal",
			roleYamlInStr: roleForTestErr1,
			err:           ErrorLoadCheckError,
		},
		{
			name:          "role ok",
			roleYamlInStr: roleForTestOk,
			err:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.roleYamlInStr)
			_, err := NewFileRole(ctx, reader)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestRole(t *testing.T) {
	var service Service
	reader := strings.NewReader(roleForTestOk)
	service, err := NewFileRole(ctx, reader)
	assert.Nil(t, err)
	assert.NotNil(t, service)

	tests := []struct {
		name     string
		roleName string
		err      error
	}{
		{
			name:     "getRoleOK",
			roleName: "owner",
			err:      nil,
		}, {
			name:     "getRoleFile",
			roleName: "notExist",
			err:      ErrorRoleNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := service.GetRole(ctx, tt.roleName)
			assert.Equal(t, tt.err, err)
			if err == nil {
				assert.Equal(t, role.Name, tt.roleName)
			}
		})
	}

	policy := types.PolicyRule{
		Verbs:           []string{"*"},
		APIGroups:       []string{"/api/core/v1/*"},
		Resources:       []string{"groups"},
		Scopes:          []string{"*"},
		NonResourceURLs: []string{"*"},
	}
	expectRole := types.Role{
		Name:        "owner",
		PolicyRules: []types.PolicyRule{policy},
	}

	roles, err := service.ListRole(ctx)
	assert.Nil(t, err)
	assert.Equal(t, len(roles), 2)
	assert.Equal(t, roles[0], expectRole)

	assert.NotNil(t, service.GetDefaultRole(ctx))
}
