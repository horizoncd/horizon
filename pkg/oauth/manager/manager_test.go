package manager

import (
	"os"
	"reflect"
	"testing"
	"time"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/oauth/generate"
	"g.hz.netease.com/horizon/pkg/oauth/models"
	"g.hz.netease.com/horizon/pkg/oauth/store"
	callbacks "g.hz.netease.com/horizon/pkg/util/ormcallbacks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

var (
	db            *gorm.DB
	tokenStore    store.TokenStore
	oauthAppStore store.OauthAppStore
	oauthManager  Manager
	ctx                         = context.TODO()
	aUser         userauth.User = &userauth.DefaultInfo{
		Name:     "alias",
		FullName: "alias",
		ID:       32,
		Email:    "",
		Admin:    false,
	}
	authorizeCodeExpireIn = time.Second * 3
	accessTokenExpireIn   = time.Hour * 24
)

func CheckOAuthApp(req *CreateOAuthAppReq, app *models.OauthApp) bool {
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
		HomeURL:     "https://exmple.com",
		Desc:        "This is an example  oauth app",
		OwnerType:   models.GroupOwnerType,
		OwnerID:     1,
		APPType:     models.HorizonOAuthAPP,
	}
	oauthApp, err := oauthManager.CreateOauthApp(ctx, createReq)
	assert.Nil(t, err)
	assert.NotNil(t, oauthApp)
	assert.True(t, CheckOAuthApp(createReq, oauthApp))

	oauthRetApp, err := oauthManager.GetOAuthApp(ctx, oauthApp.ClientID)
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(oauthRetApp, oauthApp))

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
		HomeURL:     "https://exmple.com",
		Desc:        "This is an example  oauth app",
		OwnerType:   models.GroupOwnerType,
		OwnerID:     1,
		APPType:     models.HorizonOAuthAPP,
	}
	oauthApp, err := oauthManager.CreateOauthApp(ctx, createReq)
	assert.Nil(t, err)
	assert.NotNil(t, oauthApp)
	assert.True(t, CheckOAuthApp(createReq, oauthApp))

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
func checkAuthorizeToken(req *AuthorizeGenerateRequest, token *models.Token) bool {
	if req.ClientID == token.ClientID &&
		req.Scope == token.Scope &&
		req.RedirectURL == token.RedirectURI &&
		req.State == token.State &&
		req.UserIdentify == token.UserOrRobotIdentity {
		return true
	}
	return false
}

func checkAccessToken(codeReq *AuthorizeGenerateRequest, req *AccessTokenGenerateRequest, token *models.Token) bool {
	if token.Code != req.Code &&
		token.State == req.State &&
		token.ClientID == req.ClientID &&
		token.RedirectURI == req.RedirectURL &&
		token.UserOrRobotIdentity == codeReq.UserIdentify &&
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
		HomeURL:     "https://exmple.com",
		Desc:        "This is an example  oauth app",
		OwnerType:   models.GroupOwnerType,
		OwnerID:     1,
		APPType:     models.HorizonOAuthAPP,
	}
	oauthApp, err := oauthManager.CreateOauthApp(ctx, createReq)
	assert.Nil(t, err)
	assert.NotNil(t, oauthApp)
	assert.True(t, CheckOAuthApp(createReq, oauthApp))

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
		UserIdentify: "43",
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
		State:        codeToken.State,
		Request:      nil,
	}
	accessCodeGen := generate.NewHorizonAppUserToServerAccessGenerate()

	case1Request := *accessTokenRequest
	case1Request.ClientSecret = "err-secret"
	accessToken, err := oauthManager.GenAccessToken(ctx, &case1Request, accessCodeGen)
	assert.NotNil(t, err)
	if perror.Cause(err) != herrors.ErrOAuthSecretNotValid {
		assert.Fail(t, "error is not found")
	}
	assert.Nil(t, accessToken, nil)

	// case2: client state not ok
	case2Request := *accessTokenRequest
	case2Request.State = "errState"
	accessToken, err = oauthManager.GenAccessToken(ctx, &case2Request, accessCodeGen)
	assert.NotNil(t, err)
	if perror.Cause(err) != herrors.ErrOAuthReqNotValid {
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
		UserIdentify: "43",
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
		State:        codeToken.State,
		Request:      nil,
	}
	case6Request := *accessTokenRequest
	accessToken, err = oauthManager.GenAccessToken(ctx, &case6Request, accessCodeGen)
	assert.Nil(t, err)
	assert.NotNil(t, accessToken)
	assert.True(t, checkAccessToken(authorizaGenerateReq2, &case6Request, accessToken))

	returnToken, err := oauthManager.LoadAccessToken(ctx, accessToken.Code)
	assert.Nil(t, err)
	// assert.NotNil(t, returnToken)
	// assert.Equal(t, returnToken, accessToken)
	assert.True(t, returnToken.CreatedAt.Equal(accessToken.CreatedAt))
	returnToken.CreatedAt = accessToken.CreatedAt
	assert.True(t, reflect.DeepEqual(returnToken, accessToken))

	assert.Nil(t, oauthManager.RevokeAllAccessToken(ctx, accessToken.ClientID))
	_, err = oauthManager.LoadAccessToken(ctx, accessToken.Code)
	assert.NotNil(t, err)
	if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
		assert.Fail(t, "error is not found")
	}
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Token{}, &models.OauthApp{}, &models.OauthClientSecret{}); err != nil {
		panic(err)
	}
	db = db.WithContext(context.WithValue(context.Background(), user.Key(), aUser))
	callbacks.RegisterCustomCallbacks(db)

	tokenStore = store.NewTokenStore(db)
	oauthAppStore = store.NewOauthAppStore(db)
	oauthManager = NewManager(oauthAppStore, tokenStore, generate.NewAuthorizeGenerate(),
		authorizeCodeExpireIn, accessTokenExpireIn)

	os.Exit(m.Run())
}
