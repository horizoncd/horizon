package service

import (
	"context"
	"fmt"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/util/sets"
)

type Service interface {
	// CheckUsersExists check users all exists, if true, return nil
	CheckUsersExists(ctx context.Context, emails []string) error
}

type service struct {
	userManager usermanager.Manager
}

func NewService(manager *managerparam.Manager) Service {
	return &service{
		userManager: manager.UserManager,
	}
}

func (s *service) CheckUsersExists(ctx context.Context, emails []string) error {
	if len(emails) == 0 {
		return nil
	}
	users, err := s.userManager.ListByEmail(ctx, emails)
	if err != nil {
		return err
	}
	userEmailSet := sets.NewString()
	for _, user := range users {
		userEmailSet.Insert(user.Email)
	}
	for _, email := range emails {
		if !userEmailSet.Has(email) {
			return herrors.NewErrNotFound(herrors.UserInDB,
				fmt.Sprintf("user with email %s not exists, please login in horizon first", email))
		}
	}
	return nil
}
