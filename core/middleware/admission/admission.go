package admission

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/middleware"
	admissionwebhook "github.com/horizoncd/horizon/pkg/admission"
	admissionmodels "github.com/horizoncd/horizon/pkg/admission/models"
	"github.com/horizoncd/horizon/pkg/auth"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
)

// Middleware to validate and mutate admission request
func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		// get auth record
		record, ok := c.Get(common.ContextAuthRecord)
		if !ok {
			response.AbortWithRPCError(c,
				rpcerror.BadRequestError.WithErrMsg("request with no auth record"))
			return
		}
		attr := record.(auth.AttributesRecord)
		// non resource request or read only request should be ignored
		if !attr.IsResourceRequest() || attr.IsReadOnly() {
			c.Next()
			return
		}
		var object interface{}
		// read request body and avoid side-effects on c.Request.Body
		bodyBytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			response.AbortWithRPCError(c,
				rpcerror.ParamError.WithErrMsg(fmt.Sprintf("request body is invalid, err: %v", err)))
			return
		}
		if len(bodyBytes) > 0 {
			contentType := c.ContentType()
			if contentType == binding.MIMEJSON || contentType == "" {
				if err := json.Unmarshal(bodyBytes, &object); err != nil {
					response.AbortWithRPCError(c,
						rpcerror.ParamError.WithErrMsg(fmt.Sprintf("unmarshal request body failed, err: %v", err)))
					return
				}
			} else {
				log.Errorf(c, "unsupported content type: %s", contentType)
				response.AbortWithRPCError(c,
					rpcerror.ParamError.WithErrMsg(fmt.Sprintf("unsupported content type: %s", contentType)))
				return
			}
		}
		if object != nil {
			// fill in the request url query into admission request options
			queries := c.Request.URL.Query()
			options := make(map[string]interface{}, len(queries))
			for k, v := range queries {
				if len(v) == 1 {
					options[k] = v[0]
				} else {
					options[k] = v
				}
			}
			admissionRequest := &admissionwebhook.Request{
				Operation:   admissionmodels.Operation(attr.GetVerb()),
				Resource:    attr.GetResource(),
				Name:        attr.GetName(),
				SubResource: attr.GetSubResource(),
				Version:     attr.GetAPIVersion(),
				Object:      object,
				Options:     options,
			}
			admissionRequest, err = admissionwebhook.Mutating(c, admissionRequest)
			if err != nil {
				response.AbortWithRPCError(c,
					rpcerror.ParamError.WithErrMsg(fmt.Sprintf("admission mutating failed: %v", err)))
				return
			}
			if err := admissionwebhook.Validating(c, admissionRequest); err != nil {
				response.AbortWithRPCError(c,
					rpcerror.ParamError.WithErrMsg(fmt.Sprintf("admission validating failed: %v", err)))
				return
			}
			newBodyBytes, err := json.Marshal(admissionRequest.Object)
			if err != nil {
				response.AbortWithRPCError(c,
					rpcerror.ParamError.WithErrMsg(fmt.Sprintf("marshal request body failed, err: %v", err)))
				return
			}
			// restore the request body
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(newBodyBytes))
			c.Next()
		} else {
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			c.Next()
		}
	}, skippers...)
}
