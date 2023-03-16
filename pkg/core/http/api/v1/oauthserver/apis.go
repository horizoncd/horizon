package oauthserver

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/core/common"
	"github.com/horizoncd/horizon/pkg/core/controller/oauth"
	"github.com/horizoncd/horizon/pkg/core/controller/oauthapp"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/oauth/scope"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/util/log"
)

const (
	ClientID    = "client_id"
	Scope       = "scope"
	State       = "state"
	RedirectURI = "redirect_uri"
	Authorize   = "authorize"

	Code         = "code"
	ClientSecret = "client_secret"
)

const (
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
			ClientID,
			State,
			RedirectURI,
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

	appBasicInfo, err := a.oauthAppController.Get(c, c.Query(ClientID))
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
	scopeRules := a.scopeService.GetRulesByScope(strings.Split(c.Query(Scope), " "))
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
		ClientID:    c.Query(ClientID),
		HomeURL:     appBasicInfo.HomeURL,
		State:       c.Query(State),
		Scope:       c.Query(Scope),
		RedirectURL: c.Query(RedirectURI),
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
			ClientID,
			State,
			RedirectURI,
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
	value, ok := c.GetPostForm(Authorize)
	if !ok || value != Authorized {
		const DenyKey = "error"
		const DenyDesc = "the user has denied your application access"
		q := url.Values{}
		q.Set(DenyKey, DenyDesc)
		q.Set(State, c.PostForm(State))
		location := url.URL{Path: c.PostForm(RedirectURI), RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
	} else {
		resp, err := a.oAuthServer.GenAuthorizeCode(c, &oauth.AuthorizeReq{
			ClientID:     c.PostForm(ClientID),
			Scope:        c.PostForm(Scope),
			RedirectURL:  c.PostForm(RedirectURI),
			State:        c.PostForm(State),
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
			q.Set(Code, resp.Code)
			q.Set(State, resp.State)
			location := url.URL{Path: resp.RedirectURL, RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
		}
	}
}

func (a *API) HandleAccessTokenReq(c *gin.Context) {
	var err error
	checkReq := func() bool {
		keys := []string{
			ClientID,
			ClientSecret,
			RedirectURI,
			Code,
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
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	tokenResponse, err := a.oAuthServer.GenAccessToken(c, &oauth.AccessTokenReq{
		ClientID:     c.PostForm(ClientID),
		ClientSecret: c.PostForm(ClientSecret),
		Code:         c.PostForm(Code),
		RedirectURL:  c.PostForm(RedirectURI),
		Request:      c.Request,
	})
	if err != nil {
		causeErr := perror.Cause(err)
		log.Warning(c, err.Error())
		switch causeErr {
		case herrors.ErrOAuthSecretNotValid:
			fallthrough
		case herrors.ErrOAuthReqNotValid:
			fallthrough
		case herrors.ErrOAuthAuthorizationCodeNotExist:
			response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
			return
		case herrors.ErrOAuthCodeExpired:
			response.AbortWithUnauthorized(c, common.CodeExpired, err.Error())
			return
		default:
			if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				if e.Source == herrors.OAuthInDB {
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
