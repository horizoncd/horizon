package accesstoken

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/accesstoken"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/log"
	scopeservice "github.com/horizoncd/horizon/pkg/oauth/scope"
	roleservice "github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
)

type API struct {
	accessTokenCtl accesstoken.Controller
	roleSvc        roleservice.Service
	scopeSvc       scopeservice.Service
}

func NewAPI(c accesstoken.Controller, roleSvc roleservice.Service, scopeSvc scopeservice.Service) *API {
	return &API{
		accessTokenCtl: c,
		roleSvc:        roleSvc,
		scopeSvc:       scopeSvc,
	}
}

func (a *API) CreatePersonalAccessToken(c *gin.Context) {
	const op = "access token: create pat"

	var request accesstoken.CreatePersonalAccessTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("request body is invalid, err: %v", err)))
		return
	}

	if err := a.validateCreatePersonalAccessTokenRequest(c, request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}

	resp, err := a.accessTokenCtl.CreatePersonalAccessToken(c, request)
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, resp)
}

func (a *API) CreateResourceAccessToken(c *gin.Context) {
	const op = "access token: create rat"

	var (
		request    accesstoken.CreateResourceAccessTokenRequest
		resourceID uint64
		err        error
	)
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("request body is invalid, err: %v", err)))
		return
	}

	resourceType := c.Param(common.ParamResourceType)
	resourceIDStr := c.Param(common.ParamResourceID)
	resourceID, err = strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid resource id: %s", resourceIDStr)))
		return
	}

	if err := a.validateCreateResourceAccessTokenRequest(c, request, resourceType); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}

	resp, err := a.accessTokenCtl.CreateResourceAccessToken(c, request, resourceType, uint(resourceID))
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, resp)
}

func (a *API) RevokeResourceAccessToken(c *gin.Context) {
	const op = "access token: delete"

	accessTokenIDStr := c.Param(common.ParamAccessTokenID)
	id, err := strconv.ParseUint(accessTokenIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid access token id: %s", accessTokenIDStr)))
		return
	}

	err = a.accessTokenCtl.RevokeResourceAccessToken(c, uint(id))
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}

func (a *API) RevokePersonalAccessToken(c *gin.Context) {
	const op = "access token: delete"

	accessTokenIDStr := c.Param(common.ParamAccessTokenID)
	id, err := strconv.ParseUint(accessTokenIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid access token id: %s", accessTokenIDStr)))
		return
	}

	err = a.accessTokenCtl.RevokePersonalAccessToken(c, uint(id))
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}

func (a *API) ListPersonalAccessTokens(c *gin.Context) {
	const op = "access token: list pat"
	var (
		err error
	)

	query := q.New(nil).WithPagination(c)
	tokens, total, err := a.accessTokenCtl.ListPersonalAccessTokens(c, query)
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Items: tokens,
		Total: int64(total),
	})
}

func (a *API) ListResourceAccessTokens(c *gin.Context) {
	const op = "access token: list rat"
	var (
		err error
	)

	resourceType := c.Param(common.ParamResourceType)
	resourceIDStr := c.Param(common.ParamResourceID)
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid resource id: %s", resourceIDStr)))
		return
	}

	query := q.New(nil).WithPagination(c)
	tokens, total, err := a.accessTokenCtl.ListResourceAccessTokens(c, resourceType, uint(resourceID), query)
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Items: tokens,
		Total: int64(total),
	})
}

func (a *API) validateCreatePersonalAccessTokenRequest(c context.Context,
	r accesstoken.CreatePersonalAccessTokenRequest) error {
	if err := a.validateTokenName(r.Name); err != nil {
		return err
	}

	if err := a.validateScope(r.Scopes); err != nil {
		return err
	}

	return a.validateExpiresAt(r.ExpiresAt)
}

func (a *API) validateCreateResourceAccessTokenRequest(c context.Context,
	r accesstoken.CreateResourceAccessTokenRequest, resourceType string) error {
	if err := a.validateResourceType(resourceType); err != nil {
		return err
	}

	return a.validateRole(c, r.Role)
}

func (a *API) validateTokenName(name string) error {
	if len(name) == 0 || len(name) > 40 {
		return perror.Wrap(herrors.ErrParamInvalid, "length of name should > 0 and <= 40")
	}

	pattern := `^(([a-z][-a-z0-9]*)?[a-z0-9])?$`
	reg := regexp.MustCompile(pattern)
	if !reg.MatchString(name) {
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("invalid name, regex used for validation is %v", pattern))
	}
	return nil
}

func (a *API) validateRole(ctx context.Context, role string) error {
	if _, err := a.roleSvc.GetRole(ctx, role); err != nil {
		return err
	}
	return nil
}

func (a *API) validateScope(scopes []string) error {
	scopeMap := map[string]bool{}
	validScopes := a.scopeSvc.GetAllScopeNames()
	for _, validScope := range validScopes {
		scopeMap[validScope] = true
	}
	for _, scope := range scopes {
		if _, ok := scopeMap[scope]; !ok {
			return perror.Wrap(herrors.ErrParamInvalid,
				fmt.Sprintf("invalid scope: %s", scopes))
		}
	}
	return nil
}

func (a *API) validateExpiresAt(expiresAtStr string) error {
	if expiresAtStr == "" {
		return perror.Wrap(herrors.ErrParamInvalid, "expiration time must not be empty")
	} else if expiresAtStr != accesstoken.NeverExpire {
		expiredAt, err := time.Parse(accesstoken.ExpiresAtFormat, expiresAtStr)
		if err != nil {
			return err
		}
		if !expiredAt.After(time.Now()) {
			return perror.Wrap(herrors.ErrParamInvalid, "expiration time must be later than current time")
		}
	}
	return nil
}

func (a *API) validateResourceType(resourceType string) error {
	switch resourceType {
	case common.ResourceGroup, common.ResourceApplication, common.ResourceCluster:
	default:
		return perror.Wrap(herrors.ErrParamInvalid, fmt.Sprintf("resourceType: %s is not supported", resourceType))
	}
	return nil
}
