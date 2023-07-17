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
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	"github.com/horizoncd/horizon/pkg/util/sets"
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
		userManager: manager.UserMgr,
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
