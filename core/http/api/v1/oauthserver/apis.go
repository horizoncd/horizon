package oauthserver

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/oauth"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/core/middleware/user"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

const (
	ClientID    = "client_id"
	Scope       = "scope"
	State       = "state"
	RedirectURL = "redirect_uri"
	Authorize   = "authorize"

	Code         = "code"
	ClientSecret = "client_secret"
)

type API struct {
	oAuthController   oauth.Controller
	oauthHTMLLocation string
}

func NewAPI(oauthController oauth.Controller, oauthHTMLLocation string) *API {
	return &API{
		oAuthController:   oauthController,
		oauthHTMLLocation: oauthHTMLLocation,
	}
}

type AuthorizationPageParams struct {
	RedirectURL string
	State       string
	ClientID    string
	Scope       string
	ClientName  string
}

func (a *API) HandleAuthorizationGetReq(c *gin.Context) {
	var err error
	checkReq := func() bool {
		keys := []string{
			ClientID,
			Scope,
			State,
			RedirectURL,
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

	params := AuthorizationPageParams{
		ClientID:    c.Query(ClientID),
		State:       c.Query(State),
		Scope:       c.Query(Scope),
		RedirectURL: c.Query(RedirectURL),
	}
	// authTemplate := template.Must(template.New("").ParseFiles(authFileLoc))
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
			// Scope,
			State,
			RedirectURL,
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
	user, err := user.FromContext(c)
	if err != nil {
		// TODO: redirect
		response.AbortWithForbiddenError(c, common.Forbidden, err.Error())
		return
	}
	value, ok := c.GetPostForm(Authorize)
	if !ok || value != "1" {
		const DenyKey = "error"
		const DenyDesc = "the user has denied your application access"
		q := url.Values{}
		q.Set(DenyKey, DenyDesc)
		q.Set(State, c.PostForm(State))
		location := url.URL{Path: c.PostForm(RedirectURL), RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
	} else {
		resp, err := a.oAuthController.GenAuthorizeCode(c, &oauth.AuthorizeReq{
			ClientID:     c.PostForm(ClientID),
			Scope:        c.PostForm(Scope),
			RedirectURL:  c.PostForm(RedirectURL),
			State:        c.PostForm(State),
			UserIdentity: user.GetStrID(),
			Request:      c.Request,
		})
		if err != nil {
			causeErr := perror.Cause(err)
			switch causeErr {
			case herrors.ErrOAuthReqNotValid:
				response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
			default:
				if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
					if e.Source == herrors.OAuthInDB {
						response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
						return
					}
				}
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
}
