package authenticate

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

	middleware "github.com/horizoncd/horizon/core/middleware"
	"github.com/horizoncd/horizon/pkg/config/authenticate"
	"github.com/horizoncd/horizon/pkg/log"
	"github.com/horizoncd/horizon/pkg/server/response"

	"github.com/gin-gonic/gin"
)

const (
	_timeGMT = "Mon, 02 Jan 2006 15:04:05 GMT"

	HTTPHeaderOperator    = "Operator"
	HTTPHeaderSignature   = "Signature"
	HTTPHeaderContentMD5  = "Content-MD5"
	HTTPHeaderContentType = "Content-Type"
	HTTPHeaderDate        = "Date"

	errCodeForbidden = "Forbidden"
)

func Middleware(keys authenticate.KeysConfig, skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		r := c.Request
		log.Infof(c, "request url path: %v", r.URL)

		// no operator header, skip this middleware
		operator := r.Header.Get(HTTPHeaderOperator)
		if operator == "" {
			c.Next()
			return
		}

		date := r.Header.Get(HTTPHeaderDate)
		if date == "" {
			response.Abort(c, http.StatusForbidden, errCodeForbidden, "Date Header is missing")
			return
		}
		parsedTime, err := time.Parse(_timeGMT, date)
		if err != nil {
			response.Abort(c, http.StatusForbidden, errCodeForbidden, "Invalid Date Header format")
			return
		}
		now := time.Now()
		duration := time.Minute * 10
		if parsedTime.Before(now.Add(0-duration)) || parsedTime.After(now.Add(duration)) {
			response.Abort(c, http.StatusForbidden, errCodeForbidden, "The Date has expired")
			return
		}

		if err := validatingContentMD5(c, r); err != nil {
			response.AbortWithError(c, err)
			return
		}

		signature := r.Header.Get(HTTPHeaderSignature)
		if signature == "" {
			response.Abort(c, http.StatusForbidden, errCodeForbidden, fmt.Sprintf("%v Header is missing", HTTPHeaderSignature))

			return
		}
		strs := strings.Split(signature, ":")
		if len(strs) != 2 || strings.TrimSpace(strs[0]) == "" || strings.TrimSpace(strs[1]) == "" {
			response.Abort(c, http.StatusForbidden, errCodeForbidden,
				fmt.Sprintf("Invalid %v Header format", HTTPHeaderSignature))
			return
		}

		accessKey := strings.TrimSpace(strs[0])
		secretKey := ""
		for user, ks := range keys {
			found := false
			for i := range ks {
				if ks[i].AccessKey == accessKey {
					secretKey = ks[i].SecretKey
					found = true
					break
				}
			}
			if found {
				log.Infof(c, "the caller name: %v, operator: %v", user, operator)
				break
			}
		}
		if secretKey == "" {
			response.Abort(c, http.StatusForbidden, errCodeForbidden, "invalid access key")
			return
		}

		actualSignature := signature
		expectSignature := SignRequest(c, r, accessKey, secretKey)
		if actualSignature != expectSignature {
			response.Abort(c, http.StatusForbidden, errCodeForbidden,
				fmt.Sprintf("signature verify failed on %v Header", HTTPHeaderSignature))
			return
		}

		c.Next()
	}, skippers...)
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
