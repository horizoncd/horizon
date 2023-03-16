package permission

import (
	"context"

	"github.com/horizoncd/horizon/pkg/core/common"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
)

func OnlySelfAndAdmin(ctx context.Context, self uint) error {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	if !currentUser.IsAdmin() && currentUser.GetID() != self {
		return perror.Wrap(herrors.ErrForbidden, "you can only access resources of yourself")
	}
	return nil
}
