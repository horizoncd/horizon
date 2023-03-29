package user

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/horizoncd/horizon/pkg/core/common"
	coreconfig "github.com/horizoncd/horizon/pkg/core/config"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	"github.com/horizoncd/horizon/pkg/core/middleware"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/config/authenticate"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	"github.com/horizoncd/horizon/pkg/user/models"
	"github.com/horizoncd/horizon/pkg/util/log"
)

const (
	_timeGMT = "Mon, 02 Jan 2006 15:04:05 GMT"

	HTTPHeaderOperator    = "Operator"
	HTTPHeaderSignature   = "Signature"
	HTTPHeaderContentMD5  = "Content-MD5"
	HTTPHeaderContentType = "Content-Type"
	HTTPHeaderDate        = "Date"

	NotAuthHeader = "X-OIDC-Redirect-To"
)

// Middleware check user is exists in db. If not, add user into db.
// Then attach a User object into context.
func Middleware(param *param.Param, store sessions.Store,
	config *coreconfig.Config, skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		// 1. aksk auth if operator header exists
		user, err := akskAuthn(c, config.AccessSecretKeys, param.UserManager)
		if err != nil {
			response.AbortWithRPCError(c, rpcerror.ForbiddenError.WithErrMsg(err.Error()))
			return
		}

		if user != nil {
			c.Set(common.UserContextKey(), &userauth.DefaultInfo{
				Name:     user.Name,
				FullName: user.FullName,
				ID:       user.ID,
				Email:    user.Email,
				Admin:    user.Admin,
			})
			c.Next()
			return
		}

		// 2. token auth request ( get user by token)
		if _, err := common.GetToken(c); err == nil {
			c.Next()
			return
		}

		session, err := store.Get(c.Request, common.CookieKeyAuth)
		if err != nil {
			response.Abort(c, http.StatusUnauthorized,
				http.StatusText(http.StatusUnauthorized),
				fmt.Sprintf("session is not found\n"+
					"session name = %s\n, err = %v", common.CookieKeyAuth, err))
			return
		}

		u := session.Values[common.SessionKeyAuthUser]
		if user, ok := u.(*userauth.DefaultInfo); ok && user != nil {
			// attach user to context
			common.SetUser(c, user)
			c.Next()
			return
		}

		// default status code of response is 200,
		// if status is not 200, that means it has been handled by other middleware,
		// so just omit it.
		if c.Writer.Status() != http.StatusOK ||
			// if not login, call this to login
			// if signed in, call this to link other api
			c.Request.URL.Path == common.URLLoginCallback {
			c.Next()
			return
		}

		if c.Request.URL.Path == common.URLOauthAuthorization {
			c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?redirect=%s",
				common.URLFrontLogin, url.QueryEscape(c.Request.RequestURI)))
			c.AbortWithStatus(http.StatusTemporaryRedirect)
			return
		}

		c.Header(NotAuthHeader, "NotAuth")
		response.AbortWithRPCError(c, rpcerror.Unauthorized.WithErrMsg("please login"))
	}, skippers...)
}

func akskAuthn(c *gin.Context, keys authenticate.KeysConfig, userMgr usermanager.Manager) (*models.User, error) {
	r := c.Request
	log.Infof(c, "request url path: %v", r.URL)

	operator := r.Header.Get(HTTPHeaderOperator)
	if operator == "" {
		return nil, nil
	}

	date := r.Header.Get(HTTPHeaderDate)
	if date == "" {
		return nil, perror.Wrap(herrors.ErrParamInvalid, "Date Header is missing")
	}
	parsedTime, err := time.Parse(_timeGMT, date)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid, "Invalid Date Header format")
	}
	now := time.Now()
	duration := time.Minute * 10
	if parsedTime.Before(now.Add(0-duration)) || parsedTime.After(now.Add(duration)) {
		return nil, perror.Wrap(herrors.ErrParamInvalid, "The Date has expired")
	}

	if err := validatingContentMD5(c, r); err != nil {
		return nil, err
	}

	signature := r.Header.Get(HTTPHeaderSignature)
	if signature == "" {
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "%v Header is missing", HTTPHeaderSignature)
	}
	strs := strings.Split(signature, ":")
	if len(strs) != 2 || strings.TrimSpace(strs[0]) == "" || strings.TrimSpace(strs[1]) == "" {
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "Invalid %v Header format", HTTPHeaderSignature)
	}

	var key *authenticate.Key
	accessKey := strings.TrimSpace(strs[0])
	secretKey := ""
	for user, ks := range keys {
		found := false
		for i := range ks {
			if ks[i].AccessKey == accessKey {
				secretKey = ks[i].SecretKey
				found = true
				key = ks[i]
				break
			}
		}
		if found {
			log.Infof(c, "the caller name: %v, operator: %v", user, operator)
			break
		}
	}
	if secretKey == "" {
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "invalid access key")
	}

	actualSignature := signature
	expectSignature := SignRequest(c, r, accessKey, secretKey)
	if actualSignature != expectSignature {
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "signature verify failed on %v Header", HTTPHeaderSignature)
	}

	u, err := userMgr.GetUserByIDP(c, operator, key.IDP)
	if err != nil {
		return nil, perror.Wrapf(herrors.NewErrGetFailed(herrors.UserInDB, "unauthorized"),
			"failed to get user with email = %v and idp = %v: err = %v", operator, key.IDP, err)
	}
	if u == nil {
		return nil, perror.Wrapf(herrors.NewErrNotFound(herrors.UserInDB, "unauthorized"),
			"user with email = %v and idp = %v not found", operator, key.IDP)
	}

	return u, nil
}

func validatingContentMD5(ctx context.Context, r *http.Request) error {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf(ctx, err.Error())
		return err
	}
	log.Infof(ctx, "data: %v", string(data))
	defer func() { _ = r.Body.Close() }()
	r.Body = ioutil.NopCloser(bytes.NewReader(data))

	if len(data) > 0 {
		log.Infof(ctx, "request body: %v", string(data))
		contentMD5 := r.Header.Get(HTTPHeaderContentMD5)
		if contentMD5 == "" {
			return fmt.Errorf("Content-MD5 Header is missing")
		}

		hash := md5.Sum(data)
		buf := base64.StdEncoding.EncodeToString(hash[:])
		if buf != contentMD5 {
			return fmt.Errorf("Content-MD5 Header is invalid, content:%v, expect:%v, actual:%v",
				string(data), buf, contentMD5)
		}
	}
	return nil
}

func SignRequest(ctx context.Context, request *http.Request, publicKey string, secretKey string) string {
	/*
	 * StringToSign = HTTP-Verb + "\n" +
	 *                PATH + "\n" +
	 *                Content-MD5 + "\n" +
	 *                Content-Type + "\n" +
	 *                Date + "\n"
	 */
	var stringToSign string
	stringToSign += request.Method + "\n"
	if strings.HasPrefix(request.URL.Path, "/") {
		stringToSign += "/" + url.QueryEscape(request.URL.Path[1:]) + "\n"
	} else {
		stringToSign += "/" + url.QueryEscape(request.URL.Path) + "\n"
	}
	stringToSign += request.Header.Get(HTTPHeaderContentMD5) + "\n"
	stringToSign += request.Header.Get(HTTPHeaderContentType) + "\n"
	stringToSign += request.Header.Get(HTTPHeaderDate) + "\n"
	log.Debugf(ctx, "stringToSign:\n%v", stringToSign)

	key := []byte(secretKey)
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(stringToSign))
	if err != nil {
		log.Errorf(ctx, "error: %v", err)
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	log.Debugf(ctx, "publicKey:%v, signature:\n%v", publicKey, signature)

	return publicKey + ":" + signature
}
