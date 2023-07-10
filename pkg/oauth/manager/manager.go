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
	perror "github.com/horizoncd/horizon/pkg/errors"
	oauthdao "github.com/horizoncd/horizon/pkg/oauth/dao"
	"github.com/horizoncd/horizon/pkg/oauth/models"
	"github.com/horizoncd/horizon/pkg/token/generator"
	tokenmodels "github.com/horizoncd/horizon/pkg/token/models"
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

type OauthTokensRequest struct {
	ClientID     string
	ClientSecret string
	Code         string // authorization code
	RefreshToken string // refresh token
	RedirectURL  string

	Request *http.Request

	AccessTokenGenerator  generator.AccessTokenCodeGenerator
	RefreshTokenGenerator generator.RefreshTokenCodeGenerator
}

type OauthTokensResponse struct {
	AccessToken  *tokenmodels.Token
	RefreshToken *tokenmodels.Token
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

type Manager interface {
	CreateOauthApp(ctx context.Context, info *CreateOAuthAppReq) (*models.OauthApp, error)
	GetOAuthApp(ctx context.Context, clientID string) (*models.OauthApp, error)
	DeleteOAuthApp(ctx context.Context, clientID string) error
	ListOauthApp(ctx context.Context, ownerType models.OwnerType, ownerID uint) ([]models.OauthApp, error)
	UpdateOauthApp(ctx context.Context, clientID string, req UpdateOauthAppReq) (*models.OauthApp, error)

	CreateSecret(ctx context.Context, clientID string) (*models.OauthClientSecret, error)
	DeleteSecret(ctx context.Context, ClientID string, clientSecretID uint) error
	ListSecret(ctx context.Context, ClientID string) ([]models.OauthClientSecret, error)

	GenAuthorizeCode(ctx context.Context, req *AuthorizeGenerateRequest) (*tokenmodels.Token, error)
	GenOauthTokens(ctx context.Context, req *OauthTokensRequest) (*OauthTokensResponse, error)
	RefreshOauthTokens(ctx context.Context, req *OauthTokensRequest) (*OauthTokensResponse, error)
}

var _ Manager = &OauthManager{}

func NewManager(oauthAppDAO oauthdao.DAO, tokenStorage tokenstorage.Storage,
	gen generator.AuthorizationCodeGenerator,
	authorizeCodeExpireTime,
	accessTokenExpireTime,
	refreshTokenExpireTime time.Duration) *OauthManager {
	return &OauthManager{
		oauthAppDAO:                oauthAppDAO,
		tokenStorage:               tokenStorage,
		authorizationCodeGenerator: gen,
		authorizeCodeExpireTime:    authorizeCodeExpireTime,
		accessTokenExpireTime:      accessTokenExpireTime,
		refreshTokenExpireTime:     refreshTokenExpireTime,
		clientIDGenerate:           GenClientID,
	}
}

type OauthManager struct {
	oauthAppDAO                oauthdao.DAO
	tokenStorage               tokenstorage.Storage
	authorizationCodeGenerator generator.AuthorizationCodeGenerator
	authorizeCodeExpireTime    time.Duration
	accessTokenExpireTime      time.Duration
	refreshTokenExpireTime     time.Duration
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

func (m *OauthManager) SetClientIDGenerate(gen ClientIDGenerate) {
	m.clientIDGenerate = gen
}
func (m *OauthManager) CreateOauthApp(ctx context.Context, info *CreateOAuthAppReq) (*models.OauthApp, error) {
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

func (m *OauthManager) GetOAuthApp(ctx context.Context, clientID string) (*models.OauthApp, error) {
	return m.oauthAppDAO.GetApp(ctx, clientID)
}

func (m *OauthManager) DeleteOAuthApp(ctx context.Context, clientID string) error {
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

func (m *OauthManager) ListOauthApp(ctx context.Context,
	ownerType models.OwnerType, ownerID uint) ([]models.OauthApp, error) {
	return m.oauthAppDAO.ListApp(ctx, ownerType, ownerID)
}

func (m *OauthManager) UpdateOauthApp(ctx context.Context, clientID string,
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

func (m *OauthManager) CreateSecret(ctx context.Context, clientID string) (*models.OauthClientSecret, error) {
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

func (m *OauthManager) DeleteSecret(ctx context.Context, ClientID string, clientSecretID uint) error {
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

func (m *OauthManager) ListSecret(ctx context.Context, ClientID string) ([]models.OauthClientSecret, error) {
	clientSecrets, err := m.oauthAppDAO.ListSecret(ctx, ClientID)
	if err != nil {
		return nil, err
	}

	// musk the secrets
	MuskClientSecrets(clientSecrets)

	return clientSecrets, nil
}

func (m *OauthManager) NewAuthorizationToken(req *AuthorizeGenerateRequest) *tokenmodels.Token {
	token := &tokenmodels.Token{
		ClientID:    req.ClientID,
		RedirectURI: req.RedirectURL,
		State:       req.State,
		CreatedAt:   time.Now(),
		ExpiresIn:   m.authorizeCodeExpireTime,
		Scope:       req.Scope,
		UserID:      req.UserIdentify,
	}
	token.Code = m.authorizationCodeGenerator.Generate(&generator.CodeGenerateInfo{
		Token:   *token,
		Request: req.Request,
	})
	return token
}
func (m *OauthManager) NewAccessToken(authorizationCodeToken *tokenmodels.Token,
	req *OauthTokensRequest) *tokenmodels.Token {
	token := &tokenmodels.Token{
		ClientID:    req.ClientID,
		RedirectURI: req.RedirectURL,
		CreatedAt:   time.Now(),
		ExpiresIn:   m.accessTokenExpireTime,
		Scope:       authorizationCodeToken.Scope,
		UserID:      authorizationCodeToken.UserID,
	}
	token.Code = req.AccessTokenGenerator.Generate(&generator.CodeGenerateInfo{
		Token:   *token,
		Request: req.Request,
	})
	return token
}

func (m *OauthManager) NewRefreshToken(accessToken *tokenmodels.Token,
	req *OauthTokensRequest) *tokenmodels.Token {
	token := &tokenmodels.Token{
		ClientID:    req.ClientID,
		RedirectURI: req.RedirectURL,
		CreatedAt:   time.Now(),
		ExpiresIn:   m.refreshTokenExpireTime,
		Scope:       accessToken.Scope,
		UserID:      accessToken.UserID,
	}
	token.Code = req.RefreshTokenGenerator.Generate(&generator.CodeGenerateInfo{
		Token:   *token,
		Request: req.Request,
	})
	return token
}

func (m *OauthManager) GenAuthorizeCode(ctx context.Context,
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

func (m *OauthManager) checkByAuthorizationCode(req *OauthTokensRequest, codeToken *tokenmodels.Token) error {
	if req.RedirectURL != codeToken.RedirectURI {
		return perror.Wrapf(herrors.ErrOAuthReqNotValid,
			"req redirect url = %s, code redirect url = %s", req.RedirectURL, codeToken.RedirectURI)
	}
	if codeToken.CreatedAt.Add(m.authorizeCodeExpireTime).Before(time.Now()) {
		return perror.Wrap(herrors.ErrOAuthCodeExpired, "")
	}
	return nil
}

func (m *OauthManager) GenOauthTokens(ctx context.Context, req *OauthTokensRequest) (*OauthTokensResponse, error) {
	// check client secret
	err := m.checkClientSecret(ctx, req)
	if err != nil {
		return nil, err
	}

	// get authorize token, and check by it
	authorizationCodeToken, err := m.tokenStorage.GetByCode(ctx, req.Code)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			return nil, herrors.ErrOAuthAuthorizationCodeNotExist
		}
		return nil, err
	}

	if err := m.checkByAuthorizationCode(req, authorizationCodeToken); err != nil {
		if perror.Cause(err) == herrors.ErrOAuthCodeExpired {
			if delErr := m.tokenStorage.DeleteByCode(ctx, req.Code); delErr != nil {
				log.Warningf(ctx, "delete expired code error, err = %v", delErr)
			}
		}
		return nil, err
	}

	// generate access token and store
	accessToken := m.NewAccessToken(authorizationCodeToken, req)
	accessTokenInDB, err := m.tokenStorage.Create(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// generate refresh token, store and associate with the access token
	refreshToken := m.NewRefreshToken(accessToken, req)
	refreshToken.RefID = accessTokenInDB.ID
	refreshTokenInDB, err := m.tokenStorage.Create(ctx, refreshToken)

	// delete authorize code
	err = m.tokenStorage.DeleteByCode(ctx, req.Code)
	if err != nil {
		log.Warningf(ctx, "Delete Authorization token error, code = %s, error = %v", req.Code, err)
	}

	return &OauthTokensResponse{
		AccessToken:  accessTokenInDB,
		RefreshToken: refreshTokenInDB,
	}, nil
}

func (m *OauthManager) RefreshOauthTokens(ctx context.Context,
	req *OauthTokensRequest) (*OauthTokensResponse, error) {
	// check client secret
	err := m.checkClientSecret(ctx, req)
	if err != nil {
		return nil, err
	}

	// check refresh token
	refreshToken, err := m.checkRefreshToken(ctx, req.RefreshToken, req.RedirectURL)
	if err != nil {
		return nil, err
	}

	// refresh associated access token
	accessToken, err := m.refreshAccessToken(ctx, refreshToken, req)
	if err != nil {
		return nil, err
	}

	// update refresh token code and ref_id
	refreshToken.Code = req.RefreshTokenGenerator.Generate(&generator.CodeGenerateInfo{
		Token:   *accessToken,
		Request: req.Request,
	})
	refreshToken.RefID = accessToken.ID
	err = m.tokenStorage.UpdateByID(ctx, refreshToken.ID, refreshToken)
	if err != nil {
		return nil, err
	}
	return &OauthTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (m *OauthManager) checkClientSecret(ctx context.Context, req *OauthTokensRequest) error {
	secrets, err := m.oauthAppDAO.ListSecret(ctx, req.ClientID)
	if err != nil {
		return err
	}
	for _, secret := range secrets {
		if secret.ClientSecret == req.ClientSecret {
			return nil
		}
	}
	return perror.Wrapf(herrors.ErrOAuthSecretNotValid,
		"clientId = %s, secret = %s", req.ClientID, req.ClientSecret)
}

func (m *OauthManager) checkRefreshToken(ctx context.Context,
	refreshToken, redirectURL string) (*tokenmodels.Token, error) {
	token, err := m.tokenStorage.GetByCode(ctx, refreshToken)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			return nil, herrors.ErrOAuthRefreshTokenNotExist
		}
		return nil, err
	}
	if redirectURL != token.RedirectURI {
		return nil, perror.Wrapf(herrors.ErrOAuthReqNotValid,
			"req redirect url = %s, token redirect url = %s", redirectURL, token.RedirectURI)
	}
	if token.CreatedAt.Add(m.refreshTokenExpireTime).Before(time.Now()) {
		return nil, perror.Wrap(herrors.ErrOAuthRefreshTokenExpired, "")
	}
	return token, nil
}

func (m *OauthManager) refreshAccessToken(ctx context.Context, refreshToken *tokenmodels.Token,
	req *OauthTokensRequest) (*tokenmodels.Token, error) {
	accessToken, err := m.tokenStorage.GetByID(ctx, refreshToken.RefID)
	accessTokenNotFound := false
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			// TODO(zhuxu): remove this log after adding token cleanup policy
			log.Warningf(ctx, "associated access token does not exist, id: %d", refreshToken.RefID)
			accessTokenNotFound = true
		} else {
			return nil, err
		}
	}
	if accessTokenNotFound {
		// generate new access token and insert to db
		token := m.NewAccessToken(&tokenmodels.Token{
			Scope:  refreshToken.Scope,
			UserID: refreshToken.UserID,
		}, req)
		accessToken, err = m.tokenStorage.Create(ctx, token)
		if err != nil {
			return nil, err
		}
	} else {
		// update token code and creation time
		accessToken.Code = req.AccessTokenGenerator.Generate(&generator.CodeGenerateInfo{
			Token:   *accessToken,
			Request: req.Request,
		})
		accessToken.CreatedAt = time.Now()
		err = m.tokenStorage.UpdateByID(ctx, accessToken.ID, accessToken)
		if err != nil {
			return nil, err
		}
	}
	return accessToken, nil
}
