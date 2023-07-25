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

package manager

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	perror "github.com/horizoncd/horizon/pkg/errors"
	oauthdao "github.com/horizoncd/horizon/pkg/oauth/dao"
	"github.com/horizoncd/horizon/pkg/oauth/models"
	"github.com/horizoncd/horizon/pkg/token/generator"
	tokenmanager "github.com/horizoncd/horizon/pkg/token/manager"
	tokenmodels "github.com/horizoncd/horizon/pkg/token/models"
	tokenstore "github.com/horizoncd/horizon/pkg/token/store"
	callbacks "github.com/horizoncd/horizon/pkg/util/ormcallbacks"
)

var (
	db           *gorm.DB
	tokenStore   tokenstore.Store
	oauthAppDAO  oauthdao.DAO
	oauthManager Manager
	tokenManager tokenmanager.Manager
	aUser        userauth.User = &userauth.DefaultInfo{
		Name:     "alias",
		FullName: "alias",
		ID:       32,
		Email:    "",
		Admin:    false,
	}
	ctx = context.WithValue(context.Background(), common.UserContextKey(), aUser)

	authorizeCodeExpireIn = time.Second * 3
	accessTokenExpireIn   = time.Second * 3
	refreshTokenExpireIn  = time.Second * 3
)

func checkOAuthApp(req *CreateOAuthAppReq, app *models.OauthApp) bool {
	if (req.Name == app.Name) && (req.RedirectURI == app.RedirectURL) &&
		(req.HomeURL == app.HomeURL) && (req.Desc == app.Desc) &&
		(req.OwnerType == app.OwnerType) && (req.OwnerID == app.OwnerID) &&
		req.APPType == app.AppType {
		return true
	}
	return false
}

func TestOauthAppBasic(t *testing.T) {
	createReq := &CreateOAuthAppReq{
		Name:        "OauthTest",
		RedirectURI: "https://example.com/oauth/redirect",
		HomeURL:     "https://example.com",
		Desc:        "This is an example  oauth app",
		OwnerType:   models.GroupOwnerType,
		OwnerID:     1,
		APPType:     models.HorizonOAuthAPP,
	}
	oauthApp, err := oauthManager.CreateOauthApp(ctx, createReq)
	assert.Nil(t, err)
	assert.NotNil(t, oauthApp)
	assert.True(t, checkOAuthApp(createReq, oauthApp))

	oauthRetApp, err := oauthManager.GetOAuthApp(ctx, oauthApp.ClientID)
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(oauthRetApp, oauthApp))

	updateReq := UpdateOauthAppReq{
		Name:        "OauthTest2",
		HomeURL:     "https://example2.com",
		RedirectURI: "https://example.com/oauth/redirect2",
		Desc:        "This is",
	}
	updateRet, err := oauthManager.UpdateOauthApp(ctx, oauthApp.ClientID, updateReq)
	assert.Nil(t, err)
	assert.Equal(t, oauthApp.ID, updateRet.ID)
	assert.Equal(t, oauthApp.ClientID, updateRet.ClientID)
	assert.Equal(t, updateReq.Name, updateRet.Name)
	assert.Equal(t, updateReq.HomeURL, updateRet.HomeURL)
	assert.Equal(t, updateReq.RedirectURI, updateRet.RedirectURL)

	apps, err := oauthManager.ListOauthApp(ctx, models.GroupOwnerType, createReq.OwnerID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))

	err = oauthManager.DeleteOAuthApp(ctx, oauthApp.ClientID)
	assert.Nil(t, err)

	_, err = oauthManager.GetOAuthApp(ctx, oauthApp.ClientID)
	assert.NotNil(t, err)
	if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
		assert.Fail(t, "error is not found")
	}
}

func TestClientSecretBasic(t *testing.T) {
	// create app
	createReq := &CreateOAuthAppReq{
		Name:        "OauthTest",
		RedirectURI: "https://example.com/oauth/redirect",
		HomeURL:     "https://example.com",
		Desc:        "This is an example  oauth app",
		OwnerType:   models.GroupOwnerType,
		OwnerID:     1,
		APPType:     models.HorizonOAuthAPP,
	}
	oauthApp, err := oauthManager.CreateOauthApp(ctx, createReq)
	assert.Nil(t, err)
	assert.NotNil(t, oauthApp)
	assert.True(t, checkOAuthApp(createReq, oauthApp))

	defer func() {
		err = oauthManager.DeleteOAuthApp(ctx, oauthApp.ClientID)
		assert.Nil(t, err)
	}()

	// create secret
	secret1, err := oauthManager.CreateSecret(ctx, oauthApp.ClientID)
	assert.Nil(t, err)
	assert.NotNil(t, secret1)
	assert.Equal(t, secret1.ClientID, oauthApp.ClientID)

	secret2, err := oauthManager.CreateSecret(ctx, oauthApp.ClientID)
	assert.Nil(t, err)
	assert.NotNil(t, secret2)
	assert.Equal(t, secret2.ClientID, oauthApp.ClientID)

	secrets, err := oauthManager.ListSecret(ctx, oauthApp.ClientID)
	assert.Nil(t, err)
	assert.Equal(t, len(secrets), 2)
	for _, secret := range secrets {
		if secret.CreatedBy == aUser.GetID() &&
			(secret.ClientID == secret1.ClientID || secret.ClientID == secret2.ClientID) &&
			(checkClientSecret(secret.ClientSecret, secret1.ClientSecret) ||
				checkClientSecret(secret.ClientSecret, secret2.ClientSecret)) {
			continue
		} else {
			assert.Fail(t, "secret error")
		}
	}
}

func checkAuthorizeToken(req *AuthorizeGenerateRequest, token *tokenmodels.Token) bool {
	if req.ClientID == token.ClientID &&
		req.Scope == token.Scope &&
		req.RedirectURL == token.RedirectURI &&
		req.State == token.State &&
		req.UserIdentify == token.UserID {
		return true
	}
	return false
}

func checkAccessToken(codeReq *AuthorizeGenerateRequest,
	req *OauthTokensRequest, token *tokenmodels.Token) bool {
	if token.Code != req.Code &&
		token.ClientID == req.ClientID &&
		token.RedirectURI == req.RedirectURL &&
		token.UserID == codeReq.UserIdentify &&
		token.Scope == codeReq.Scope {
		return true
	}
	return false
}

func TestOauthAuthorizeAndAccessBasic(t *testing.T) {
	// create app
	createReq := &CreateOAuthAppReq{
		Name:        "OauthTest",
		RedirectURI: "https://example.com/oauth/redirect",
		HomeURL:     "https://example.com",
		Desc:        "This is an example  oauth app",
		OwnerType:   models.GroupOwnerType,
		OwnerID:     1,
		APPType:     models.HorizonOAuthAPP,
	}
	oauthApp, err := oauthManager.CreateOauthApp(ctx, createReq)
	assert.Nil(t, err)
	assert.NotNil(t, oauthApp)
	assert.True(t, checkOAuthApp(createReq, oauthApp))

	defer func() {
		err = oauthManager.DeleteOAuthApp(ctx, oauthApp.ClientID)
		assert.Nil(t, err)
	}()

	// create secret
	secret, err := oauthManager.CreateSecret(ctx, oauthApp.ClientID)
	assert.Nil(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, secret.ClientID, oauthApp.ClientID)

	// GenAuthorizeCode
	authorizaGenerateReq := &AuthorizeGenerateRequest{
		ClientID:     oauthApp.ClientID,
		RedirectURL:  oauthApp.RedirectURL,
		State:        "dadk2sadjhkj24980",
		Scope:        "",
		UserIdentify: 43,
		Request:      nil,
	}
	codeToken, err := oauthManager.GenAuthorizeCode(ctx, authorizaGenerateReq)
	assert.Nil(t, err)
	assert.NotNil(t, codeToken)
	assert.True(t, checkAuthorizeToken(authorizaGenerateReq, codeToken))
	assert.Equal(t, codeToken.ExpiresIn, authorizeCodeExpireIn)

	// GenOauthTokens by Authorize Code
	// case1: client secret not ok
	accessTokenRequest := &OauthTokensRequest{
		ClientID:              codeToken.ClientID,
		ClientSecret:          secret.ClientSecret,
		Code:                  codeToken.Code,
		RedirectURL:           codeToken.RedirectURI,
		Request:               nil,
		AccessTokenGenerator:  generator.NewHorizonAppUserToServerAccessGenerator(),
		RefreshTokenGenerator: generator.NewRefreshTokenGenerator(),
	}

	case1Request := *accessTokenRequest
	case1Request.ClientSecret = "err-secret"
	oauthTokens, err := oauthManager.GenOauthTokens(ctx, &case1Request)
	assert.NotNil(t, err)
	if perror.Cause(err) != herrors.ErrOAuthSecretNotValid {
		assert.Fail(t, "error is not found")
	}
	assert.Nil(t, oauthTokens, nil)

	// case3: Redirect URL not right
	case3Request := *accessTokenRequest
	case3Request.RedirectURL = "err-url"
	oauthTokens, err = oauthManager.GenOauthTokens(ctx, &case3Request)
	assert.NotNil(t, err)
	if perror.Cause(err) != herrors.ErrOAuthReqNotValid {
		assert.Fail(t, "error is not found")
	}
	assert.Nil(t, oauthTokens, nil)

	// case 4: code expired
	time.Sleep(authorizeCodeExpireIn)
	case4Request := *accessTokenRequest
	_, err = oauthManager.GenOauthTokens(ctx, &case4Request)
	assert.NotNil(t, err)
	if perror.Cause(err) != herrors.ErrOAuthCodeExpired {
		assert.Fail(t, "error is not found")
	}

	// cast 5: code not exist
	case5Request := *accessTokenRequest
	_, err = oauthManager.GenOauthTokens(ctx, &case5Request)
	assert.NotNil(t, err)
	if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok || e.Source != herrors.TokenInDB {
		assert.Fail(t, "error is not found")
	}

	// case 6: ok
	authorizaGenerateReq2 := &AuthorizeGenerateRequest{
		ClientID:     oauthApp.ClientID,
		RedirectURL:  oauthApp.RedirectURL,
		State:        "dadk2sadjhkj24980",
		Scope:        "",
		UserIdentify: 43,
		Request:      nil,
	}
	codeToken, err = oauthManager.GenAuthorizeCode(ctx, authorizaGenerateReq)
	assert.Nil(t, err)
	assert.NotNil(t, codeToken)
	assert.True(t, checkAuthorizeToken(authorizaGenerateReq2, codeToken))
	assert.Equal(t, codeToken.ExpiresIn, authorizeCodeExpireIn)

	accessTokenRequest = &OauthTokensRequest{
		ClientID:              codeToken.ClientID,
		ClientSecret:          secret.ClientSecret,
		Code:                  codeToken.Code,
		RedirectURL:           codeToken.RedirectURI,
		Request:               nil,
		AccessTokenGenerator:  generator.NewHorizonAppUserToServerAccessGenerator(),
		RefreshTokenGenerator: generator.NewRefreshTokenGenerator(),
	}
	case6Request := *accessTokenRequest
	oauthTokens, err = oauthManager.GenOauthTokens(ctx, &case6Request)
	assert.Nil(t, err)
	assert.NotNil(t, oauthTokens)
	assert.True(t, checkAccessToken(authorizaGenerateReq2, &case6Request, oauthTokens.AccessToken))
	assert.True(t, checkAccessToken(authorizaGenerateReq2, &case6Request, oauthTokens.RefreshToken))
	assert.Equal(t, oauthTokens.AccessToken.ID, oauthTokens.RefreshToken.RefID)

	returnToken, err := tokenManager.LoadTokenByCode(ctx, oauthTokens.AccessToken.Code)
	assert.Nil(t, err)
	// assert.NotNil(t, returnToken)
	// assert.Equal(t, returnToken, accessToken)
	assert.True(t, returnToken.CreatedAt.Equal(oauthTokens.AccessToken.CreatedAt))
	returnToken.CreatedAt = oauthTokens.AccessToken.CreatedAt
	assert.True(t, reflect.DeepEqual(returnToken, oauthTokens.AccessToken))

	refreshToken, err := tokenManager.LoadTokenByCode(ctx, oauthTokens.RefreshToken.Code)
	assert.Nil(t, err)
	refreshToken.CreatedAt = oauthTokens.RefreshToken.CreatedAt
	assert.True(t, reflect.DeepEqual(refreshToken, oauthTokens.RefreshToken))

	assert.Nil(t, tokenManager.RevokeTokenByClientID(ctx, oauthTokens.AccessToken.ClientID))
	_, err = tokenManager.LoadTokenByCode(ctx, oauthTokens.AccessToken.Code)
	assert.NotNil(t, err)
	if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
		assert.Fail(t, "error is not found")
	}
	assert.Nil(t, tokenManager.RevokeTokenByID(ctx, oauthTokens.RefreshToken.ID))
}

func TestRefreshOauthToken(t *testing.T) {
	// create app
	createReq := &CreateOAuthAppReq{
		Name:        "refresh-token-test",
		RedirectURI: "https://refresh.com/oauth/redirect",
		HomeURL:     "https://refresh.com",
		Desc:        "This is an oauth app for testing refresh token",
		OwnerType:   models.GroupOwnerType,
		OwnerID:     1,
		APPType:     models.HorizonOAuthAPP,
	}
	oauthApp, err := oauthManager.CreateOauthApp(ctx, createReq)
	assert.Nil(t, err)
	assert.NotNil(t, oauthApp)
	assert.True(t, checkOAuthApp(createReq, oauthApp))

	defer func() {
		err = oauthManager.DeleteOAuthApp(ctx, oauthApp.ClientID)
		assert.Nil(t, err)
	}()

	// create secret
	secret, err := oauthManager.CreateSecret(ctx, oauthApp.ClientID)
	assert.Nil(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, secret.ClientID, oauthApp.ClientID)

	// GenAuthorizeCode
	authorizeGenerateReq := &AuthorizeGenerateRequest{
		ClientID:     oauthApp.ClientID,
		RedirectURL:  oauthApp.RedirectURL,
		State:        "test-state",
		Scope:        "",
		UserIdentify: 43,
		Request:      nil,
	}
	authorizeCode, err := oauthManager.GenAuthorizeCode(ctx, authorizeGenerateReq)
	assert.Nil(t, err)
	assert.NotNil(t, authorizeCode)
	assert.True(t, checkAuthorizeToken(authorizeGenerateReq, authorizeCode))
	assert.Equal(t, authorizeCode.ExpiresIn, authorizeCodeExpireIn)

	// GenOauthTokens by Authorize Code
	oauthTokensRequest := &OauthTokensRequest{
		ClientID:              authorizeCode.ClientID,
		ClientSecret:          secret.ClientSecret,
		Code:                  authorizeCode.Code,
		RedirectURL:           authorizeCode.RedirectURI,
		Request:               nil,
		AccessTokenGenerator:  generator.NewOauthAccessGenerator(),
		RefreshTokenGenerator: generator.NewRefreshTokenGenerator(),
	}
	oauthTokens, err := oauthManager.GenOauthTokens(ctx, oauthTokensRequest)
	assert.Nil(t, err)
	assert.NotNil(t, oauthTokens.AccessToken)
	assert.NotNil(t, oauthTokens.RefreshToken)

	oauthTokensRequest.RefreshToken = oauthTokens.RefreshToken.Code

	// case 1: client secret is wrong
	testReq1 := *oauthTokensRequest
	testReq1.ClientSecret = "wrong-secret"
	_, err = oauthManager.RefreshOauthTokens(ctx, &testReq1)
	assert.Equal(t, perror.Cause(err), herrors.ErrOAuthSecretNotValid)

	// case 2.1: refresh token is not valid
	testReq21 := *oauthTokensRequest
	testReq21.RefreshToken = oauthTokens.AccessToken.Code
	_, err = oauthManager.RefreshOauthTokens(ctx, &testReq21)
	assert.Equal(t, perror.Cause(err), herrors.ErrOAuthReqNotValid)

	// case 2.2: refresh token is not exist
	testReq22 := *oauthTokensRequest
	testReq22.RefreshToken = "hr_wrong-refresh-token"
	_, err = oauthManager.RefreshOauthTokens(ctx, &testReq22)
	if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok || e.Source != herrors.TokenInDB {
		assert.Fail(t, "error is not found")
	}

	// case 3: redirectURL of refresh token is mismatched
	testReq3 := *oauthTokensRequest
	testReq3.RedirectURL = "https://wrong.com"
	_, err = oauthManager.RefreshOauthTokens(ctx, &testReq3)
	assert.Equal(t, perror.Cause(err), herrors.ErrOAuthReqNotValid)

	// case 4: refresh token is expired
	time.Sleep(refreshTokenExpireIn)
	testReq4 := *oauthTokensRequest
	_, err = oauthManager.RefreshOauthTokens(ctx, &testReq4)
	assert.Equal(t, perror.Cause(err), herrors.ErrOAuthRefreshTokenExpired)

	// regenerate oauth tokens
	authorizeCode, err = oauthManager.GenAuthorizeCode(ctx, authorizeGenerateReq)
	assert.Nil(t, err)
	oauthTokensRequest = &OauthTokensRequest{
		ClientID:              authorizeCode.ClientID,
		ClientSecret:          secret.ClientSecret,
		Code:                  authorizeCode.Code,
		RedirectURL:           authorizeCode.RedirectURI,
		Request:               nil,
		AccessTokenGenerator:  generator.NewOauthAccessGenerator(),
		RefreshTokenGenerator: generator.NewRefreshTokenGenerator(),
	}
	oauthTokens, err = oauthManager.GenOauthTokens(ctx, oauthTokensRequest)
	assert.Nil(t, err)
	oauthTokensRequest.RefreshToken = oauthTokens.RefreshToken.Code

	// case 5: access token is deleted, a new access token will be generated
	assert.Nil(t, tokenManager.RevokeTokenByID(ctx, oauthTokens.AccessToken.ID))
	testReq5 := *oauthTokensRequest
	newTokens, err := oauthManager.RefreshOauthTokens(ctx, &testReq5)
	assert.Nil(t, err)
	assert.NotEqual(t, newTokens.AccessToken.ID, oauthTokens.AccessToken.ID)
	assert.NotEqual(t, newTokens.AccessToken.Code, oauthTokens.AccessToken.Code)
	assert.Equal(t, newTokens.RefreshToken.ID, oauthTokens.RefreshToken.ID)
	assert.NotEqual(t, newTokens.RefreshToken.Code, oauthTokens.RefreshToken.Code)
	assert.Equal(t, newTokens.AccessToken.ID, newTokens.RefreshToken.RefID)

	// regenerate oauth tokens
	authorizeCode, err = oauthManager.GenAuthorizeCode(ctx, authorizeGenerateReq)
	assert.Nil(t, err)
	oauthTokensRequest = &OauthTokensRequest{
		ClientID:              authorizeCode.ClientID,
		ClientSecret:          secret.ClientSecret,
		Code:                  authorizeCode.Code,
		RedirectURL:           authorizeCode.RedirectURI,
		Request:               nil,
		AccessTokenGenerator:  generator.NewOauthAccessGenerator(),
		RefreshTokenGenerator: generator.NewRefreshTokenGenerator(),
	}
	oauthTokens, err = oauthManager.GenOauthTokens(ctx, oauthTokensRequest)
	assert.Nil(t, err)
	oauthTokensRequest.RefreshToken = oauthTokens.RefreshToken.Code
	// case 6: normal case
	testReq6 := *oauthTokensRequest
	newTokens2, err := oauthManager.RefreshOauthTokens(ctx, &testReq6)
	assert.Nil(t, err)
	assert.Equal(t, newTokens2.AccessToken.ID, oauthTokens.AccessToken.ID)
	assert.NotEqual(t, newTokens2.AccessToken.Code, oauthTokens.AccessToken.Code)
	assert.Equal(t, newTokens2.RefreshToken.ID, oauthTokens.RefreshToken.ID)
	assert.NotEqual(t, newTokens2.RefreshToken.Code, oauthTokens.RefreshToken.Code)

	assert.Nil(t, tokenManager.RevokeTokenByID(ctx, newTokens2.AccessToken.ID))
	assert.Nil(t, tokenManager.RevokeTokenByID(ctx, newTokens2.RefreshToken.ID))
}

func TestUpdateToken(t *testing.T) {
	// create
	gen := generator.NewGeneralAccessTokenGenerator()
	code := gen.Generate(&generator.CodeGenerateInfo{
		Token: tokenmodels.Token{UserID: aUser.GetID()},
	})
	token := &tokenmodels.Token{
		Name:      "tokenName",
		Code:      code,
		Scope:     "clusters:read-write",
		CreatedAt: time.Now(),
		ExpiresIn: time.Hour * 24,
		UserID:    aUser.GetID(),
		RefID:     1,
	}
	tokenInDB, err := tokenManager.CreateToken(ctx, token)
	assert.Nil(t, err)
	assert.Equal(t, token.Name, tokenInDB.Name)
	assert.Equal(t, token.Code, tokenInDB.Code)
	assert.Equal(t, token.RefID, tokenInDB.RefID)

	// update
	newCode := gen.Generate(&generator.CodeGenerateInfo{
		Token: tokenmodels.Token{UserID: aUser.GetID()},
	})
	createdAt := time.Now()
	refID := uint(2)

	tokenInDB.Code = newCode
	tokenInDB.CreatedAt = createdAt
	tokenInDB.RefID = refID
	err = tokenStore.UpdateByID(ctx, tokenInDB.ID, tokenInDB)
	assert.Nil(t, err)
	tokenUpdated, err := tokenManager.LoadTokenByID(ctx, tokenInDB.ID)
	assert.Nil(t, err)
	assert.Equal(t, newCode, tokenUpdated.Code)
	assert.Equal(t, createdAt.Unix(), tokenUpdated.CreatedAt.Unix())
	assert.Equal(t, refID, tokenUpdated.RefID)

	// delete
	err = tokenManager.RevokeTokenByID(ctx, tokenInDB.ID)
	assert.Nil(t, err)
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&tokenmodels.Token{}, &models.OauthApp{}, &models.OauthClientSecret{}); err != nil {
		panic(err)
	}
	db = db.WithContext(context.WithValue(context.Background(), common.UserContextKey(), aUser))
	callbacks.RegisterCustomCallbacks(db)

	tokenStore = tokenstore.NewStore(db)
	oauthAppDAO = oauthdao.NewDAO(db)
	oauthManager = NewManager(oauthAppDAO, tokenStore, generator.NewOauthAccessGenerator(),
		authorizeCodeExpireIn, accessTokenExpireIn, refreshTokenExpireIn)
	tokenManager = tokenmanager.New(db)
	os.Exit(m.Run())
}
