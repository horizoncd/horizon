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
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	tokenconfig "github.com/horizoncd/horizon/pkg/config/token"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	tokenmanager "github.com/horizoncd/horizon/pkg/token/manager"
	tokenmodels "github.com/horizoncd/horizon/pkg/token/models"
	"github.com/horizoncd/horizon/pkg/util/log"
	callbacks "github.com/horizoncd/horizon/pkg/util/ormcallbacks"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db           *gorm.DB
	tokenManager tokenmanager.Manager
	tokenSvc     Service
	aUser        userauth.User = &userauth.DefaultInfo{
		Name:     "alias",
		FullName: "alias",
		ID:       32,
		Email:    "",
		Admin:    false,
	}
	ctx = context.WithValue(context.Background(), common.UserContextKey(), aUser) // nolint
)

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&tokenmodels.Token{}); err != nil {
		panic(err)
	}
	db = db.WithContext(context.WithValue(context.Background(), common.UserContextKey(), aUser)) // nolint
	callbacks.RegisterCustomCallbacks(db)

	manager := managerparam.InitManager(db)
	tokenManager = manager.TokenManager
	tokenSvc = NewService(manager, tokenconfig.Config{
		JwtSigningKey:         "UZMccEsEgXA/phl3w/OK1gZU6lhKJIswZqsyfQEPqpc=",
		CallbackTokenExpireIn: 2 * time.Hour,
	})

	os.Exit(m.Run())
}

func TestService(t *testing.T) {
	// Create User AccessToken
	name := "token"
	expiresAtStr := time.Now().Add(time.Hour * 72).Format(ExpiresAtFormat)
	scopes := make([]string, 2)
	scopes = append(scopes, "clusters:read-write")
	scopes = append(scopes, "applications:read-only")
	token, err := tokenSvc.CreateAccessToken(ctx, name, expiresAtStr, aUser.GetID(), scopes)
	assert.Nil(t, err)
	tokenInDB, err := tokenManager.LoadTokenByID(ctx, token.ID)
	assert.Nil(t, err)
	assert.Equal(t, name, tokenInDB.Name)
	assert.Equal(t, strings.Join(scopes, " "), tokenInDB.Scope)

	// Create JWT token
	jwtToken, err := tokenSvc.CreateJWTToken(strconv.Itoa(int(aUser.GetID())), 2*time.Hour,
		WithPipelinerunID(12))
	assert.Nil(t, err)
	log.Infof(ctx, "%s", jwtToken)
	// Parse JWT token
	claims, err := tokenSvc.ParseJWTToken(jwtToken)
	assert.Nil(t, err)
	log.Infof(ctx, "%+v", claims)
	log.Infof(ctx, "%v", *claims.PipelinerunID)
	userID, err := strconv.ParseUint(claims.Subject, 10, 64)
	assert.Nil(t, err)
	assert.Equal(t, aUser.GetID(), uint(userID))
}
