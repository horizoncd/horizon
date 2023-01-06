package gitlab

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	herrors "github.com/horizoncd/horizon/core/errors"
	gitlablib "github.com/horizoncd/horizon/lib/gitlab"
	gitconfig "github.com/horizoncd/horizon/pkg/config/git"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/git"
	"github.com/xanzy/go-gitlab"
)

const Kind = "gitlab"

func init() {
	git.Register(Kind, New)
}

type Helper struct {
	client gitlablib.Interface
	url    string
}

func New(ctx context.Context, config *gitconfig.Repo) (git.Helper, error) {
	gitlabLib, err := gitlablib.New(config.Token, config.URL, "")
	if err != nil {
		return nil, err
	}

	return &Helper{
		client: gitlabLib,
		url:    config.URL,
	}, nil
}

func (h Helper) GetTagArchive(ctx context.Context, gitURL, tagName string) (*git.Tag, error) {
	URL, err := url.Parse(gitURL)
	if err != nil || gitURL == "" {
		return nil, perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("failed to parse gitlab url: %s", err))
	}

	pidPattern := regexp.MustCompile(`^/(.+?)(?:\.git)?$`)
	matches := pidPattern.FindStringSubmatch(URL.Path)
	if len(matches) != 2 {
		return nil, perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("failed to parse gitlab url: %s", err))
	}

	pid := matches[1]

	archiveData, err := h.client.GetRepositoryArchive(ctx, pid, tagName)
	if err != nil {
		return nil, err
	}

	t, err := h.client.GetTag(ctx, pid, tagName)
	if err != nil {
		return nil, err
	}
	shortID := ""
	if t.Commit != nil {
		shortID = t.Commit.ID
	}

	return &git.Tag{
		ShortID:     shortID,
		Name:        tagName,
		ArchiveData: archiveData,
	}, nil
}

func (h Helper) GetCommit(ctx context.Context, gitURL string, refType string, ref string) (*git.Commit, error) {
	pid, err := git.ExtractProjectPathFromURL(gitURL)
	if err != nil {
		return nil, err
	}
	switch refType {
	case git.GitRefTypeCommit:
		commit, err := h.client.GetCommit(ctx, pid, ref)
		if err != nil {
			return nil, err
		}
		return &git.Commit{
			ID:      commit.ID,
			Message: commit.Message,
		}, nil
	case git.GitRefTypeTag:
		tag, err := h.client.GetTag(ctx, pid, ref)
		if err != nil {
			return nil, err
		}
		return &git.Commit{
			ID:      tag.Commit.ID,
			Message: tag.Commit.Message,
		}, nil
	case git.GitRefTypeBranch:
		branch, err := h.client.GetBranch(ctx, pid, ref)
		if err != nil {
			return nil, err
		}
		return &git.Commit{
			ID:      branch.Commit.ID,
			Message: branch.Commit.Message,
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
	listParam := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    params.PageNumber,
			PerPage: params.PageSize,
		},
		Search: &params.Filter,
	}
	branches, err := h.client.ListBranch(ctx, pid, listParam)
	if err != nil {
		return nil, err
	}
	branchNames := make([]string, 0)
	for _, branch := range branches {
		branchNames = append(branchNames, branch.Name)
	}
	return branchNames, nil
}

func (h Helper) ListTag(ctx context.Context, gitURL string, params *git.SearchParams) ([]string, error) {
	pid, err := git.ExtractProjectPathFromURL(gitURL)
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
	tags, err := h.client.ListTag(ctx, pid, listParam)
	if err != nil {
		return nil, err
	}
	tagNames := make([]string, 0)
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
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

	return fmt.Sprintf("%s/-/commits/%s", httpLink, commit), nil
}
