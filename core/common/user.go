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

package common

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	herror "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/authentication/user"
	perror "github.com/horizoncd/horizon/pkg/errors"
)

const (
	UserQueryName = "filter"
	UserQueryType = "userType"
	UserQueryID   = "id"
)

const (
	contextUserKey = "contextUser"

	AuthorizationHeaderKey = "Authorization"
	TokenHeaderValuePrefix = "Bearer"
)

func UserContextKey() string {
	return contextUserKey
}

func UserFromContext(ctx context.Context) (user.User, error) {
	u, ok := ctx.Value(contextUserKey).(user.User)
	if !ok {
		return nil, herror.ErrFailedToGetUser
	}
	return u, nil
}

func WithContext(parent context.Context, user user.User) context.Context {
	return context.WithValue(parent, UserContextKey(), user) // nolint
}

func SetUser(c *gin.Context, user user.User) {
	// attach user to context
	c.Set(contextUserKey, user)
}

func GetToken(c *gin.Context) (string, error) {
	if _, ok := c.Request.Header[AuthorizationHeaderKey]; !ok {
		return "", perror.Wrap(herror.ErrAuthorizationHeaderNotFound, "")
	}
	token, err := func() (string, error) {
		parts := strings.Split(c.Request.Header.Get(AuthorizationHeaderKey), " ")
		if len(parts) != 2 || parts[0] != TokenHeaderValuePrefix {
			return "", perror.Wrapf(herror.ErrOAuthTokenFormatError, "header = %s",
				c.Request.Header.Get(AuthorizationHeaderKey))
		}
		return parts[1], nil
	}()
	return token, err
}
