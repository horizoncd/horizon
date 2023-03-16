package github

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/go-github/v41/github"
	gitconfig "github.com/horizoncd/horizon/pkg/config/git"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/git"
	"golang.org/x/oauth2"
)

const Kind = "github"

func init() {
	git.Register(Kind, New)
}

type Helper struct {
	client *github.Client
	url    string
}

func New(ctx context.Context, config *gitconfig.Repo) (git.Helper, error) {
	if config.Token == "" {
		return &Helper{client: github.NewClient(nil), url: config.URL}, nil
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Helper{client: client, url: config.URL}, nil
}

func (h Helper) GetTagArchive(ctx context.Context, gitURL, tagName string) (*git.Tag, error) {
	pattern := regexp.MustCompile(`^(?:https://|git@)github.com[:/]([^/]+)/([^/]+?)(?:\.git)?$`)
	matches := pattern.FindAllStringSubmatch(gitURL, -1)
	if len(matches) == 0 || len(matches[0]) < 3 {
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "git url is incorrect: git url = %s", gitURL)
	}
	owner, repo := matches[0][1], matches[0][2]
	ref, respRef, err := h.client.Git.GetRef(ctx, owner, repo, fmt.Sprintf("tags/%s", tagName))
	if err != nil {
		return nil, perror.Wrapf(herrors.NewErrGetFailed(herrors.GithubResource,
			" failed to get ref"), "failed to get tag from github: err = %v", err)
	}
	defer respRef.Body.Close()
	gitObject := ref.GetObject()

	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", owner, repo, tagName), nil)
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrHTTPRequestFailed, "failed to create request: err = %v", err)
	}
	buf := bytes.Buffer{}
	_, err = h.client.Do(ctx, req, &buf)
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrHTTPRequestFailed, "failed to get tag: err = %v", err)
	}
	archiveData := buf.Bytes()
	sha := ""
	if gitObject.SHA != nil {
		sha = *gitObject.SHA
	}
	return &git.Tag{
		ShortID:     sha,
		Name:        tagName,
		ArchiveData: archiveData,
	}, nil
}

func (h Helper) GetCommit(ctx context.Context, gitURL string, refType string, ref string) (*git.Commit, error) {
	pid, err := git.ExtractProjectPathFromURL(gitURL)
	if err != nil {
		return nil, err
	}
	paths := strings.Split(pid, "/")
	switch refType {
	case git.GitRefTypeCommit:
		commit, _, err := h.client.Repositories.GetCommit(ctx, paths[0], paths[1], ref, &github.ListOptions{})
		if err != nil {
			return nil, err
		}
		return &git.Commit{
			ID:      commit.GetSHA(),
			Message: commit.Commit.GetMessage(),
		}, nil
	case git.GitRefTypeTag:
		// todo handle the situation that more than 100 tags exist
		tags, _, err := h.client.Repositories.ListTags(ctx, paths[0], paths[1], &github.ListOptions{
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			return nil, perror.Wrapf(herrors.ErrParamInvalid, "git ref type %s is invalid", refType)
		}
		var sha string
		for _, tag := range tags {
			if tag.GetName() == ref {
				sha = tag.Commit.GetSHA()
				break
			}
		}
		if sha == "" {
			return nil, perror.Wrapf(herrors.ErrParamInvalid, "git ref type %s is invalid", refType)
		}
		commit, _, err := h.client.Repositories.GetCommit(ctx, paths[0], paths[1], sha, &github.ListOptions{})
		if err != nil {
			return nil, err
		}
		return &git.Commit{
			ID:      commit.GetSHA(),
			Message: commit.Commit.GetMessage(),
		}, nil
	case git.GitRefTypeBranch:
		branch, _, err := h.client.Repositories.GetBranch(ctx, paths[0], paths[1], ref, true)
		if err != nil {
			return nil, err
		}
		return &git.Commit{
			ID:      branch.Commit.GetSHA(),
			Message: branch.Commit.Commit.GetMessage(),
		}, nil
	default:
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "git ref type %s is invalid", refType)
	}
}

func (h Helper) ListBranch(ctx context.Context, gitURL string, params *git.SearchParams) ([]string, error) {
	pid, err := git.ExtractProjectPathFromURL(gitURL)
	if err != nil {
		return nil, err
	}
	// GitHub Repo HttpURL's format is: https://github/${owner}/${repo}.git
	paths := strings.Split(pid, "/")
	branches, _, err := h.client.Repositories.ListBranches(ctx, paths[0], paths[1], &github.BranchListOptions{
		ListOptions: github.ListOptions{
			Page:    params.PageNumber,
			PerPage: params.PageSize,
		},
	})
	if err != nil {
		return nil, err
	}
	branchNames := make([]string, 0)
	for _, branch := range branches {
		branchNames = append(branchNames, branch.GetName())
	}

	return branchNames, nil
}

func (h Helper) ListTag(ctx context.Context, gitURL string, params *git.SearchParams) ([]string, error) {
	pid, err := git.ExtractProjectPathFromURL(gitURL)
	if err != nil {
		return nil, err
	}
	paths := strings.Split(pid, "/")
	tags, _, err := h.client.Repositories.ListTags(ctx, paths[0], paths[1], &github.ListOptions{
		Page:    params.PageNumber,
		PerPage: params.PageSize,
	})
	if err != nil {
		return nil, err
	}
	tagNames := make([]string, 0)
	for _, tag := range tags {
		tagNames = append(tagNames, tag.GetName())
	}

	return tagNames, nil
}

func (h Helper) GetHTTPLink(gitURL string) (string, error) {
	pid, err := git.ExtractProjectPathFromURL(gitURL)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", h.url, pid), nil
}

func (h Helper) GetCommitHistoryLink(gitURL string, commit string) (string, error) {
	httpLink, err := h.GetHTTPLink(gitURL)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/commits/%s", httpLink, commit), nil
}
