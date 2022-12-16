package service

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
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
	tokenSvc = NewService(manager)
	os.Exit(m.Run())
}

func TestService(t *testing.T) {
	// Create User AccessToken
	name := "token"
	expiresAtStr := time.Now().Add(time.Hour * 72).Format(ExpiresAtFormat)
	scopes := make([]string, 2)
	scopes = append(scopes, "clusters:read-write")
	scopes = append(scopes, "applications:read-only")
	token, err := tokenSvc.CreateUserAccessToken(ctx, name, expiresAtStr, aUser.GetID(), scopes)
	assert.Nil(t, err)
	tokenInDB, err := tokenManager.LoadTokenByID(ctx, token.ID)
	assert.Nil(t, err)
	assert.Equal(t, name, tokenInDB.Name)
	assert.Equal(t, strings.Join(scopes, " "), tokenInDB.Scope)

	// Create Internal AccessToken
	expiresIn := time.Hour * 2
	token, err = tokenSvc.CreateInternalAccessToken(ctx, name, expiresIn, aUser.GetID(), scopes)
	assert.Nil(t, err)
	log.Infof(ctx, "%+v", token)
	tokenInDB, err = tokenManager.LoadTokenByID(ctx, token.ID)
	assert.Nil(t, err)
	assert.Equal(t, name, tokenInDB.Name)
	assert.Equal(t, expiresIn, tokenInDB.ExpiresIn)
}
