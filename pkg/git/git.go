package git

import (
	"context"
	"regexp"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/config/git"
	perror "github.com/horizoncd/horizon/pkg/errors"
)

// Helper interface to do git operations
type Helper interface {
	// GetCommit to get commit of a branch/tag/commitID for a specified git URL
	GetCommit(ctx context.Context, gitURL string, refType string, ref string) (*Commit, error)
	ListBranch(ctx context.Context, gitURL string, params *SearchParams) ([]string, error)
	ListTag(ctx context.Context, gitURL string, params *SearchParams) ([]string, error)
	GetHTTPLink(gitURL string) (string, error)
	GetCommitHistoryLink(gitURL string, commit string) (string, error)
}

type Constructor func(ctx context.Context, config *git.Repo) (Helper, error)

var factory = make(map[string]Constructor)

func Register(kind string, constructor Constructor) {
	factory[kind] = constructor
}

func NewHelper(ctx context.Context, config *git.Repo) (Helper, error) {
	for kind, constructor := range factory {
		if kind == config.Kind {
			return constructor(ctx, config)
		}
	}
	return nil, perror.Wrapf(herrors.ErrParamInvalid, "Repo initializes failed, kind = %v is not implement", config.Kind)
}

// ExtractProjectPathFromURL extract git project path from gitURL.
func ExtractProjectPathFromURL(gitURL string) (string, error) {
	pattern := regexp.MustCompile(`^(?:http(?:s?)|ssh)://.+?/(.+?)(?:.git)?$`)
	matches := pattern.FindStringSubmatch(gitURL)
	if len(matches) != 2 {
		return "", perror.Wrapf(herrors.ErrParamInvalid, "error to extract project path from git ssh url: %v", gitURL)
	}
	return matches[1], nil
}
