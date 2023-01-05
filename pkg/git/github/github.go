package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v41/github"
	herrors "github.com/horizoncd/horizon/core/errors"
	gitconfig "github.com/horizoncd/horizon/pkg/config/git"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/git"
	"golang.org/x/oauth2"
)

const Kind = "github"

func init() {
	git.Register(Kind, New)
}

type helper struct {
	client *github.Client
	url    string
}

func New(ctx context.Context, config *gitconfig.Repo) (git.Helper, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &helper{client: client, url: config.URL}, nil
}

func (h helper) GetCommit(ctx context.Context, gitURL string, refType string, ref string) (*git.Commit, error) {
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

func (h helper) ListBranch(ctx context.Context, gitURL string, params *git.SearchParams) ([]string, error) {
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

func (h helper) ListTag(ctx context.Context, gitURL string, params *git.SearchParams) ([]string, error) {
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

func (h helper) GetHTTPLink(gitURL string) (string, error) {
	pid, err := git.ExtractProjectPathFromURL(gitURL)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", h.url, pid), nil
}

func (h helper) GetCommitHistoryLink(gitURL string, commit string) (string, error) {
	httpLink, err := h.GetHTTPLink(gitURL)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/commits/%s", httpLink, commit), nil
}
