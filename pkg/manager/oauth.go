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
	"net/http"
	"strings"
	"time"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	oauthdao "github.com/horizoncd/horizon/pkg/dao"
	perror "github.com/horizoncd/horizon/pkg/errors"
	tokenmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/oauth/models"
	"github.com/horizoncd/horizon/pkg/token/generator"
	tokenstorage "github.com/horizoncd/horizon/pkg/token/storage"
	"github.com/horizoncd/horizon/pkg/util/log"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/util/rand"
)

type AuthorizeGenerateRequest struct {
	ClientID    string
	RedirectURL string
	State       string

	Scope        string
	UserIdentify uint
	Request      *http.Request
}

type AccessTokenGenerateRequest struct {
	ClientID     string
	ClientSecret string
	Code         string // authorization code
	RedirectURL  string

	Request *http.Request
}

type CreateOAuthAppReq struct {
	Name        string
	RedirectURI string
	HomeURL     string
	Desc        string
	OwnerType   models.OwnerType
	OwnerID     uint
	APPType     models.AppType
}

type UpdateOauthAppReq struct {
	Name        string
	HomeURL     string
	RedirectURI string
	Desc        string
}

type OAuthManager interface {
	CreateOauthApp(ctx context.Context, info *CreateOAuthAppReq) (*models.OauthApp, error)
	GetOAuthApp(ctx context.Context, clientID string) (*models.OauthApp, error)
	DeleteOAuthApp(ctx context.Context, clientID string) error
	ListOauthApp(ctx context.Context, ownerType models.OwnerType, ownerID uint) ([]models.OauthApp, error)
	UpdateOauthApp(ctx context.Context, clientID string, req UpdateOauthAppReq) (*models.OauthApp, error)

	CreateSecret(ctx context.Context, clientID string) (*models.OauthClientSecret, error)
	DeleteSecret(ctx context.Context, ClientID string, clientSecretID uint) error
	ListSecret(ctx context.Context, ClientID string) ([]models.OauthClientSecret, error)

	GenAuthorizeCode(ctx context.Context, req *AuthorizeGenerateRequest) (*tokenmodels.Token, error)
	GenAccessToken(ctx context.Context, req *AccessTokenGenerateRequest,
		accessCodeGenerator generator.AccessTokenCodeGenerator) (*tokenmodels.Token, error)
}

var _ OAuthManager = &oauthManager{}

func NewOAuthManager(oauthAppDAO oauthdao.OAuthDAO, tokenStorage tokenstorage.Storage,
	gen generator.AuthorizationCodeGenerator,
	authorizeCodeExpireTime time.Duration,
	// nolint
	accessTokenExpireTime time.Duration) *oauthManager {
	return &oauthManager{
		oauthAppDAO:                oauthAppDAO,
		tokenStorage:               tokenStorage,
		authorizationCodeGenerator: gen,
		authorizeCodeExpireTime:    authorizeCodeExpireTime,
		accessTokenExpireTime:      accessTokenExpireTime,
		clientIDGenerate:           GenClientID,
	}
}

type oauthManager struct {
	oauthAppDAO                oauthdao.OAuthDAO
	tokenStorage               tokenstorage.Storage
	authorizationCodeGenerator generator.AuthorizationCodeGenerator
	authorizeCodeExpireTime    time.Duration
	accessTokenExpireTime      time.Duration
	clientIDGenerate           ClientIDGenerate
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

func (m *oauthManager) SetClientIDGenerate(gen ClientIDGenerate) {
	m.clientIDGenerate = gen
}
func (m *oauthManager) CreateOauthApp(ctx context.Context, info *CreateOAuthAppReq) (*models.OauthApp, error) {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	clientID := m.clientIDGenerate(info.APPType)
	oauthApp := models.OauthApp{
		Name:        info.Name,
		ClientID:    clientID,
		RedirectURL: info.RedirectURI,
		HomeURL:     info.HomeURL,
		Desc:        info.Desc,
		OwnerType:   info.OwnerType,
		OwnerID:     info.OwnerID,
		AppType:     info.APPType,
		CreatedBy:   user.GetID(),
		UpdatedBy:   user.GetID(),
	}
	if err := m.oauthAppDAO.CreateApp(ctx, oauthApp); err != nil {
		return nil, err
	}
	return m.oauthAppDAO.GetApp(ctx, clientID)
}

func (m *oauthManager) GetOAuthApp(ctx context.Context, clientID string) (*models.OauthApp, error) {
	return m.oauthAppDAO.GetApp(ctx, clientID)
}

func (m *oauthManager) DeleteOAuthApp(ctx context.Context, clientID string) error {
	// revoke all the token
	if err := m.tokenStorage.DeleteByClientID(ctx, clientID); err != nil {
		return err
	}

	// delete all the secret
	if err := m.oauthAppDAO.DeleteSecretByClientID(ctx, clientID); err != nil {
		return err
	}
	// delete the app
	return m.oauthAppDAO.DeleteApp(ctx, clientID)
}

func (m *oauthManager) ListOauthApp(ctx context.Context,
	ownerType models.OwnerType, ownerID uint) ([]models.OauthApp, error) {
	return m.oauthAppDAO.ListApp(ctx, ownerType, ownerID)
}

func (m *oauthManager) UpdateOauthApp(ctx context.Context, clientID string,
	req UpdateOauthAppReq) (*models.OauthApp, error) {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return m.oauthAppDAO.UpdateApp(ctx, clientID, models.OauthApp{
		Name:        req.Name,
		RedirectURL: req.RedirectURI,
		HomeURL:     req.HomeURL,
		Desc:        req.Desc,
		UpdatedBy:   user.GetID(),
	})
}

func (m *oauthManager) CreateSecret(ctx context.Context, clientID string) (*models.OauthClientSecret, error) {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	newSecret := &models.OauthClientSecret{
		// ID:           0, // filled by return
		ClientID:     clientID,
		ClientSecret: rand.String(OauthClientSecretLength),
		CreatedAt:    time.Now(),
		CreatedBy:    user.GetID(),
	}
	return m.oauthAppDAO.CreateSecret(ctx, newSecret)
}

func (m *oauthManager) DeleteSecret(ctx context.Context, ClientID string, clientSecretID uint) error {
	return m.oauthAppDAO.DeleteSecret(ctx, ClientID, clientSecretID)
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

func (m *oauthManager) ListSecret(ctx context.Context, ClientID string) ([]models.OauthClientSecret, error) {
	clientSecrets, err := m.oauthAppDAO.ListSecret(ctx, ClientID)
	if err != nil {
		return nil, err
	}

	// musk the secrets
	MuskClientSecrets(clientSecrets)

	return clientSecrets, nil
}

func (m *oauthManager) NewAuthorizationToken(req *AuthorizeGenerateRequest) *tokenmodels.Token {
	token := &tokenmodels.Token{
		ClientID:    req.ClientID,
		RedirectURI: req.RedirectURL,
		State:       req.State,
		CreatedAt:   time.Now(),
		ExpiresIn:   m.authorizeCodeExpireTime,
		Scope:       req.Scope,
		UserID:      req.UserIdentify,
	}
	token.Code = m.authorizationCodeGenerator.GenCode(&generator.CodeGenerateInfo{
		Token:   *token,
		Request: req.Request,
	})
	return token
}
func (m *oauthManager) NewAccessToken(authorizationCodeToken *tokenmodels.Token,
	req *AccessTokenGenerateRequest, accessCodeGenerator generator.AccessTokenCodeGenerator) *tokenmodels.Token {
	token := &tokenmodels.Token{
		ClientID:    req.ClientID,
		RedirectURI: req.RedirectURL,
		CreatedAt:   time.Now(),
		ExpiresIn:   m.accessTokenExpireTime,
		Scope:       authorizationCodeToken.Scope,
		UserID:      authorizationCodeToken.UserID,
	}
	token.Code = accessCodeGenerator.GenCode(&generator.CodeGenerateInfo{
		Token:   *token,
		Request: req.Request,
	})
	return token
}

func (m *oauthManager) GenAuthorizeCode(ctx context.Context,
	req *AuthorizeGenerateRequest) (*tokenmodels.Token, error) {
	oauthApp, err := m.oauthAppDAO.GetApp(ctx, req.ClientID)
	if err != nil {
		return nil, err
	} else if req.RedirectURL != oauthApp.RedirectURL {
		log.Warningf(ctx, "redirect URL not match")
		return nil, perror.Wrapf(herrors.ErrOAuthReqNotValid, "redirect URL not match")
	}

	authorizationToken := m.NewAuthorizationToken(req)
	_, err = m.tokenStorage.Create(ctx, authorizationToken)
	return authorizationToken, err
}

func (m *oauthManager) CheckByAuthorizationCode(req *AccessTokenGenerateRequest, codeToken *tokenmodels.Token) error {
	if req.RedirectURL != codeToken.RedirectURI {
		return perror.Wrapf(herrors.ErrOAuthReqNotValid,
			"req redirect url = %s, code redirect url = %s", req.RedirectURL, codeToken.RedirectURI)
	}
	if codeToken.CreatedAt.Add(m.authorizeCodeExpireTime).Before(time.Now()) {
		return perror.Wrap(herrors.ErrOAuthCodeExpired, "")
	}
	return nil
}

func (m *oauthManager) GenAccessToken(ctx context.Context, req *AccessTokenGenerateRequest,
	accessCodeGenerator generator.AccessTokenCodeGenerator) (*tokenmodels.Token, error) {
	// check client secret ok
	secrets, err := m.oauthAppDAO.ListSecret(ctx, req.ClientID)
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
	authorizationCodeToken, err := m.tokenStorage.GetByCode(ctx, req.Code)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			return nil, herrors.ErrOAuthAuthorizationCodeNotExist
		}
		return nil, err
	}

	if err := m.CheckByAuthorizationCode(req, authorizationCodeToken); err != nil {
		if perror.Cause(err) == herrors.ErrOAuthCodeExpired {
			if delErr := m.tokenStorage.DeleteByCode(ctx, req.Code); delErr != nil {
				log.Warningf(ctx, "delete expired code error, err = %v", delErr)
			}
		}
		return nil, err
	}

	// new access token and store
	accessToken := m.NewAccessToken(authorizationCodeToken, req, accessCodeGenerator)
	_, err = m.tokenStorage.Create(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// delete authorize code
	err = m.tokenStorage.DeleteByCode(ctx, req.Code)
	if err != nil {
		log.Warningf(ctx, "Delete Authorization token error, code = %s, error = %v", req.Code, err)
	}

	return accessToken, nil
}
