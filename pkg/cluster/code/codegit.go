package code

import (
	"context"
	"regexp"

	herrors "g.hz.netease.com/horizon/core/errors"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	perror "g.hz.netease.com/horizon/pkg/errors"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	"github.com/xanzy/go-gitlab"
)

const (
	_gitlabName = "control"
)

// TODO: git connector (support all kinds of git code repo)

// GitGetter interface to get commit for user code
type GitGetter interface {
	// GetCommit to get commit of a branch/tag/commitID for a specified git URL
	// gitURL is a ssh url, looks like: ssh://git@g.hz.netease.com:22222/music-cloud-native/horizon/horizon.git
	GetCommit(ctx context.Context, gitURL string, refType string, ref string) (*Commit, error)
	ListBranch(ctx context.Context, gitURL string, params *SearchParams) ([]string, error)
	ListTag(ctx context.Context, gitURL string, params *SearchParams) ([]string, error)
}

var _ GitGetter = (*gitGetter)(nil)

type gitGetter struct {
	gitlabLib gitlablib.Interface
}

// NewGitGetter new a GitGetter instance
func NewGitGetter(ctx context.Context, gitlabFactory gitlabfty.Factory) (GitGetter, error) {
	gitlabLib, err := gitlabFactory.GetByName(ctx, _gitlabName)
	if err != nil {
		return nil, err
	}
	return &gitGetter{
		gitlabLib: gitlabLib,
	}, nil
}

func (g *gitGetter) ListBranch(ctx context.Context, gitURL string, params *SearchParams) ([]string, error) {
	pid, err := extractProjectPathFromSSHURL(gitURL)
	if err != nil {
		return nil, err
	}
	listParam := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    params.PageNumber,
			PerPage: params.PageSize,
		},
		Search: &params.Filter,
	}
	branches, err := g.gitlabLib.ListBranch(ctx, pid, listParam)
	if err != nil {
		return nil, err
	}
	branchNames := make([]string, 0)
	for _, branch := range branches {
		branchNames = append(branchNames, branch.Name)
	}
	return branchNames, nil
}

func (g *gitGetter) ListTag(ctx context.Context, gitURL string, params *SearchParams) ([]string, error) {
	pid, err := extractProjectPathFromSSHURL(gitURL)
	if err != nil {
		return nil, err
	}
	listParam := &gitlab.ListTagsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    params.PageNumber,
			PerPage: params.PageSize,
		},
		Search: &params.Filter,
	}
	tags, err := g.gitlabLib.ListTag(ctx, pid, listParam)
	if err != nil {
		return nil, err
	}
	tagNames := make([]string, 0)
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	return tagNames, nil
}

func (g *gitGetter) GetCommit(ctx context.Context, gitURL string, refType string, ref string) (*Commit, error) {
	pid, err := extractProjectPathFromSSHURL(gitURL)
	if err != nil {
		return nil, err
	}

	switch refType {
	case GitRefTypeCommit:
		commit, err := g.gitlabLib.GetCommit(ctx, pid, ref)
		if err != nil {
			return nil, err
		}
		return &Commit{
			ID:      commit.ID,
			Message: commit.Message,
		}, nil
	case GitRefTypeTag:
		tag, err := g.gitlabLib.GetTag(ctx, pid, ref)
		if err != nil {
			return nil, err
		}
		return &Commit{
			ID:      tag.Commit.ID,
			Message: tag.Commit.Message,
		}, nil
	case GitRefTypeBranch:
		branch, err := g.gitlabLib.GetBranch(ctx, pid, ref)
		if err != nil {
			return nil, err
		}
		return &Commit{
			ID:      branch.Commit.ID,
			Message: branch.Commit.Message,
		}, nil
	default:
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "git ref type %s is invalid", refType)
	}
}

// extractProjectPathFromSSHURL extract gitlab project path from ssh url.
// ssh url looks like: ssh://git@g.hz.netease.com:22222/music-cloud-native/horizon/horizon.git
func extractProjectPathFromSSHURL(gitURL string) (string, error) {
	pattern := regexp.MustCompile(`ssh://.+?/(.+).git`)
	matches := pattern.FindStringSubmatch(gitURL)
	if len(matches) != 2 {
		return "", perror.Wrapf(herrors.ErrParamInvalid, "error to extract project path from git ssh url: %v", gitURL)
	}
	return matches[1], nil
}
