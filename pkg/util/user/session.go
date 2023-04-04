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

package user

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	perror "github.com/horizoncd/horizon/pkg/errors"
	usermodels "github.com/horizoncd/horizon/pkg/models"
)

func SetSession(ss *sessions.Session,
	request *http.Request, response http.ResponseWriter,
	user *usermodels.User) error {
	ss.Values[common.SessionKeyAuthUser] = &userauth.DefaultInfo{
		Name:     user.Name,
		FullName: user.FullName,
		ID:       user.ID,
		Email:    user.Email,
		Admin:    user.Admin,
	}

	if err := ss.Save(request, response); err != nil {
		return perror.Wrapf(herrors.ErrSessionSaveFailed,
			"err = %v", err)
	}
	return nil
}

func GetSession(store sessions.Store, request *http.Request) (*sessions.Session, error) {
	session, err := store.Get(request, common.CookieKeyAuth)
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrSessionNotFound,
			"session name = %s\n err = %v", common.CookieKeyAuth, err)
	}
	return session, nil
}
