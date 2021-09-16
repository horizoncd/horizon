package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/util/errors"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
)

const (
	// baseURL & token is the parameters for a specified gitlab used for unit test
	token   = ""
	baseURL = "http://cicd.mockserver.org/"

	// rootGroupName & rootGroupID is the root group. Our unit tests will do some operations under this group.
	// http://cicd.mockserver.org/subgroup-for-unit-test
	rootGroupName = "subgroup-for-unit-test"
	rootGroupID   = 3512
)

var (
	ctx = context.Background()
	g   Gitlab
)

func intToPtr(i int) *int {
	return &i
}

func TestMain(m *testing.M) {
	var err error
	g, err = New(token, baseURL)
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	groupName := "group"
	groupPath := fmt.Sprintf("%v/%v", rootGroupName, groupName)

	projectName := "project"

	defer func() {
		_ = g.DeleteGroup(ctx, groupPath)
	}()

	var (
		err     error
		group   *gitlab.Group
		project *gitlab.Project
	)

	// 1. get group first. will return 404 error
	_, err = g.GetGroup(ctx, groupPath)
	assert.NotNil(t, err)
	assert.Equal(t, errors.ENotFound, errors.ErrorCode(err))

	// 2. create this group
	group, err = g.CreateGroup(ctx, groupName, groupName, intToPtr(rootGroupID))
	assert.Nil(t, err)
	b, err := json.Marshal(group)
	assert.Nil(t, err)
	t.Log(string(b))
	assert.Equal(t, groupName, group.Name)
	assert.Equal(t, groupPath, group.FullPath)

	// 3. get this group, will be ok
	group, err = g.GetGroup(ctx, groupPath)
	assert.Nil(t, err)
	assert.Equal(t, groupName, group.Name)
	assert.Equal(t, groupPath, group.FullPath)

	// 4. get project, will return 404 first
	pid := fmt.Sprintf("%v/%v", groupPath, projectName)
	project, err = g.GetProject(ctx, pid)
	assert.NotNil(t, err)
	assert.Nil(t, project)

	// 5. create a project
	project, err = g.CreateProject(ctx, projectName, group.ID)
	assert.Nil(t, err)
	assert.Equal(t, projectName, project.Name)

	// 6. get project again
	project, err = g.GetProject(ctx, pid)
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// 7. create a branch
	newBranch := "gitops"
	startBranch := "master"
	branch, err := g.CreateBranch(ctx, pid, newBranch, startBranch)
	assert.Nil(t, err)
	assert.Equal(t, branch.Name, newBranch)

	// 8. get a branch
	branch, err = g.GetBranch(ctx, pid, newBranch)
	assert.Nil(t, err)
	assert.Equal(t, branch.Name, newBranch)

	// 9. delete a branch
	err = g.DeleteBranch(ctx, pid, newBranch)
	assert.Nil(t, err)

	// 10. get this branch again, will return 404 error
	_, err = g.GetBranch(ctx, pid, newBranch)
	assert.NotNil(t, err)
	assert.Equal(t, errors.ENotFound, errors.ErrorCode(err))

	// 11. write files to new branch
	projectBytes, err := json.MarshalIndent(project, "", "    ")
	assert.Nil(t, err)
	actions := []CommitAction{
		{
			Action:   FileCreate,
			FilePath: "a/b.json",
			Content:  string(projectBytes),
		},
		{
			Action:   FileCreate,
			FilePath: "c",
			Content:  "this is content for c",
		},
	}
	commit, err := g.WriteFiles(ctx, pid, newBranch, "commit", &startBranch, actions)
	assert.Nil(t, err)
	b, _ = json.Marshal(commit)
	t.Logf(string(b))

	// 12. get file
	content, err := g.GetFile(ctx, pid, newBranch, "c")
	assert.Nil(t, err)
	assert.Equal(t, "this is content for c", string(content))

	// 13. create a mr
	mr, err := g.CreateMR(ctx, pid, newBranch, startBranch, "this is title")
	assert.Nil(t, err)
	t.Logf(mr.WebURL)
	t.Logf("mr.ID: %v", mr.ID)
	t.Logf("mr.IID: %v", mr.IID)

	// 14. accept a mr
	_, err = g.AcceptMR(ctx, pid, mr.IID, nil, nil)
	assert.Nil(t, err)

	// 15. delete project
	err = g.DeleteProject(ctx, pid)
	assert.Nil(t, err)
}
