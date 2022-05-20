package manager

import (
	"net/http"
	"strings"
	"time"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/oauth/generate"
	"g.hz.netease.com/horizon/pkg/oauth/models"
	"g.hz.netease.com/horizon/pkg/oauth/store"
	"g.hz.netease.com/horizon/pkg/util/log"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/util/rand"
)

type AuthorizeGenerateRequest struct {
	ClientID    string
	RedirectURL string
	State       string

	Scope        string
	UserIdentify string
	Request      *http.Request
}

type AccessTokenGenerateRequest struct {
	ClientID     string
	ClientSecret string
	Code         string // authorization code
	RedirectURL  string
	State        string

	Request *http.Request
}

type Manager interface {
	CreateOauthApp(ctx context.Context, info *CreateOAuthAppReq) (*models.OauthApp, error)
	GetOAuthApp(ctx context.Context, clientID string) (*models.OauthApp, error)
	DeleteOAuthApp(ctx context.Context, clientID string) error
	// TODO: ListOauthApp by owner

	CreateSecret(ctx context.Context, clientID string) (*models.OauthClientSecret, error)
	DeleteSecret(ctx context.Context, ClientID string, clientSecretID uint) error
	ListSecret(ctx context.Context, ClientID string) ([]models.OauthClientSecret, error)

	GenAuthorizeCode(ctx context.Context, req *AuthorizeGenerateRequest) (*models.Token, error)
	GenAccessToken(ctx context.Context, req *AccessTokenGenerateRequest,
		accessCodeGenerate generate.AccessTokenCodeGenerate) (*models.Token, error)
	RevokeAllAccessToken(ctx context.Context, clientID string) error
	LoadAccessToken(ctx context.Context, AccessToken string) (*models.Token, error)
}

var _ Manager = &manager{}

func NewManager(oauthAppStore store.OauthAppStore, tokenStore store.TokenStore,
	gen generate.AuthorizationCodeGenerate,
	authorizeCodeExpireTime time.Duration,
	accessTokenExpireTime time.Duration) *manager {
	return &manager{
		oauthAppStore:           oauthAppStore,
		tokenStore:              tokenStore,
		authorizationGenerate:   gen,
		authorizeCodeExpireTime: authorizeCodeExpireTime,
		accessTokenExpireTime:   accessTokenExpireTime,
		clientIdGenerate:        GenClientID,
	}
}

type manager struct {
	oauthAppStore           store.OauthAppStore
	tokenStore              store.TokenStore
	authorizationGenerate   generate.AuthorizationCodeGenerate
	authorizeCodeExpireTime time.Duration
	accessTokenExpireTime   time.Duration
	clientIdGenerate        ClientIDGenerate
}

const HorizonAPPClientIDPrefix = "ho_"
const BasicOauthClientLength = 20
const OauthClientSecretLength = 40

func GenClientID(appType models.AppType) string {
	if appType == models.HorizonOAuthAPP {
		return HorizonAPPClientIDPrefix + rand.String(BasicOauthClientLength)
	} else if appType == models.DirectOAuthAPP {
		return rand.String(BasicOauthClientLength)
	} else {
		return rand.String(BasicOauthClientLength)
	}
}

type ClientIDGenerate func(appType models.AppType) string

type CreateOAuthAppReq struct {
	Name        string
	RedirectURI string
	HomeURL     string
	Desc        string
	OwnerType   models.OwnerType
	OwnerID     uint
	APPType     models.AppType
}

func (m *manager) SetClientIDGenerate(gen ClientIDGenerate) {
	m.clientIdGenerate = gen
}
func (m *manager) CreateOauthApp(ctx context.Context, info *CreateOAuthAppReq) (*models.OauthApp, error) {
	clientID := m.clientIdGenerate(info.APPType)
	oauthApp := models.OauthApp{
		Name:        info.Name,
		ClientID:    clientID,
		RedirectURL: info.RedirectURI,
		HomeURL:     info.HomeURL,
		Desc:        info.Desc,
		OwnerType:   info.OwnerType,
		OwnerID:     info.OwnerID,
		AppType:     info.APPType,
	}
	if err := m.oauthAppStore.CreateApp(ctx, oauthApp); err != nil {
		return nil, err
	}
	return m.oauthAppStore.GetApp(ctx, clientID)
}

func (m *manager) GetOAuthApp(ctx context.Context, clientID string) (*models.OauthApp, error) {
	return m.oauthAppStore.GetApp(ctx, clientID)
}

func (m *manager) DeleteOAuthApp(ctx context.Context, clientID string) error {
	// revoke all the token
	if err := m.tokenStore.DeleteByClientID(ctx, clientID); err != nil {
		return err
	}

	// delete all the secret
	if err := m.oauthAppStore.DeleteSecretByClientID(ctx, clientID); err != nil {
		return err
	}
	// delete the app
	return m.oauthAppStore.DeleteApp(ctx, clientID)
}

func (m *manager) CreateSecret(ctx context.Context, clientID string) (*models.OauthClientSecret, error) {
	newSecret := &models.OauthClientSecret{
		// ID:           0, // filled by return
		ClientID:     clientID,
		ClientSecret: rand.String(OauthClientSecretLength),
		CreatedAt:    time.Now(),
		// CreatedBy:     0, // filled by middleware
	}
	return m.oauthAppStore.CreateSecret(ctx, newSecret)
}

func (m *manager) DeleteSecret(ctx context.Context, ClientID string, clientSecretID uint) error {
	return m.oauthAppStore.DeleteSecret(ctx, ClientID, clientSecretID)
}

// musk the secrets
const (
	CutPostNum = 8
	MustPrefix = "*****"
)

func MuskClientSecrets(clientSecrets []models.OauthClientSecret) {
	for i := 0; i < len(clientSecrets); i++ {
		originSecret := clientSecrets[i].ClientSecret
		muskedSecret := MustPrefix + originSecret[len(originSecret)-CutPostNum:]
		clientSecrets[i].ClientSecret = muskedSecret
	}
}
func checkMusicSecret(muskedSecret, realSecret string) bool {
	if strings.HasPrefix(muskedSecret, MustPrefix) &&
		strings.HasSuffix(realSecret, strings.TrimPrefix(muskedSecret, MustPrefix)) {
		return true
	}
	return false
}

func (m *manager) ListSecret(ctx context.Context, ClientID string) ([]models.OauthClientSecret, error) {
	clientSecrets, err := m.oauthAppStore.ListSecret(ctx, ClientID)
	if err != nil {
		return nil, err
	}

	// musk the secrets
	MuskClientSecrets(clientSecrets)

	return clientSecrets, nil
}

func (m *manager) NewAuthorizationToken(req *AuthorizeGenerateRequest) *models.Token {
	token := &models.Token{
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURL,
		State:               req.State,
		CreatedAt:           time.Now(),
		ExpiresIn:           m.authorizeCodeExpireTime,
		Scope:               req.Scope,
		UserOrRobotIdentity: req.UserIdentify,
	}
	token.Code = m.authorizationGenerate.GenCode(&generate.CodeGenerateInfo{
		Token:   *token,
		Request: req.Request,
	})
	return token
}
func (m *manager) NewAccessToken(authorizationCodeToken *models.Token,
	req *AccessTokenGenerateRequest, accessCodeGenerate generate.AccessTokenCodeGenerate) *models.Token {
	token := &models.Token{
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURL,
		State:               req.State,
		CreatedAt:           time.Now(),
		ExpiresIn:           m.accessTokenExpireTime,
		Scope:               authorizationCodeToken.Scope,
		UserOrRobotIdentity: authorizationCodeToken.UserOrRobotIdentity,
	}
	token.Code = accessCodeGenerate.GenCode(&generate.CodeGenerateInfo{
		Token:   *token,
		Request: req.Request,
	})
	return token
}

func (m *manager) GenAuthorizeCode(ctx context.Context, req *AuthorizeGenerateRequest) (*models.Token, error) {
	oauthApp, err := m.oauthAppStore.GetApp(ctx, req.ClientID)
	if err != nil {
		return nil, err
	} else if req.RedirectURL != oauthApp.RedirectURL {
		log.Warningf(ctx, "redirect URL not match")
		return nil, perror.Wrapf(herrors.ErrOAuthReqNotValid, "redirect URL not match")
	}

	authorizationToken := m.NewAuthorizationToken(req)
	err = m.tokenStore.Create(ctx, authorizationToken)
	return authorizationToken, err
}
func (m *manager) CheckByAuthorizationCode(req *AccessTokenGenerateRequest, codeToken *models.Token) error {
	if req.State != codeToken.State {
		return perror.Wrapf(herrors.ErrOAuthReqNotValid,
			"req state = %s, code state = %s", req.State, codeToken.State)
	}
	if req.RedirectURL != codeToken.RedirectURI {
		return perror.Wrapf(herrors.ErrOAuthReqNotValid,
			"req redirect url = %s, code redirect url = %s", req.RedirectURL, codeToken.RedirectURI)
	}

	if codeToken.CreatedAt.Add(m.authorizeCodeExpireTime).Before(time.Now()) {
		return perror.Wrap(herrors.ErrOAuthCodeExpired, "")
	}
	return nil
}
func (m *manager) GenAccessToken(ctx context.Context, req *AccessTokenGenerateRequest,
	accessCodeGenerate generate.AccessTokenCodeGenerate) (*models.Token, error) {
	// check client secret ok
	secrets, err := m.oauthAppStore.ListSecret(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}
	secretOk := false
	for _, secret := range secrets {
		if secret.ClientSecret == req.ClientSecret {
			secretOk = true
		}
	}
	if !secretOk {
		return nil, perror.Wrapf(herrors.ErrOAuthSecretNotValid,
			"clientId = %s, secret = %s", req.ClientID, req.ClientSecret)
	}

	// get authorize token, and check by it
	authorizationCodeToken, err := m.tokenStore.Get(ctx, req.Code)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			return nil, herrors.ErrOAuthAuthorizationCodeNotExist
		}
		return nil, err
	}

	if err := m.CheckByAuthorizationCode(req, authorizationCodeToken); err != nil {
		if perror.Cause(err) == herrors.ErrOAuthCodeExpired {
			if delErr := m.tokenStore.DeleteByCode(ctx, req.Code); delErr != nil {
				log.Warningf(ctx, "delete expired code error, err = %v", delErr)
			}
		}
		return nil, err
	}

	// new access token  and store
	accessToken := m.NewAccessToken(authorizationCodeToken, req, accessCodeGenerate)
	err = m.tokenStore.Create(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// delete authorize code
	err = m.tokenStore.DeleteByCode(ctx, req.Code)
	if err != nil {
		log.Warningf(ctx, "Delete Authorization token error, code = %s, error = %v", req.Code, err)
	}

	return accessToken, nil
}
func (m *manager) RevokeAllAccessToken(ctx context.Context, clientID string) error {
	return m.tokenStore.DeleteByClientID(ctx, clientID)
}
func (m *manager) LoadAccessToken(ctx context.Context, accessToken string) (*models.Token, error) {
	return m.tokenStore.Get(ctx, accessToken)
}
