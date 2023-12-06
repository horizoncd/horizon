package admission

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/middleware"
	admissionwebhook "github.com/horizoncd/horizon/pkg/admission"
	admissionmodels "github.com/horizoncd/horizon/pkg/admission/models"
	"github.com/horizoncd/horizon/pkg/auth"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
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
		if err := c.ShouldBind(object); err != nil {
			response.AbortWithRPCError(c,
				rpcerror.ParamError.WithErrMsg(fmt.Sprintf("request body is invalid, err: %v", err)))
			return
		}
		admissionRequest := &admissionwebhook.Request{
			Operation:    admissionmodels.Operation(attr.GetVerb()),
			Resource:     attr.GetResource(),
			ResourceName: attr.GetName(),
			SubResource:  attr.GetSubResource(),
			Version:      attr.GetAPIVersion(),
			Object:       object,
			OldObject:    nil,
		}
		if err := admissionwebhook.Validating(c, admissionRequest); err != nil {
			response.AbortWithRPCError(c,
				rpcerror.ParamError.WithErrMsg(fmt.Sprintf("admission validating failed: %v", err)))
			return
		}
		c.Next()
	}, skippers...)
}
