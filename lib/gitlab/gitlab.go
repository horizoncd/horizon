package gitlab

import (
	"context"
	"crypto/tls"
	"net/http"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"

	"github.com/xanzy/go-gitlab"
)

// Interface to interact with gitlab
type Interface interface {
	// GetGroup gets a group's detail with the given gid.
	// The gid can be the group's ID or relative path such as first/second/third.
	// See https://docs.gitlab.com/ee/api/groups.html#details-of-a-group for more information.
	GetGroup(ctx context.Context, gid interface{}) (*gitlab.Group, error)

	// ListGroupProjects list a group's project
	ListGroupProjects(ctx context.Context, gid interface{}, page, perPage int) ([]*gitlab.Project, error)

	// CreateGroup create a gitlab group with the given name and path.
	// The parentID is alternative, if you specify the parentID, it will
	// create a subgroup of this parent.
	// See https://docs.gitlab.com/ee/api/groups.html#new-group for more information.
	CreateGroup(ctx context.Context, name, path string, parentID *int) (*gitlab.Group, error)

	// DeleteGroup delete a gitlab group with the given gid.
	// The gid can be the group's ID or relative path such as first/second/third.
	// See https://docs.gitlab.com/ee/api/groups.html#remove-group for more information.
	DeleteGroup(ctx context.Context, gid interface{}) error

	// GetProject get a project with the specified pid.
	// The pid can be the project's ID or relative path such as fist/second.
	// See https://docs.gitlab.com/ee/api/projects.html#get-single-project for more information.
	GetProject(ctx context.Context, pid interface{}) (*gitlab.Project, error)

	// CreateProject create a project under the specified group.
	// See https://docs.gitlab.com/ee/api/projects.html#create-project for more information.
	CreateProject(ctx context.Context, name string, groupID int) (*gitlab.Project, error)

	// DeleteProject delete a project with the given pid.
	// The pid can be the project's ID or relative path such as fist/second.
	// See https://docs.gitlab.com/ee/api/projects.html#delete-project for more information.
	DeleteProject(ctx context.Context, pid interface{}) error

	// GetCommit get a specified commit
	// The pid can be the project's ID or relative path such as fist/second.
	// See https://docs.gitlab.com/ee/api/commits.html#get-a-single-commit for more information.
	GetCommit(ctx context.Context, pid interface{}, commit string) (_ *gitlab.Commit, err error)

	// GetBranch get branch of the specified project.
	// The pid can be the project's ID or relative path such as fist/second.
	// See https://docs.gitlab.com/ee/api/branches.html#get-single-repository-branch for more information.
	GetBranch(ctx context.Context, pid interface{}, branch string) (*gitlab.Branch, error)

	// CreateBranch create a branch from fromRef for the specified project.
	// The pid can be the project's ID or relative path such as fist/second.
	// The fromRef can be the name of branch, tag or commit.
	// See https://docs.gitlab.com/ee/api/branches.html#create-repository-branch for more information.
	CreateBranch(ctx context.Context, pid interface{}, branch, fromRef string) (*gitlab.Branch, error)

	// DeleteBranch delete a branch for the specified project.
	// The pid can be the project's ID or relative path such as fist/second.
	// See https://docs.gitlab.com/ee/api/branches.html#delete-repository-branch for more information.
	DeleteBranch(ctx context.Context, pid interface{}, branch string) error

	// ListBranch list the branch, Get a list of repository branches from a project, sorted by name alphabetically.
	// see https://docs.gitlab.com/ee/api/branches.html#list-repository-branches
	ListBranch(ctx context.Context, pid interface{},
		listBranchOptions *gitlab.ListBranchesOptions) (_ []*gitlab.Branch, err error)

	// CreateMR create a merge request from source to target with the specified title in project.
	// The pid can be the project's ID or relative path such as fist/second.
	// See https://docs.gitlab.com/ee/api/merge_requests.html#create-mr for more information.
	CreateMR(ctx context.Context, pid interface{}, source, target, title string) (*gitlab.MergeRequest, error)

	// ListMRs list merge requests for specified project.
	// The pid should be the project's ID.
	// See https://docs.gitlab.com/ee/api/merge_requests.html#list-project-merge-requests for more information.
	ListMRs(ctx context.Context, pid interface{}, source, target, state string) ([]*gitlab.MergeRequest, error)

	// AcceptMR merge a merge request for specified project.
	// The pid can be the project's ID or relative path such as fist/second.
	// See https://docs.gitlab.com/ee/api/merge_requests.html#accept-mr for more information.
	AcceptMR(ctx context.Context, pid interface{}, mrID int,
		mergeCommitMsg *string, shouldRemoveSourceBranch *bool) (*gitlab.MergeRequest, error)

	// WriteFiles write including create, delete, update multiple files within a specified project.
	// The pid can be the project's ID or relative path such as fist/second.
	// See https://docs.gitlab.com/ee/api/commits.html#create-a-commit-with-multiple-files-and-actions
	// for more information.
	WriteFiles(ctx context.Context, pid interface{}, branch, commitMsg string,
		startBranch *string, actions []CommitAction) (*gitlab.Commit, error)

	// GetFile get a file content for specified filepath in the specified project with the ref.
	// The pid can be the project's ID or relative path such as fist/second.
	// The ref can be the name of branch, tag or commit.
	// See https://docs.gitlab.com/ee/api/repository_files.html#get-file-from-repository for more information.
	GetFile(ctx context.Context, pid interface{}, ref, filepath string) ([]byte, error)

	// TransferProject transfer a project with the specified pid to the new group with the gid.
	// The pid can be the project's ID or relative path such as fist/second.
	// The gid can be the group's ID or relative path such as first/third.
	TransferProject(ctx context.Context, pid interface{}, gid interface{}) error

	// EditNameAndPathForProject update name and path for a specified project.
	// The pid can be the project's ID or relative path such as fist/second.
	EditNameAndPathForProject(ctx context.Context, pid interface{}, newName, newPath *string) error

	// Compare branches, tags or commits.
	// The pid can be the project's ID or relative path such as fist/second.
	// See https://docs.gitlab.com/ee/api/repositories.html#compare-branches-tags-or-commits for more information.
	Compare(ctx context.Context, pid interface{}, from, to string, straight *bool) (*gitlab.Compare, error)

	GetSSHURL(ctx context.Context) string
}

var _ Interface = (*helper)(nil)

type FileAction string

// The available file actions.
const (
	FileCreate FileAction = "create"
	FileUpdate FileAction = "update"
)

// CommitAction represents a single file action within a commit.
type CommitAction struct {
	Action   FileAction
	FilePath string
	Content  string
}

func (a FileAction) toFileActionValuePtr() *gitlab.FileActionValue {
	s := gitlab.FileActionValue(a)
	return &s
}

type helper struct {
	client  *gitlab.Client
	httpURL string
	sshURL  string
}

// New an instance of Gitlab
func New(token, httpURL, sshURL string) (Interface, error) {
	client, err := gitlab.NewClient(token,
		gitlab.WithBaseURL(httpURL),
		gitlab.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}))
	if err != nil {
		return nil, herrors.NewErrCreateFailed(herrors.GitlabResource, err.Error())
	}
	return &helper{
		client:  client,
		httpURL: httpURL,
		sshURL:  sshURL,
	}, nil
}

func (h *helper) GetGroup(ctx context.Context, gid interface{}) (_ *gitlab.Group, err error) {
	const op = "gitlab: get group"
	defer wlog.Start(ctx, op).StopPrint()

	group, rsp, err := h.client.Groups.GetGroup(gid, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, parseError(rsp, err)
	}

	return group, nil
}

func (h *helper) ListGroupProjects(ctx context.Context, gid interface{},
	page, perPage int) (_ []*gitlab.Project, err error) {
	const op = "gitlab: list group projects"
	defer wlog.Start(ctx, op).StopPrint()

	if page < 1 {
		return nil, perror.Wrap(herrors.ErrParamInvalid, "page cannot be less 1")
	}
	if perPage < 1 {
		return nil, perror.Wrap(herrors.ErrParamInvalid, "perPage cannot be less 1")
	}

	projects, rsp, err := h.client.Groups.ListGroupProjects(gid, &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, parseError(rsp, err)
	}

	return projects, nil
}

func (h *helper) CreateGroup(ctx context.Context, name, path string, parentID *int) (_ *gitlab.Group, err error) {
	const op = "gitlab: create group"
	defer wlog.Start(ctx, op).StopPrint()

	group, rsp, err := h.client.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Name:     &name,
		Path:     &path,
		ParentID: parentID,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(rsp, err)
	}

	return group, nil
}

func (h *helper) DeleteGroup(ctx context.Context, gid interface{}) (err error) {
	const op = "gitlab: delete group"
	defer wlog.Start(ctx, op).StopPrint()

	rsp, err := h.client.Groups.DeleteGroup(gid, gitlab.WithContext(ctx))

	return parseError(rsp, err)
}

func (h *helper) GetProject(ctx context.Context, pid interface{}) (_ *gitlab.Project, err error) {
	const op = "gitlab: get project"
	defer wlog.Start(ctx, op).StopPrint()

	project, rsp, err := h.client.Projects.GetProject(pid, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, parseError(rsp, err)
	}
	return project, nil
}

func (h *helper) CreateProject(ctx context.Context, name string, groupID int) (_ *gitlab.Project, err error) {
	const op = "gitlab: create project"
	defer wlog.Start(ctx, op).StopPrint()

	project, rsp, err := h.client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		InitializeWithReadme: func() *bool { b := true; return &b }(),
		Name:                 &name,
		Path:                 &name,
		NamespaceID:          &groupID,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(rsp, err)
	}

	return project, err
}

func (h *helper) DeleteProject(ctx context.Context, pid interface{}) (err error) {
	const op = "gitlab: delete project"
	defer wlog.Start(ctx, op).StopPrint()

	rsp, err := h.client.Projects.DeleteProject(pid, gitlab.WithContext(ctx))

	return parseError(rsp, err)
}

func (h *helper) GetBranch(ctx context.Context, pid interface{}, branch string) (_ *gitlab.Branch, err error) {
	const op = "gitlab: get branch"
	defer wlog.Start(ctx, op).StopPrint()

	b, rsp, err := h.client.Branches.GetBranch(pid, branch, gitlab.WithContext(ctx))
	if err != nil {
		return nil, parseError(rsp, err)
	}

	return b, nil
}

func (h *helper) ListBranch(ctx context.Context, pid interface{},
	listBranchOptions *gitlab.ListBranchesOptions) (_ []*gitlab.Branch, err error) {
	const op = "gitlab: list branch"
	defer wlog.Start(ctx, op).StopPrint()

	branches, rsp, err := h.client.Branches.ListBranches(pid, listBranchOptions, nil)
	if err != nil {
		return nil, parseError(rsp, err)
	}
	return branches, nil
}

func (h *helper) GetCommit(ctx context.Context, pid interface{}, commit string) (_ *gitlab.Commit, err error) {
	const op = "gitlab: get commit"
	defer wlog.Start(ctx, op).StopPrint()

	c, rsp, err := h.client.Commits.GetCommit(pid, commit, gitlab.WithContext(ctx))
	if err != nil {
		return nil, parseError(rsp, err)
	}

	return c, nil
}

func (h *helper) CreateBranch(ctx context.Context, pid interface{},
	branch, fromRef string) (_ *gitlab.Branch, err error) {
	const op = "gitlab: create branch"
	defer wlog.Start(ctx, op).StopPrint()

	b, rsp, err := h.client.Branches.CreateBranch(pid, &gitlab.CreateBranchOptions{
		Branch: &branch,
		Ref:    &fromRef,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(rsp, err)
	}

	return b, nil
}

func (h *helper) DeleteBranch(ctx context.Context, pid interface{}, branch string) (err error) {
	const op = "gitlab: delete branch"
	defer wlog.Start(ctx, op).StopPrint()

	rsp, err := h.client.Branches.DeleteBranch(pid, branch, gitlab.WithContext(ctx))

	return parseError(rsp, err)
}

func (h *helper) CreateMR(ctx context.Context, pid interface{},
	source, target, title string) (_ *gitlab.MergeRequest, err error) {
	const op = "gitlab: create mr"
	defer wlog.Start(ctx, op).StopPrint()

	mr, rsp, err := h.client.MergeRequests.CreateMergeRequest(pid, &gitlab.CreateMergeRequestOptions{
		Title:        &title,
		SourceBranch: &source,
		TargetBranch: &target,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(rsp, err)
	}

	return mr, nil
}

func (h *helper) ListMRs(ctx context.Context, pid interface{},
	source, target, state string) (_ []*gitlab.MergeRequest, err error) {
	mrs, rsp, err := h.client.MergeRequests.ListProjectMergeRequests(pid, &gitlab.ListProjectMergeRequestsOptions{
		SourceBranch: &source,
		TargetBranch: &target,
		State:        &state,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, perror.WithMessagef(parseError(rsp, err),
			"failed to list merge requests for project: %v", pid)
	}

	return mrs, nil
}

func (h *helper) AcceptMR(ctx context.Context, pid interface{}, mrID int,
	mergeCommitMsg *string, shouldRemoveSourceBranch *bool) (_ *gitlab.MergeRequest, err error) {
	const op = "gitlab: accept mr"
	defer wlog.Start(ctx, op).StopPrint()

	mr, rsp, err := h.client.MergeRequests.AcceptMergeRequest(pid, mrID, &gitlab.AcceptMergeRequestOptions{
		MergeCommitMessage:       mergeCommitMsg,
		ShouldRemoveSourceBranch: shouldRemoveSourceBranch,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(rsp, err)
	}

	return mr, nil
}

func (h *helper) WriteFiles(ctx context.Context, pid interface{}, branch, commitMsg string,
	startBranch *string, actions []CommitAction) (_ *gitlab.Commit, err error) {
	const op = "gitlab: write files"
	defer wlog.Start(ctx, op).StopPrint()

	commit, rsp, err := h.client.Commits.CreateCommit(pid, &gitlab.CreateCommitOptions{
		Branch:        &branch,
		CommitMessage: &commitMsg,
		StartBranch:   startBranch,
		Actions: func() []*gitlab.CommitActionOptions {
			acts := make([]*gitlab.CommitActionOptions, 0)
			for i := range actions {
				acts = append(acts, &gitlab.CommitActionOptions{
					Action:   actions[i].Action.toFileActionValuePtr(),
					FilePath: &actions[i].FilePath,
					Content:  &actions[i].Content,
				})
			}
			return acts
		}(),
	}, gitlab.WithContext(ctx))

	if err != nil {
		log.Errorf(ctx, "err: %v", err)
		return nil, parseError(rsp, err)
	}

	return commit, nil
}

func (h *helper) GetFile(ctx context.Context, pid interface{}, ref, filepath string) (_ []byte, err error) {
	const op = "gitlab: get file"
	defer wlog.Start(ctx, op).StopPrint()

	content, rsp, err := h.client.RepositoryFiles.GetRawFile(pid, filepath, &gitlab.GetRawFileOptions{
		Ref: &ref,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(rsp, err)
	}

	return content, nil
}

func (h *helper) TransferProject(ctx context.Context, pid interface{}, gid interface{}) (err error) {
	const op = "gitlab: transfer project"
	defer wlog.Start(ctx, op).StopPrint()

	if _, rsp, err := h.client.Projects.TransferProject(pid, &gitlab.TransferProjectOptions{
		Namespace: gid,
	}, gitlab.WithContext(ctx)); err != nil {
		return parseError(rsp, err)
	}

	return nil
}

func (h *helper) EditNameAndPathForProject(ctx context.Context, pid interface{}, newName, newPath *string) (err error) {
	const op = "gitlab: edit name and path for project"
	defer wlog.Start(ctx, op).StopPrint()

	if _, rsp, err := h.client.Projects.EditProject(pid, &gitlab.EditProjectOptions{
		Name: newName,
		Path: newPath,
	}, gitlab.WithContext(ctx)); err != nil {
		return parseError(rsp, err)
	}

	return nil
}

func (h *helper) Compare(ctx context.Context, pid interface{}, from, to string,
	straight *bool) (_ *gitlab.Compare, err error) {
	const op = "gitlab: compare branchs"
	defer wlog.Start(ctx, op).StopPrint()

	compare, rsp, err := h.client.Repositories.Compare(pid, &gitlab.CompareOptions{
		From:     &from,
		To:       &to,
		Straight: straight,
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, parseError(rsp, err)
	}

	return compare, nil
}

func (h *helper) GetSSHURL(ctx context.Context) string {
	return h.sshURL
}

func parseError(resp *gitlab.Response, err error) error {
	if err == nil {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return herrors.NewErrNotFound(herrors.GitlabResource, err.Error())
	}

	return perror.Wrap(herrors.ErrGitlabInternal, err.Error())
}
