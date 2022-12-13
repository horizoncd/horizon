package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/oauth"
	"github.com/horizoncd/horizon/core/controller/oauthapp"
	oauthcheckctl "github.com/horizoncd/horizon/core/controller/oauthcheck"
	clusterAPI "github.com/horizoncd/horizon/core/http/api/v1/cluster"
	"github.com/horizoncd/horizon/core/http/api/v1/oauthserver"
	tokenmiddle "github.com/horizoncd/horizon/core/middleware/token"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	oauthconfig "github.com/horizoncd/horizon/pkg/config/oauth"
	"github.com/horizoncd/horizon/pkg/oauth/generate"
	oauthmanager "github.com/horizoncd/horizon/pkg/oauth/manager"
	"github.com/horizoncd/horizon/pkg/oauth/models"
	"github.com/horizoncd/horizon/pkg/oauth/scope"
	"github.com/horizoncd/horizon/pkg/oauth/store"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/rbac/types"
	callbacks "github.com/horizoncd/horizon/pkg/util/ormcallbacks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var (
	authFileLoc               = ""
	aUser       userauth.User = &userauth.DefaultInfo{
		Name:     "alias",
		FullName: "alias",
		ID:       32,
		Email:    "",
		Admin:    false,
	}
	ctx                   = context.WithValue(context.Background(), common.UserContextKey(), aUser)
	authorizeCodeExpireIn = time.Minute * 30
	accessTokenExpireIn   = time.Second * 3
	manager               *managerparam.Manager
)

func Test(t *testing.T) {
	params := oauthserver.AuthorizationPageParams{
		RedirectURL: "http://overmind.com/callback",
		State:       "dasdasjl32j4sd",
		ClientID:    "as87dskkh9nsaljkdalsk",
		Scope:       "dasdasdsa",
		ClientName:  "overmind",
	}
	authTemplate, err := template.ParseFiles(authFileLoc)
	if err != nil {
		return
	}
	assert.Nil(t, err)
	err = authTemplate.Execute(os.Stdout, params)
	assert.Nil(t, err)
}

func createOauthScopeConfig() oauthconfig.Scopes {
	var roles = make([]types.Role, 0)
	roles = append(roles, types.Role{
		Name:        "applications:read-only",
		Desc:        "应用(application)及相关子资源的只读权限，包括应用配置等等",
		PolicyRules: nil,
	})
	roles = append(roles, types.Role{
		Name:        "applications:read-write",
		Desc:        "应用(application)及其相关子资源的读写删权限，包括XXX等等",
		PolicyRules: nil,
	})
	roles = append(roles, types.Role{
		Name:        "clusters:read-only",
		Desc:        "集群(cluster)及其相关子资源的读写删权限，包括XXX等等",
		PolicyRules: nil,
	})

	return oauthconfig.Scopes{
		DefaultScopes: []string{"applications:read-only", "applications:read-write", "clusters:read-only"},
		Roles:         roles,
	}
}

func TestServer(t *testing.T) {
	db, _ := orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
	if err := db.AutoMigrate(&models.Token{}, &models.OauthApp{}, &models.OauthClientSecret{}); err != nil {
		panic(err)
	}
	db = db.WithContext(context.WithValue(context.Background(), common.UserContextKey(), aUser))
	callbacks.RegisterCustomCallbacks(db)

	tokenStore := store.NewTokenStore(db)
	oauthAppStore := store.NewOauthAppStore(db)

	oauthManager := oauthmanager.NewManager(oauthAppStore, tokenStore, generate.NewAuthorizeGenerate(),
		authorizeCodeExpireIn, accessTokenExpireIn)
	clientID := "ho_t65dvkmfqb8v8xzxfbc5"
	clientIDGen := func(appType models.AppType) string {
		return clientID
	}
	secret, err := oauthManager.CreateSecret(ctx, clientID)
	assert.Nil(t, err)
	assert.NotNil(t, secret)
	t.Logf("client secret is %s", secret.ClientSecret)

	oauthManager.SetClientIDGenerate(clientIDGen)
	createReq := &oauthmanager.CreateOAuthAppReq{
		Name:        "Overmind",
		RedirectURI: "http://localhost:8083/auth/callback",
		HomeURL:     "http://localhost:8083",
		Desc:        "This is an example  oauth app",
		OwnerType:   models.GroupOwnerType,
		OwnerID:     1,
		APPType:     models.HorizonOAuthAPP,
	}
	authApp, err := oauthManager.CreateOauthApp(ctx, createReq)
	assert.Nil(t, err)
	assert.Equal(t, authApp.ClientID, clientID)

	authGetApp, err := oauthManager.GetOAuthApp(ctx, clientID)
	assert.Nil(t, err)
	assert.Equal(t, authGetApp.ClientID, authApp.ClientID)

	oauthServerController := oauth.NewController(&param.Param{Manager: manager, OauthManager: oauthManager})

	oauthAppController := oauthapp.NewController(&param.Param{Manager: manager})

	authScopeService, err := scope.NewFileScopeService(createOauthScopeConfig())
	assert.Nil(t, err)

	api := oauthserver.NewAPI(oauthServerController, oauthAppController, "authFileLoc", authScopeService)

	userMiddleWare := func(c *gin.Context) {
		common.SetUser(c, aUser)
	}
	oauthCheckerCtl := oauthcheckctl.NewOauthChecker(&param.Param{Manager: manager, OauthManager: oauthManager})
	middlewares := []gin.HandlerFunc{
		tokenmiddle.MiddleWare(oauthCheckerCtl),
		userMiddleWare,
	}

	// init server
	r := gin.New()
	r.Use(middlewares...)

	oauthserver.RegisterRoutes(r, api)
	clusterAPI.RegisterRoutes(r, nil)
	ListenPort := ":18181"

	go func() {
		log.Print(r.Run(ListenPort))
	}()

	// wait server to start
	time.Sleep(time.Second * 5)

	// post authorize request
	// nolint
	scope := ""
	// nolint
	state := "98237dhka21dasd"
	data := url.Values{
		oauthserver.ClientID:    {clientID},
		oauthserver.Scope:       {scope},
		oauthserver.State:       {state},
		oauthserver.RedirectURI: {createReq.RedirectURI},
		oauthserver.Authorize:   {oauthserver.Authorized},
	}
	authorizeURI := oauthserver.BasicPath + oauthserver.AuthorizePath

	ErrMyError := errors.New("Not Redirect Error")
	httpClient := http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return ErrMyError
	}}
	resp, err := httpClient.PostForm("http://localhost"+ListenPort+authorizeURI, data)
	assert.NotNil(t, err)
	defer resp.Body.Close()
	urlErr, ok := err.(*url.Error)
	assert.True(t, ok)
	assert.Equal(t, urlErr.Err, ErrMyError)
	assert.NotNil(t, resp)

	parsedURL, err := url.Parse(urlErr.URL)
	assert.Nil(t, err)
	authorizeCode := parsedURL.Query().Get(oauthserver.Code)
	t.Logf("code = %s", authorizeCode)

	// get  the access token
	accessTokenReqData := url.Values{
		oauthserver.Code:         {authorizeCode},
		oauthserver.ClientID:     {clientID},
		oauthserver.ClientSecret: {secret.ClientSecret},
		oauthserver.RedirectURI:  {createReq.RedirectURI},
	}

	accessTokenURI := oauthserver.BasicPath + oauthserver.AccessTokenPath
	resp, err = http.PostForm("http://localhost"+ListenPort+accessTokenURI, accessTokenReqData)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.NotNil(t, resp)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	bytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	tokenResponse := oauth.AccessTokenResponse{}
	assert.Nil(t, json.Unmarshal(bytes, &tokenResponse))
	t.Logf("%v", tokenResponse)
	switch createReq.APPType {
	case models.HorizonOAuthAPP:
		assert.True(t, strings.HasPrefix(tokenResponse.AccessToken, generate.HorizonAppUserToServerAccessTokenPrefix))
	case models.DirectOAuthAPP:
		assert.True(t, strings.HasPrefix(tokenResponse.AccessToken, generate.OauthAPPAccessTokenPrefix))
	default:
		assert.Fail(t, "unSupport")
	}

	//  token expired
	time.Sleep(accessTokenExpireIn)
	resourceURI := "/apis/core/v1/clusters/123"
	req, err := http.NewRequest("GET", "http://localhost"+ListenPort+resourceURI, nil)
	assert.Nil(t, err)
	req.Header.Set(common.AuthorizationHeaderKey, common.TokenHeaderValuePrefix+" "+tokenResponse.AccessToken)
	client := &http.Client{}
	resp, err = client.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	bytes, err = ioutil.ReadAll(resp.Body)
	t.Logf("%s", string(bytes))
	assert.True(t, strings.Contains(string(bytes), common.CodeExpired))
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusUnauthorized)
}
