package manager

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/core/common"
	"github.com/horizoncd/horizon/pkg/token/generator"
	tokenmodels "github.com/horizoncd/horizon/pkg/token/models"
	callbacks "github.com/horizoncd/horizon/pkg/util/ormcallbacks"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/rand"
)

var (
	db                       *gorm.DB
	tokenManager             Manager
	userAccessTokenGenerator generator.AccessTokenCodeGenerator
	aUser                    userauth.User = &userauth.DefaultInfo{
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

	tokenManager = New(db)
	userAccessTokenGenerator = generator.NewGeneralAccessTokenGenerator()
	os.Exit(m.Run())
}

func TestTokenBasic(t *testing.T) {
	// create
	code := userAccessTokenGenerator.GenCode(&generator.CodeGenerateInfo{
		Token: tokenmodels.Token{UserID: aUser.GetID()},
	})
	token := &tokenmodels.Token{
		Name:      "tokenName",
		Code:      code,
		Scope:     "clusters:read-write",
		CreatedAt: time.Now(),
		ExpiresIn: time.Hour * 24,
		UserID:    aUser.GetID(),
	}
	tokenInDB, err := tokenManager.CreateToken(ctx, token)
	assert.Nil(t, err)

	// load
	tokenInDB, err = tokenManager.LoadTokenByID(ctx, tokenInDB.ID)
	assert.Nil(t, err)
	assert.Equal(t, token.Code, tokenInDB.Code)
	tokenInDB, err = tokenManager.LoadTokenByCode(ctx, tokenInDB.Code)
	assert.Nil(t, err)
	assert.Equal(t, token.Name, tokenInDB.Name)

	// revoke
	err = tokenManager.RevokeTokenByID(ctx, tokenInDB.ID)
	assert.Nil(t, err)
	_, err = tokenManager.LoadTokenByID(ctx, tokenInDB.ID)
	assert.NotNil(t, err)

	tokenWithClientID := &tokenmodels.Token{
		Name:      "tokenName",
		Code:      code,
		ClientID:  rand.String(10),
		Scope:     "clusters:read-write",
		CreatedAt: time.Now(),
		ExpiresIn: time.Hour * 24,
		UserID:    aUser.GetID(),
	}
	tokenWithClientIDInDB, err := tokenManager.CreateToken(ctx, tokenWithClientID)
	assert.Nil(t, err)
	err = tokenManager.RevokeTokenByClientID(ctx, tokenWithClientIDInDB.ClientID)
	assert.Nil(t, err)
	_, err = tokenManager.LoadTokenByID(ctx, tokenWithClientIDInDB.ID)
	assert.NotNil(t, err)
}
