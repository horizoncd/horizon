package permission

import (
	"context"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
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
