// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package git

import (
	"context"
	"regexp"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/config/git"
	perror "github.com/horizoncd/horizon/pkg/errors"
)

// Helper interface to do git operations
//
//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/git/git_mock.go -package=mock_git
type Helper interface {
	// GetCommit to get commit of a branch/tag/commitID for a specified git URL
	GetCommit(ctx context.Context, gitURL string, refType string, ref string) (*Commit, error)
	ListBranch(ctx context.Context, gitURL string, params *SearchParams) ([]string, error)
	ListTag(ctx context.Context, gitURL string, params *SearchParams) ([]string, error)
	GetHTTPLink(gitURL string) (string, error)
	GetCommitHistoryLink(gitURL string, commit string) (string, error)
	GetTagArchive(ctx context.Context, gitURL, tagName string) (*Tag, error)
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
