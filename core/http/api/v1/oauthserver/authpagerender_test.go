package oauthserver

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

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/oauth"
	"g.hz.netease.com/horizon/core/controller/oauthapp"
	"g.hz.netease.com/horizon/lib/orm"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	oauthconfig "g.hz.netease.com/horizon/pkg/config/oauth"
	"g.hz.netease.com/horizon/pkg/oauth/generate"
	"g.hz.netease.com/horizon/pkg/oauth/manager"
	"g.hz.netease.com/horizon/pkg/oauth/models"
	scope2 "g.hz.netease.com/horizon/pkg/oauth/scope"
	"g.hz.netease.com/horizon/pkg/oauth/store"
	"g.hz.netease.com/horizon/pkg/rbac/types"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	callbacks "g.hz.netease.com/horizon/pkg/util/ormcallbacks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var (
	ctx                       = context.TODO()
	authFileLoc               = "/Users/tomsun/Workspace/cloudmusic/code/horizon/horizon/core/http/api/v1/oauthserver/auth.html" // nolint
	aUser       userauth.User = &userauth.DefaultInfo{
		Name:     "alias",
		FullName: "alias",
		ID:       32,
		Email:    "",
		Admin:    false,
	}
	authorizeCodeExpireIn = time.Minute * 30
	accessTokenExpireIn   = time.Hour * 24
)

func Test(t *testing.T) {
	params := AuthorizationPageParams{
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
func UserMiddleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		common.SetUser(c, aUser)
	}, skippers...)
}

func createOauthScopeConfig() oauthconfig.Config {
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

	return oauthconfig.Config{
		DefaultScopes: []string{"applications:read-only", "applications:read-write", "clusters:read-only"},
		Roles:         roles,
	}
}

func TestServer(t *testing.T) {
	db, _ := orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Token{}, &models.OauthApp{}, &models.OauthClientSecret{}); err != nil {
		panic(err)
	}
	db = db.WithContext(context.WithValue(context.Background(), common.Key(), aUser))
	callbacks.RegisterCustomCallbacks(db)

	tokenStore := store.NewTokenStore(db)
	oauthAppStore := store.NewOauthAppStore(db)

	oauthManager := manager.NewManager(oauthAppStore, tokenStore, generate.NewAuthorizeGenerate(),
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
	createReq := &manager.CreateOAuthAppReq{
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

	oauthServerController := oauth.NewController(oauthManager)

	oauthAppController := oauthapp.NewController(oauthManager)

	authScopeService, err := scope2.NewFileScopeService(createOauthScopeConfig())
	assert.Nil(t, err)

	api := NewAPI(oauthServerController, oauthAppController, "authFileLoc", authScopeService)

	userMiddleWare := func(c *gin.Context) {
		common.SetUser(c, aUser)
	}
	// init server
	r := gin.New()
	r.Use(userMiddleWare)
	RegisterRoutes(r, api)
	ListenPort := ":8181"

	go func() { log.Print(r.Run(ListenPort)) }()

	// wait server to start
	time.Sleep(time.Second * 5)

	// post authorize request
	scope := ""
	state := "98237dhka21dasd"
	data := url.Values{
		ClientID:    {clientID},
		Scope:       {scope},
		State:       {state},
		RedirectURL: {createReq.RedirectURI},
		Authorize:   {Authorized},
	}
	authorizeURI := BasicPath + AuthorizePath

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
	authorizeCode := parsedURL.Query().Get(Code)
	t.Logf("code = %s", authorizeCode)

	// get  the access token
	accessTokenReqData := url.Values{
		Code:         {authorizeCode},
		ClientID:     {clientID},
		ClientSecret: {secret.ClientSecret},
		RedirectURL:  {createReq.RedirectURI},
	}

	accessTokenURI := BasicPath + AccessTokenPath
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
}
