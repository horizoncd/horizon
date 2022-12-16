package validate

import (
	"fmt"
	"regexp"

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
