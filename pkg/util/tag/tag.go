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

package tag

import (
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/tag/models"
	"github.com/horizoncd/horizon/pkg/util/sets"
	"k8s.io/apimachinery/pkg/labels"
)

func ParseTagSelector(tagSelectorStr string) ([]models.TagSelector, error) {
	// parse tagSelector
	if tagSelectorStr != "" {
		k8sSelector, err := labels.Parse(tagSelectorStr)
		if err != nil {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"failed to parse selector:\n"+
					"tagSelector = %s\nerr = %v", tagSelectorStr, err)
		}
		requirements, selectable := k8sSelector.Requirements()
		if !selectable || len(requirements) == 0 {
			return nil, nil
		}

		var tagSelectors []models.TagSelector
		for _, requirement := range requirements {
			values := sets.NewString(requirement.Values().List()...)
			tagSelectors = append(tagSelectors, models.TagSelector{
				Key:      requirement.Key(),
				Values:   values,
				Operator: string(requirement.Operator()),
			})
		}

		return tagSelectors, nil
	}
	return nil, nil
}
