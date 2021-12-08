package service

import (
	"context"
	"fmt"

	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/util/sets"
)

var (
	Svc = NewService()
)

type Service interface {
	// CheckUsersExists check users all exists, if true, return nil
	CheckUsersExists(ctx context.Context, emails []string) error
}

type service struct {
	userManager usermanager.Manager
}

func NewService() Service {
	return &service{
		userManager: usermanager.Mgr,
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
			return fmt.Errorf("user with email %s not exists, please login in horizon first", email)
		}
	}
	return nil
}
