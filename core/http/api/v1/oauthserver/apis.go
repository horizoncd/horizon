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

package oauthserver

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/oauth"
	"github.com/horizoncd/horizon/core/controller/oauthapp"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/oauth/scope"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/util/log"
)

const (
	KeyClientID    = "client_id"
	KeyScope       = "scope"
	KeyState       = "state"
	KeyRedirectURI = "redirect_uri"
	KeyAuthorize   = "authorize"

	KeyCode         = "code"
	KeyRefreshToken = "refresh_token"
	KeyClientSecret = "client_secret"

	KeyGrantType          = "grant_type"
	GrantTypeAuthCode     = "authorization_code"
	GrantTypeRefreshToken = "refresh_token"

	Authorized = "1"
)

type API struct {
	oauthAppController oauthapp.Controller
	oAuthServer        oauth.Controller
	oauthHTMLLocation  string
	scopeService       scope.Service
}

func NewAPI(oauthServerController oauth.Controller,
	oauthAppController oauthapp.Controller, oauthHTMLLocation string, scopeService scope.Service) *API {
	return &API{
		oAuthServer:        oauthServerController,
		oauthAppController: oauthAppController,
		oauthHTMLLocation:  oauthHTMLLocation,
		scopeService:       scopeService,
	}
}

type ScopeBasic struct {
	Name string
	Desc string
}
type AuthorizationPageParams struct {
	UserName    string
	RedirectURL string
	State       string
	ClientID    string
	HomeURL     string
	Scope       string
	ClientName  string
	ScopeBasic  []ScopeBasic
}

func (a *API) HandleAuthorizationGetReq(c *gin.Context) {
	var err error
	checkReq := func() bool {
		keys := []string{
			KeyClientID,
			KeyState,
			KeyRedirectURI,
		}
		for _, key := range keys {
			if _, ok := c.GetQuery(key); !ok {
				err = fmt.Errorf("%s not exist", key)
				log.Warning(c, err.Error())
				return false
			}
		}
		return true
	}
	if !checkReq() {
		response.AbortWithRequestError(c, common.InvalidRequestBody, err.Error())
		return
	}

	appBasicInfo, err := a.oauthAppController.Get(c, c.Query(KeyClientID))
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.OAuthInDB {
				response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
				return
			}
		}
		log.Error(c, err.Error())
		response.AbortWithInternalError(c, err.Error())
		return
	}
	scopeRules := a.scopeService.GetRulesByScope(strings.Split(c.Query(KeyScope), " "))
	scopeInfo := func() []ScopeBasic {
		scopeBasics := make([]ScopeBasic, 0)
		for _, scope := range scopeRules {
			scopeBasics = append(scopeBasics, ScopeBasic{
				Name: scope.Name,
				Desc: scope.Desc,
			})
		}
		return scopeBasics
	}

	currentUser, err := common.UserFromContext(c)
	if err != nil {
		response.AbortWithForbiddenError(c, common.Forbidden, err.Error())
		return
	}

	params := AuthorizationPageParams{
		UserName:    currentUser.GetName(),
		ClientName:  appBasicInfo.AppName,
		ClientID:    c.Query(KeyClientID),
		HomeURL:     appBasicInfo.HomeURL,
		State:       c.Query(KeyState),
		Scope:       c.Query(KeyScope),
		RedirectURL: c.Query(KeyRedirectURI),
		ScopeBasic:  scopeInfo(),
	}
	authTemplate, err := template.ParseFiles(a.oauthHTMLLocation)
	if err != nil {
		log.Errorf(c, "render auth html err, err = %s", err.Error())
		response.AbortWithInternalError(c, err.Error())
		return
	}
	// var b bytes.Buffer
	c.Status(http.StatusOK)
	err = authTemplate.Execute(c.Writer, params)
	if err != nil {
		log.Errorf(c, "auth html template err, err = %s", err.Error())
	}
}

func (a *API) HandleAuthorizationReq(c *gin.Context) {
	var err error
	checkReq := func() bool {
		keys := []string{
			KeyClientID,
			KeyState,
			KeyRedirectURI,
		}
		for _, key := range keys {
			if _, ok := c.GetPostForm(key); !ok {
				err = fmt.Errorf("%s not exist", key)
				log.Warning(c, err.Error())
				return false
			}
		}
		return true
	}
	if !checkReq() {
		response.AbortWithRequestError(c, common.InvalidRequestBody, err.Error())
		return
	}
	a.handlerPostAuthorizationReq(c)
}

func (a *API) handlerPostAuthorizationReq(c *gin.Context) {
	user, err := common.UserFromContext(c)
	if err != nil {
		// TODO: redirect
		response.AbortWithForbiddenError(c, common.Forbidden, err.Error())
		return
	}
	value, ok := c.GetPostForm(KeyAuthorize)
	if !ok || value != Authorized {
		const DenyKey = "error"
		const DenyDesc = "the user has denied your application access"
		q := url.Values{}
		q.Set(DenyKey, DenyDesc)
		q.Set(KeyState, c.PostForm(KeyState))
		location := url.URL{Path: c.PostForm(KeyRedirectURI), RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
	} else {
		resp, err := a.oAuthServer.GenAuthorizeCode(c, &oauth.AuthorizeReq{
			ClientID:     c.PostForm(KeyClientID),
			Scope:        c.PostForm(KeyScope),
			RedirectURL:  c.PostForm(KeyRedirectURI),
			State:        c.PostForm(KeyState),
			UserIdentity: user.GetID(),
			Request:      c.Request,
		})
		if err != nil {
			causeErr := perror.Cause(err)
			switch causeErr {
			case herrors.ErrOAuthReqNotValid:
				log.Warning(c, err.Error())
				response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
			default:
				if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
					if e.Source == herrors.OAuthInDB {
						response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
						return
					}
				}
				log.Error(c, err.Error())
				response.AbortWithInternalError(c, err.Error())
			}
		} else {
			q := url.Values{}
			q.Set(KeyCode, resp.Code)
			q.Set(KeyState, resp.State)
			location := url.URL{Path: resp.RedirectURL, RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
		}
	}
}

func (a *API) HandleAccessTokenReq(c *gin.Context) {
	// check grant type
	grantType, ok := c.GetPostForm(KeyGrantType)
	if !ok {
		response.AbortWithRequestError(c, common.InvalidRequestParam, "grant_type not exist")
		return
	}

	keys := []string{
		KeyClientID,
		KeyClientSecret,
		KeyRedirectURI,
	}
	if grantType == GrantTypeAuthCode {
		keys = append(keys, KeyCode)
	} else if grantType == GrantTypeRefreshToken {
		keys = append(keys, KeyRefreshToken)
	} else {
		response.AbortWithRequestError(c, common.InvalidRequestParam, "grant_type not supported")
		return
	}

	// check post form keys exist
	var err error
	checkKeysExist := func(keys []string) bool {
		for _, key := range keys {
			if _, ok := c.GetPostForm(key); !ok {
				err = fmt.Errorf("%s not exist", key)
				log.Warning(c, err.Error())
				return false
			}
		}
		return true
	}
	if !checkKeysExist(keys) {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	var tokenResponse *oauth.AccessTokenResponse
	baseTokenReq := oauth.BaseTokenReq{
		ClientID:     c.PostForm(KeyClientID),
		ClientSecret: c.PostForm(KeyClientSecret),
		RedirectURL:  c.PostForm(KeyRedirectURI),
		Request:      c.Request,
	}
	if grantType == GrantTypeAuthCode {
		tokenResponse, err = a.oAuthServer.GenAccessToken(c, &oauth.AccessTokenReq{
			BaseTokenReq: baseTokenReq,
			Code:         c.PostForm(KeyCode),
		})
	} else {
		tokenResponse, err = a.oAuthServer.RefreshToken(c, &oauth.RefreshTokenReq{
			BaseTokenReq: baseTokenReq,
			RefreshToken: c.PostForm(KeyRefreshToken),
		})
	}
	if err != nil {
		causeErr := perror.Cause(err)
		log.Warning(c, err.Error())
		switch causeErr {
		case herrors.ErrOAuthSecretNotValid, herrors.ErrOAuthReqNotValid:
			response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
			return
		case herrors.ErrOAuthCodeExpired, herrors.ErrOAuthRefreshTokenExpired:
			response.AbortWithUnauthorized(c, common.CodeExpired, err.Error())
			return
		default:
			if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				if e.Source == herrors.OAuthInDB || e.Source == herrors.TokenInDB {
					response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
					return
				}
			}
			response.AbortWithInternalError(c, err.Error())
			log.Error(c, err.Error())
			return
		}
	}
	c.JSON(http.StatusOK, tokenResponse)
}
