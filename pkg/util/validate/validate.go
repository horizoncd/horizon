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

package validate

import (
	"fmt"
	"regexp"

	"github.com/google/go-containerregistry/pkg/name"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
)

func CheckURL(u string) error {
	re := `^http(s)?://.+$`
	pattern := regexp.MustCompile(re)
	if !pattern.MatchString(u) {
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("invalid url, should satisfies the pattern %v", re))
	}
	return nil
}

func CheckGitURL(gitURL string) error {
	re := `^(?:git|ssh|https?|git@[-\w.]+):(//)?(.*?)(\.git)(/?|#[-\d\w._]+?)$`
	pattern := regexp.MustCompile(re)
	if !pattern.MatchString(gitURL) {
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("invalid git url, should satisfies the pattern %v", re))
	}
	return nil
}

// CheckImageURL validate OCI container image url
func CheckImageURL(imageURL string) error {
	_, err := name.ParseReference(imageURL)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("invalid image url, error: %v", err.Error()))
	}
	return nil
}
