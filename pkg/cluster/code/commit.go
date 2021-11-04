package code

import (
	"context"
	"fmt"
	"regexp"

	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
)

const (
	_gitlabName = "control"
)

type Commit struct {
	ID      string
	Message string
}

// CommitGetter interface to get commit for user code
type CommitGetter interface {
	// GetCommit to get commit for specified gitURL and branch
	// gitURL is a ssh url, looks like: ssh://git@g.hz.netease.com:22222/music-cloud-native/horizon/horizon.git
	GetCommit(ctx context.Context, gitURL, branch string) (*Commit, error)
}

var _ CommitGetter = (*commitGetter)(nil)

type commitGetter struct {
	gitlabLib gitlablib.Interface
}

// NewCommitGetter new a CommitGetter instance
func NewCommitGetter(ctx context.Context, gitlabFactory gitlabfty.Factory) (CommitGetter, error) {
	gitlabLib, err := gitlabFactory.GetByName(ctx, _gitlabName)
	if err != nil {
		return nil, err
	}
	return &commitGetter{
		gitlabLib: gitlabLib,
	}, nil
}

func (g *commitGetter) GetCommit(ctx context.Context, gitURL, branch string) (*Commit, error) {
	pid, err := extractProjectPathFromSSHURL(gitURL)
	if err != nil {
		return nil, err
	}
	gitlabBranch, err := g.gitlabLib.GetBranch(ctx, pid, branch)
	if err != nil {
		return nil, err
	}
	return &Commit{
		ID:      gitlabBranch.Commit.ID,
		Message: gitlabBranch.Commit.Message,
	}, nil
}

// extractProjectPathFromSSHURL extract gitlab project path from ssh url.
// ssh url looks like: ssh://git@g.hz.netease.com:22222/music-cloud-native/horizon/horizon.git
func extractProjectPathFromSSHURL(gitURL string) (string, error) {
	pattern := regexp.MustCompile(`ssh://.+?/(.+).git`)
	matches := pattern.FindStringSubmatch(gitURL)
	if len(matches) != 2 {
		return "", fmt.Errorf("error to extract project path from git ssh url: %v", gitURL)
	}
	return matches[1], nil
}
