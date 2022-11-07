package accesstoken

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/accesstoken"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	perror "g.hz.netease.com/horizon/pkg/errors"
	scopeservice "g.hz.netease.com/horizon/pkg/oauth/scope"
	roleservice "g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
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

func (a *API) CreateAccessToken(c *gin.Context) {
	const op = "access token: create"

	var request accesstoken.CreateAccessTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("request body is invalid, err: %v", err)))
		return
	}

	resourceType := c.Param(common.ParamResourceType)
	if resourceType != "" {
		resourceIDStr := c.Param(common.ParamResourceID)
		resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
		if err != nil {
			response.AbortWithRPCError(c, rpcerror.ParamError.
				WithErrMsg(fmt.Sprintf("invalid resource id: %s", resourceIDStr)))
			return
		}
		request.Resource = &accesstoken.Resource{
			ResourceType: resourceType,
			ResourceID:   uint(resourceID),
		}
	}

	if err := a.validateCreateTokenRequest(c, request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}

	resp, err := a.accessTokenCtl.CreateAccessToken(c, request)
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, resp)
}

func (a *API) RevokeAccessToken(c *gin.Context) {
	const op = "access token: delete"

	accessTokenIDStr := c.Param(common.ParamAccessTokenID)
	id, err := strconv.ParseUint(accessTokenIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid access token id: %s", accessTokenIDStr)))
		return
	}

	err = a.accessTokenCtl.RevokeAccessToken(c, uint(id))
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}

func (a *API) ListAccessTokens(c *gin.Context) {
	const op = "access token: list"
	var (
		opts       *accesstoken.Resource
		pageNumber int64
		pageSize   int64
		err        error
	)

	resourceType := c.Param(common.ParamResourceType)
	if resourceType != "" {
		resourceIDStr := c.Param(common.ParamResourceID)
		resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
		if err != nil {
			response.AbortWithRPCError(c, rpcerror.ParamError.
				WithErrMsg(fmt.Sprintf("invalid resource id: %s", resourceIDStr)))
			return
		}
		opts = &accesstoken.Resource{
			ResourceType: resourceType,
			ResourceID:   uint(resourceID),
		}
	}

	pageNumberStr := c.Param(common.PageNumber)
	if pageNumberStr != "" {
		pageNumber, err = strconv.ParseInt(pageNumberStr, 10, 0)
		if err != nil {
			response.AbortWithRPCError(c, rpcerror.ParamError.
				WithErrMsg(fmt.Sprintf("invalid page number: %s", pageNumberStr)))
			return
		}
	}
	pageSizeStr := c.Param(common.PageSize)
	if pageSizeStr != "" {
		pageSize, err = strconv.ParseInt(pageSizeStr, 10, 0)
		if err != nil {
			response.AbortWithRPCError(c, rpcerror.ParamError.
				WithErrMsg(fmt.Sprintf("invalid page size: %s", pageSizeStr)))
			return
		}
	}

	tokens, total, err := a.accessTokenCtl.ListTokens(c, opts, &q.Query{
		PageNumber: int(pageNumber),
		PageSize:   int(pageSize),
	})
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

func (a *API) validateCreateTokenRequest(c context.Context, r accesstoken.CreateAccessTokenRequest) error {
	if err := a.validateTokenName(r.Name); err != nil {
		return err
	}

	if err := a.validateScope(r.Scopes); err != nil {
		return err
	}

	if err := a.validateExpiresAt(r.ExpiresAt); err != nil {
		return err
	}

	if r.Resource != nil {
		// rat: validate resource type
		if err := a.validateResourceType(r.Resource.ResourceType); err != nil {
			return err
		}
		if err := a.validateRole(c, r.Role); err != nil {
			return err
		}
	}
	return nil
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
	rules := a.scopeSvc.GetRulesByScope(scopes)
	if len(rules) != len(scopes) {
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("invalid scope: %s", scopes))
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
