package gitlab

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"

	"g.hz.netease.com/horizon/util/errors"
	"g.hz.netease.com/horizon/util/log"
	"g.hz.netease.com/horizon/util/wlog"
	"github.com/xanzy/go-gitlab"
)

// Gitlab interface to interact with gitlab
type Gitlab interface {
	// GetGroup gets a group's detail with the given gid.
	// The gid can be the group's ID or relative path such as first/second/third.
	// See https://docs.gitlab.com/ee/api/groups.html#details-of-a-group for more information.
	GetGroup(ctx context.Context, gid interface{}) (*gitlab.Group, error)

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

	// CreateMR create a merge request from source to target with the specified title in project.
	// The pid can be the project's ID or relative path such as fist/second.
	// See https://docs.gitlab.com/ee/api/merge_requests.html#create-mr for more information.
	CreateMR(ctx context.Context, pid interface{}, source, target, title string) (*gitlab.MergeRequest, error)

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
}

var _ Gitlab = (*helper)(nil)

type FileAction string

// The available file actions.
const (
	FileCreate FileAction = "create"
	FileDelete FileAction = "delete"
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
	client *gitlab.Client
}

// New new a instance of Gitlab
func New(token, baseURL string) (Gitlab, error) {
	client, err := gitlab.NewClient(token,
		gitlab.WithBaseURL(baseURL),
		gitlab.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}))
	if err != nil {
		return nil, err
	}
	return &helper{
		client: client,
	}, nil
}

func (h *helper) GetGroup(ctx context.Context, gid interface{}) (_ *gitlab.Group, err error) {
	const op = "gitlab: get group"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	group, resp, err := h.client.Groups.GetGroup(gid, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, parseError(op, resp, err)
	}

	return group, nil
}

func (h *helper) CreateGroup(ctx context.Context, name, path string, parentID *int) (_ *gitlab.Group, err error) {
	const op = "gitlab: create group"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	group, resp, err := h.client.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Name:     &name,
		Path:     &path,
		ParentID: parentID,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(op, resp, err)
	}

	return group, nil
}

func (h *helper) DeleteGroup(ctx context.Context, gid interface{}) (err error) {
	const op = "gitlab: delete group"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	resp, err := h.client.Groups.DeleteGroup(gid, gitlab.WithContext(ctx))

	return parseError(op, resp, err)
}

func (h *helper) GetProject(ctx context.Context, pid interface{}) (_ *gitlab.Project, err error) {
	const op = "gitlab: get project"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	project, resp, err := h.client.Projects.GetProject(pid, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, parseError(op, resp, err)
	}
	return project, nil
}

func (h *helper) CreateProject(ctx context.Context, name string, groupID int) (_ *gitlab.Project, err error) {
	const op = "gitlab: create project"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	project, resp, err := h.client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		InitializeWithReadme: func() *bool { b := true; return &b }(),
		Name:                 &name,
		NamespaceID:          &groupID,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(op, resp, err)
	}

	return project, err
}

func (h *helper) DeleteProject(ctx context.Context, pid interface{}) (err error) {
	const op = "gitlab: delete project"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	resp, err := h.client.Projects.DeleteProject(pid, gitlab.WithContext(ctx))

	return parseError(op, resp, err)
}

func (h *helper) GetBranch(ctx context.Context, pid interface{}, branch string) (_ *gitlab.Branch, err error) {
	const op = "gitlab: get branch"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	b, resp, err := h.client.Branches.GetBranch(pid, branch, gitlab.WithContext(ctx))
	if err != nil {
		return nil, parseError(op, resp, err)
	}

	return b, nil
}

func (h *helper) CreateBranch(ctx context.Context, pid interface{},
	branch, fromRef string) (_ *gitlab.Branch, err error) {
	const op = "gitlab: create branch"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	b, resp, err := h.client.Branches.CreateBranch(pid, &gitlab.CreateBranchOptions{
		Branch: &branch,
		Ref:    &fromRef,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(op, resp, err)
	}

	return b, nil
}

func (h *helper) DeleteBranch(ctx context.Context, pid interface{}, branch string) (err error) {
	const op = "gitlab: delete branch"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	resp, err := h.client.Branches.DeleteBranch(pid, branch, gitlab.WithContext(ctx))

	return parseError(op, resp, err)
}

func (h *helper) CreateMR(ctx context.Context, pid interface{},
	source, target, title string) (_ *gitlab.MergeRequest, err error) {
	const op = "gitlab: create mr"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	mr, resp, err := h.client.MergeRequests.CreateMergeRequest(pid, &gitlab.CreateMergeRequestOptions{
		Title:        &title,
		SourceBranch: &source,
		TargetBranch: &target,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(op, resp, err)
	}

	return mr, nil
}

func (h *helper) AcceptMR(ctx context.Context, pid interface{}, mrID int,
	mergeCommitMsg *string, shouldRemoveSourceBranch *bool) (_ *gitlab.MergeRequest, err error) {
	const op = "gitlab: accept mr"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	mr, resp, err := h.client.MergeRequests.AcceptMergeRequest(pid, mrID, &gitlab.AcceptMergeRequestOptions{
		MergeCommitMessage:       mergeCommitMsg,
		ShouldRemoveSourceBranch: shouldRemoveSourceBranch,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(op, resp, err)
	}

	return mr, nil
}

func (h *helper) WriteFiles(ctx context.Context, pid interface{}, branch, commitMsg string,
	startBranch *string, actions []CommitAction) (_ *gitlab.Commit, err error) {
	const op = "gitlab: write files"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	commit, resp, err := h.client.Commits.CreateCommit(pid, &gitlab.CreateCommitOptions{
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
		return nil, parseError(op, resp, err)
	}

	return commit, nil
}

func (h *helper) GetFile(ctx context.Context, pid interface{}, ref, filepath string) (_ []byte, err error) {
	const op = "gitlab: get file"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	content, resp, err := h.client.RepositoryFiles.GetRawFile(pid, filepath, &gitlab.GetRawFileOptions{
		Ref: &ref,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return nil, parseError(op, resp, err)
	}

	return content, nil
}

func parseError(op errors.Op, resp *gitlab.Response, err error) error {
	if err == nil {
		return nil
	}
	if resp == nil {
		return errors.E(op, err)
	}
	b, _ := json.Marshal(resp)
	log.Errorf(context.TODO(), "resp: %v", string(b))
	return errors.E(op, resp.StatusCode, err)
}
