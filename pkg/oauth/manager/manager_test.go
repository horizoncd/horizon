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
	tokenstorage "github.com/horizoncd/horizon/pkg/token/storage"
	callbacks "github.com/horizoncd/horizon/pkg/util/ormcallbacks"
)

var (
	db           *gorm.DB
	tokenStorage tokenstorage.Storage
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
	accessTokenExpireIn   = time.Hour * 24
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
			(checkMusicSecret(secret.ClientSecret, secret1.ClientSecret) ||
				checkMusicSecret(secret.ClientSecret, secret2.ClientSecret)) {
			// noop
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
	req *AccessTokenGenerateRequest, token *tokenmodels.Token) bool {
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

	// GenAccessToken by  Authorize Code
	// case1: client secret not ok
	accessTokenRequest := &AccessTokenGenerateRequest{
		ClientID:     codeToken.ClientID,
		ClientSecret: secret.ClientSecret,
		Code:         codeToken.Code,
		RedirectURL:  codeToken.RedirectURI,
		Request:      nil,
	}
	accessCodeGen := generator.NewHorizonAppUserToServerAccessGenerator()

	case1Request := *accessTokenRequest
	case1Request.ClientSecret = "err-secret"
	accessToken, err := oauthManager.GenAccessToken(ctx, &case1Request, accessCodeGen)
	assert.NotNil(t, err)
	if perror.Cause(err) != herrors.ErrOAuthSecretNotValid {
		assert.Fail(t, "error is not found")
	}
	assert.Nil(t, accessToken, nil)

	// case3: Redirect URL not right
	case3Request := *accessTokenRequest
	case3Request.RedirectURL = "err-url"
	accessToken, err = oauthManager.GenAccessToken(ctx, &case3Request, accessCodeGen)
	assert.NotNil(t, err)
	if perror.Cause(err) != herrors.ErrOAuthReqNotValid {
		assert.Fail(t, "error is not found")
	}
	assert.Nil(t, accessToken, nil)

	// case 4: code expired
	time.Sleep(authorizeCodeExpireIn)
	case4Request := *accessTokenRequest
	_, err = oauthManager.GenAccessToken(ctx, &case4Request, accessCodeGen)
	assert.NotNil(t, err)
	if perror.Cause(err) != herrors.ErrOAuthCodeExpired {
		assert.Fail(t, "error is not found")
	}

	// cast 5: code not exist
	case5Request := *accessTokenRequest
	_, err = oauthManager.GenAccessToken(ctx, &case5Request, accessCodeGen)
	assert.NotNil(t, err)
	if perror.Cause(err) != herrors.ErrOAuthAuthorizationCodeNotExist {
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

	accessTokenRequest = &AccessTokenGenerateRequest{
		ClientID:     codeToken.ClientID,
		ClientSecret: secret.ClientSecret,
		Code:         codeToken.Code,
		RedirectURL:  codeToken.RedirectURI,
		Request:      nil,
	}
	case6Request := *accessTokenRequest
	accessToken, err = oauthManager.GenAccessToken(ctx, &case6Request, accessCodeGen)
	assert.Nil(t, err)
	assert.NotNil(t, accessToken)
	assert.True(t, checkAccessToken(authorizaGenerateReq2, &case6Request, accessToken))

	returnToken, err := tokenManager.LoadTokenByCode(ctx, accessToken.Code)
	assert.Nil(t, err)
	// assert.NotNil(t, returnToken)
	// assert.Equal(t, returnToken, accessToken)
	assert.True(t, returnToken.CreatedAt.Equal(accessToken.CreatedAt))
	returnToken.CreatedAt = accessToken.CreatedAt
	assert.True(t, reflect.DeepEqual(returnToken, accessToken))

	assert.Nil(t, tokenManager.RevokeTokenByClientID(ctx, accessToken.ClientID))
	_, err = tokenManager.LoadTokenByCode(ctx, accessToken.Code)
	assert.NotNil(t, err)
	if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
		assert.Fail(t, "error is not found")
	}
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&tokenmodels.Token{}, &models.OauthApp{}, &models.OauthClientSecret{}); err != nil {
		panic(err)
	}
	db = db.WithContext(context.WithValue(context.Background(), common.UserContextKey(), aUser))
	callbacks.RegisterCustomCallbacks(db)

	tokenStorage = tokenstorage.NewStorage(db)
	oauthAppDAO = oauthdao.NewDAO(db)
	oauthManager = NewManager(oauthAppDAO, tokenStorage, generator.NewOauthAccessGenerator(),
		authorizeCodeExpireIn, accessTokenExpireIn)
	tokenManager = tokenmanager.New(db)
	os.Exit(m.Run())
}
