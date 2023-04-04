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

package service

import (
	"context"
	"fmt"

	herrors "github.com/horizoncd/horizon/core/errors"
	usermanager "github.com/horizoncd/horizon/pkg/manager"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/util/sets"
)

type UserService interface {
	// CheckUsersExists check users all exists, if true, return nil
	CheckUsersExists(ctx context.Context, emails []string) error
}

type userService struct {
	userManager usermanager.UserManager
}

func NewUserService(manager *managerparam.Manager) UserService {
	return &userService{
		userManager: manager.UserManager,
	}
}

func (s *userService) CheckUsersExists(ctx context.Context, emails []string) error {
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
