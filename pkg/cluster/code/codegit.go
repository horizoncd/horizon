package code

import (
	"context"
	"regexp"

	herrors "github.com/horizoncd/horizon/core/errors"
	gitconfig "github.com/horizoncd/horizon/pkg/config/git"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/git"
	"github.com/horizoncd/horizon/pkg/git/github"
)

//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/cluster/code/mock_codegit.go -package=mock_code
type GitGetter interface {
	// GetCommit to get commit of a branch/tag/commitID for a specified git URL
	GetCommit(ctx context.Context, gitURL string, refType string, ref string) (*git.Commit, error)
	ListBranch(ctx context.Context, gitURL string, params *git.SearchParams) ([]string, error)
	ListTag(ctx context.Context, gitURL string, params *git.SearchParams) ([]string, error)
	GetHTTPLink(gitURL string) (string, error)
	GetCommitHistoryLink(gitURL string, commit string) (string, error)
	GetTagArchive(ctx context.Context, gitURL, tagName string) (*git.Tag, error)
}

var _ GitGetter = (*gitGetter)(nil)

var regexHost = regexp.MustCompile(`^(?:https?:)?(?://)?(?:[^@\n]+@)?([^:/\n]+)`)

const (
	githubHost = "github.com"
	githubURL  = "https://github.com"
)

type gitGetter struct {
	gitMap map[string]git.Helper
}

// NewGitGetter new a GitGetter instance
func NewGitGetter(ctx context.Context, repos []*gitconfig.Repo) (GitGetter, error) {
	gitMap := map[string]git.Helper{}
	// Check if GitHub has been configured, and add it if not.
	githubConfigured := false
	for _, repo := range repos {
		host, err := extractHostFromURL(repo.URL)
		if err != nil {
			return nil, err
		}
		helper, err := git.NewHelper(ctx, repo)
		if err != nil {
			return nil, err
		}
		gitMap[host] = helper
		if host == githubHost {
			githubConfigured = true
		}
	}
	if githubConfigured == false {
		helper, _ := git.NewHelper(ctx, &gitconfig.Repo{
			Kind: github.Kind,
			URL:  githubURL,
		})
		gitMap[githubHost] = helper
	}
	return &gitGetter{
		gitMap: gitMap,
	}, nil
}

func (g *gitGetter) ListBranch(ctx context.Context, gitURL string, params *git.SearchParams) ([]string, error) {
	helper, err := g.getGitHelper(gitURL)
	if err != nil {
		return nil, err
	}

	return helper.ListBranch(ctx, gitURL, params)
}

func (g *gitGetter) ListTag(ctx context.Context, gitURL string, params *git.SearchParams) ([]string, error) {
	helper, err := g.getGitHelper(gitURL)
	if err != nil {
		return nil, err
	}

	return helper.ListTag(ctx, gitURL, params)
}

func (g *gitGetter) GetCommit(ctx context.Context, gitURL string, refType string, ref string) (*git.Commit, error) {
	helper, err := g.getGitHelper(gitURL)
	if err != nil {
		return nil, err
	}

	return helper.GetCommit(ctx, gitURL, refType, ref)
}

func (g *gitGetter) GetHTTPLink(gitURL string) (string, error) {
	if gitURL == "" {
		return "", nil
	}
	helper, err := g.getGitHelper(gitURL)
	if err != nil {
		return "", err
	}
	return helper.GetHTTPLink(gitURL)
}

func (g *gitGetter) GetCommitHistoryLink(gitURL string, commit string) (string, error) {
	if gitURL == "" || commit == "" {
		return "", nil
	}
	helper, err := g.getGitHelper(gitURL)
	if err != nil {
		return "", err
	}
	return helper.GetCommitHistoryLink(gitURL, commit)
}

func (g *gitGetter) GetTagArchive(ctx context.Context, gitURL, tagName string) (*git.Tag, error) {
	helper, err := g.getGitHelper(gitURL)
	if err != nil {
		return nil, err
	}
	return helper.GetTagArchive(ctx, gitURL, tagName)
}

func (g *gitGetter) getGitHelper(gitURL string) (git.Helper, error) {
	host, err := extractHostFromURL(gitURL)
	if err != nil {
		return nil, err
	}

	var ok bool
	var h git.Helper
	if h, ok = g.gitMap[host]; !ok {
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "no git repo corresponding to url: %v", gitURL)
	}
	return h, nil
}

func extractHostFromURL(gitURL string) (string, error) {
	matches := regexHost.FindStringSubmatch(gitURL)
	if len(matches) != 2 {
		return "", perror.Wrapf(herrors.ErrParamInvalid, "error to extract host from url: %v", gitURL)
	}
	return matches[1], nil
}
