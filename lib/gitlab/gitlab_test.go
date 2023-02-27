package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
)

/*
NOTE: gitlab params must set by environment variable.
env name is GITLAB_PARAMS_FOR_TEST, and the value is a json string, look like:
{
	"token": "xxx",
	"baseURL": "http://cicd.mockserver.org",
	"rootGroupName": "xxx",
	"rootGroupID": xxx
}

1. token is used for auth, see https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html for more information.
2. baseURL is the basic URL for gitlab.
3. rootGroupName is a root group, our unit tests will do some operations under this group.
4. rootGroupID is the ID for this root group.


You can run this unit test just like this:

export GITLAB_PARAMS_FOR_TEST="$(cat <<\EOF
{
	"token": "xxx",
	"baseURL": "http://cicd.mockserver.org",
	"rootGroupName": "xxx",
	"rootGroupID": xxx
}
EOF
)"
go test -v ./lib/gitlab/

NOTE: when there is no GITLAB_PARAMS_FOR_TEST environment variable, skip this test.
*/

var (
	ctx = context.Background()
	g   Interface

	rootGroupName string
	rootGroupID   int
)

func intToPtr(i int) *int {
	return &i
}

type Param struct {
	Token         string `json:"token"`
	BaseURL       string `json:"baseURL"`
	RootGroupName string `json:"rootGroupName"`
	RootGroupID   int    `json:"rootGroupId"`
}

func TestMain(m *testing.M) {
	var err error

	param := os.Getenv("GITLAB_PARAMS_FOR_TEST")
	if param == "" {
		return
	}

	var p *Param
	if err := json.Unmarshal([]byte(param), &p); err != nil {
		panic(err)
	}

	g, err = New(p.Token, p.BaseURL)
	if err != nil {
		panic(err)
	}

	rootGroupName = p.RootGroupName
	rootGroupID = p.RootGroupID
	if rootGroupName != "" && rootGroupID == 0 {
		return
	}

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	groupName := "horizon-unittest-group"
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
	assert.Equal(t, &herrors.HorizonErrNotFound{Source: herrors.GitlabResource}, perror.Cause(err))

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

	// 7.1 list branch
	branches, err := g.ListBranch(ctx, pid, nil)
	assert.Nil(t, err)
	assert.Equal(t, len(branches), 2)

	var filter = "mas"
	branches, err = g.ListBranch(ctx, pid, &gitlab.ListBranchesOptions{
		Search: &filter,
	})
	assert.Nil(t, err)
	assert.Equal(t, len(branches), 1)
	assert.Equal(t, branches[0].Name, "master")

	// 8. get a branch
	branch, err = g.GetBranch(ctx, pid, newBranch)
	assert.Nil(t, err)
	assert.Equal(t, branch.Name, newBranch)

	branchCommit, err := g.GetCommit(ctx, pid, branch.Commit.ID)
	assert.Nil(t, err)
	assert.Equal(t, branchCommit.ID, branch.Commit.ID)
	assert.Equal(t, branchCommit.Message, branch.Commit.Message)

	// 9. delete a branch
	err = g.DeleteBranch(ctx, pid, newBranch)
	assert.Nil(t, err)

	// 10. get this branch again, will return 404 error
	_, err = g.GetBranch(ctx, pid, newBranch)
	assert.NotNil(t, err)
	assert.Equal(t, &herrors.HorizonErrNotFound{Source: herrors.GitlabResource}, perror.Cause(err))

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
		{
			Action:   FileCreate,
			FilePath: "old_name.yaml",
			Content:  "this is a yaml file for testing rename and delete file",
		},
	}
	commit, err := g.WriteFiles(ctx, pid, newBranch, "commit to create file",
		&startBranch, actions)
	assert.Nil(t, err)
	b, _ = json.Marshal(commit)
	t.Logf(string(b))

	commit, err = g.WriteFiles(ctx, pid, newBranch, "commit to rename file",
		&startBranch, []CommitAction{
			{
				Action:       FileMove,
				FilePath:     "new_dir/new_name.yaml",
				PreviousPath: "old_name.yaml",
			},
		},
	)
	assert.Nil(t, err)
	b, _ = json.Marshal(commit)
	t.Logf(string(b))

	commit, err = g.WriteFiles(ctx, pid, newBranch, "commit to delete file",
		&startBranch, []CommitAction{
			{
				Action:   FileDelete,
				FilePath: "new_dir/new_name.yaml",
			},
		},
	)
	assert.Nil(t, err)
	b, _ = json.Marshal(commit)
	t.Logf(string(b))

	compare, err := g.Compare(ctx, pid, startBranch, newBranch, nil)
	assert.Nil(t, err)
	for _, diff := range compare.Diffs {
		t.Logf(diff.Diff)
	}

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

	// 14. close a mr
	mr2, err := g.CloseMR(ctx, pid, mr.IID)
	assert.Nil(t, err)
	assert.Equal(t, mr.IID, mr2.IID)

	// 15. create a mr
	mr, err = g.CreateMR(ctx, pid, newBranch, startBranch, "this is title")
	assert.Nil(t, err)
	t.Logf(mr.WebURL)
	t.Logf("mr.ID: %v", mr.ID)
	t.Logf("mr.IID: %v", mr.IID)

	// 16. accept a mr
	_, err = g.AcceptMR(ctx, pid, mr.IID, nil, nil)
	assert.Nil(t, err)

	// 17. transfer a project
	newGroupName := "newGroup"
	_, err = g.CreateGroup(ctx, newGroupName, newGroupName, intToPtr(rootGroupID))
	assert.Nil(t, err)

	newGroupPath := fmt.Sprintf("%v/%v", rootGroupName, newGroupName)
	err = g.TransferProject(ctx, pid, newGroupPath)
	assert.Nil(t, err)

	defer func() {
		_ = g.DeleteGroup(ctx, newGroupPath)
	}()

	// 18. edit a project's name and path
	pid = fmt.Sprintf("%v/%v", newGroupPath, projectName)
	newProjectName := fmt.Sprintf("%v-%d", projectName, 1)
	newProjectPath := newProjectName
	err = g.EditNameAndPathForProject(ctx, pid, &newProjectName, &newProjectPath)
	assert.Nil(t, err)

	// 19. delete project
	pid = fmt.Sprintf("%v/%v", newGroupPath, newProjectPath)
	err = g.DeleteProject(ctx, pid)
	assert.Nil(t, err)
}
