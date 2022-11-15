package tag

import (
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/tag/models"
	"g.hz.netease.com/horizon/pkg/util/sets"
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
