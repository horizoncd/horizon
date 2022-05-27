package auth

import (
	"net/http"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/auth"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/rbac"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/sets"
	"github.com/gin-gonic/gin"
)

var RequestInfoFty RequestInfoFactory

func init() {
	RequestInfoFty = RequestInfoFactory{
		APIPrefixes: sets.NewString("apis"),
	}
}

func Middleware(authorizer rbac.Authorizer, skipMatchers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		// 1. get requestInfo
		requestInfo, err := RequestInfoFty.NewRequestInfo(c.Request)
		if err != nil {
			response.AbortWithRequestError(c, common.RequestInfoError, err.Error())
			return
		}
		// 2. get user
		currentUser, err := common.FromContext(c)
		if err != nil {
			response.AbortWithForbiddenError(c, common.Forbidden, err.Error())
			return
		}

		// 3. do rbac auth
		authRecord := auth.AttributesRecord{
			User:            currentUser,
			Verb:            requestInfo.Verb,
			APIGroup:        requestInfo.APIGroup,
			APIVersion:      requestInfo.APIVersion,
			Resource:        requestInfo.Resource,
			SubResource:     requestInfo.Subresource,
			Name:            requestInfo.Name,
			Scope:           requestInfo.Scope,
			ResourceRequest: requestInfo.IsResourceRequest,
			Path:            requestInfo.Path,
		}
		var decision auth.Decision
		var reason string
		decision, reason, err = authorizer.Authorize(c, authRecord)
		if err != nil {
			log.Warningf(c, "auth failed with err = %s", err.Error())
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
			if perror.Cause(err) == herrors.ErrParamInvalid {
				response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
				return
			}
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
			return
		}
		if decision == auth.DecisionDeny {
			log.Warningf(c, "denied request with reason = %s", reason)
			response.AbortWithForbiddenError(c, common.Forbidden, reason)
			return
		}
		log.Debugf(c, "passed request with reason = %s", reason)
		c.Next()
	}, skipMatchers...)
}

type RequestInfoResolver interface {
	NewRequestInfo(req *http.Request) (*RequestInfo, error)
}

type RequestInfo struct {
	// IsResourceRequest indicates whether or not the request is
	// for an API resource or subresource
	IsResourceRequest bool

	// Path is the URL path of the request
	Path string

	// Verb is the verb associated with the request for API requests.
	// not the http verb.  This includes things like list and watch.
	// for non-resource requests, this is the lowercase http verb
	Verb string

	APIPrefix  string
	APIGroup   string
	APIVersion string

	// Resource is the name of the resource being requested.
	// This is not the kind.  For example: pods
	Resource string
	Name     string

	// Subresource is the name of the subresource being requested.
	Scope       string
	Subresource string
	Parts       []string
}

type RequestInfoFactory struct {
	APIPrefixes sets.String
}

func (r *RequestInfoFactory) NewRequestInfo(req *http.Request) (*RequestInfo, error) {
	requestInfo := RequestInfo{
		IsResourceRequest: false,
		Path:              req.URL.Path,
		Verb:              strings.ToLower(req.Method),
		Scope:             req.URL.Query().Get("scope"),
	}

	currentParts := splitPath(req.URL.Path)
	if len(currentParts) < 3 {
		// return a non-resource request
		return &requestInfo, nil
	}

	if !r.APIPrefixes.Has(currentParts[0]) {
		// return a non-resource request
		return &requestInfo, nil
	}

	requestInfo.APIPrefix = currentParts[0]
	currentParts = currentParts[1:]

	requestInfo.APIGroup = currentParts[0]
	currentParts = currentParts[1:]

	requestInfo.APIVersion = currentParts[0]
	requestInfo.IsResourceRequest = true
	currentParts = currentParts[1:]

	switch req.Method {
	case "POST":
		requestInfo.Verb = "create"
	case "GET", "HEAD":
		requestInfo.Verb = "get"
	case "PUT":
		requestInfo.Verb = "update"
	case "PATCH":
		requestInfo.Verb = "patch"
	case "DELETE":
		requestInfo.Verb = "delete"
	default:
		requestInfo.Verb = ""
	}

	requestInfo.Parts = currentParts

	// parts  resource/resourceName/subresource
	switch {
	case len(requestInfo.Parts) >= 3:
		requestInfo.Subresource = requestInfo.Parts[2]
		fallthrough
	case len(requestInfo.Parts) >= 2:
		requestInfo.Name = requestInfo.Parts[1]
		fallthrough
	case len(requestInfo.Parts) >= 1:
		requestInfo.Resource = requestInfo.Parts[0]
	}

	if len(requestInfo.Name) == 0 && requestInfo.Verb == "get" {
		requestInfo.Verb = "list"
	}

	// TODO(tom): the subresource name
	// get the scope from the query
	requestInfo.Scope = req.URL.Query().Get("scope")

	return &requestInfo, nil
}

// splitPath returns the segments for a URL path.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}
