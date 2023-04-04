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

package member

import (
	"context"
	"fmt"
	"strconv"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/member"
	herrors "github.com/horizoncd/horizon/core/errors"
	memberctx "github.com/horizoncd/horizon/pkg/context"
	perror "github.com/horizoncd/horizon/pkg/errors"
	membermodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"

	"github.com/gin-gonic/gin"
)

const (
	_paramGroupID              = "groupID"
	_paramApplicationID        = "applicationID"
	_paramApplicationClusterID = "clusterID"
	_paramTemplateID           = "templateID"
	_paramMemberID             = "memberID"
	_querySelf                 = "self"
	_queryEmail                = "email"
	_queryDirectMemberOnly     = "directMemberOnly"
)

type API struct {
	memberCtrl  member.Controller
	roleService role.Service
}

// NewAPI initializes a new group api
func NewAPI(controller member.Controller, rservice role.Service) *API {
	return &API{
		memberCtrl:  controller,
		roleService: rservice,
	}
}

func (a *API) CreateGroupMember(c *gin.Context) {
	op := "member: create group member"
	resourceIDStr := c.Param(_paramGroupID)
	uintID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("%v", err))
		return
	}

	var postMember *member.PostMember
	err = c.ShouldBindJSON(&postMember)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("%v", err))
		return
	}

	postMember.ResourceType = common.ResourceGroup
	postMember.ResourceID = uint(uintID)

	if err := a.validatePostMember(c, postMember); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			err.Error())
		return
	}

	retMember, err := a.memberCtrl.CreateMember(c, postMember)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if perror.Cause(err) == herrors.ErrNameConflict {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.ConflictError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, retMember)
}

func (a *API) CreateApplicationMember(c *gin.Context) {
	op := "member: create application member"
	resourceIDStr := c.Param(_paramApplicationID)
	uintID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("%v", err))
		return
	}

	var postMember *member.PostMember
	err = c.ShouldBindJSON(&postMember)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("%v", err))
		return
	}

	postMember.ResourceType = common.ResourceApplication
	postMember.ResourceID = uint(uintID)

	if err := a.validatePostMember(c, postMember); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			err.Error())
		return
	}

	retMember, err := a.memberCtrl.CreateMember(c, postMember)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if perror.Cause(err) == herrors.ErrNameConflict {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.ConflictError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, retMember)
}

func (a *API) CreateApplicationClusterMember(c *gin.Context) {
	op := "member: create cluster member"
	resourceIDStr := c.Param(_paramApplicationClusterID)
	uintID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("%v", err))
		return
	}

	var postMember *member.PostMember
	err = c.ShouldBindJSON(&postMember)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("%v", err))
		return
	}

	postMember.ResourceType = common.ResourceCluster
	postMember.ResourceID = uint(uintID)

	if err := a.validatePostMember(c, postMember); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			err.Error())
		return
	}

	retMember, err := a.memberCtrl.CreateMember(c, postMember)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if perror.Cause(err) == herrors.ErrNameConflict {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.ConflictError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, retMember)
}

func (a *API) CreateTemplateMember(c *gin.Context) {
	op := "member: create template member"
	resourceIDStr := c.Param(_paramTemplateID)
	uintID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("%v", err))
		return
	}

	var postMember *member.PostMember
	err = c.ShouldBindJSON(&postMember)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("%v", err))
		return
	}

	postMember.ResourceType = string(membermodels.TypeTemplate)
	postMember.ResourceID = uint(uintID)

	if err := a.validatePostMember(c, postMember); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			err.Error())
		return
	}

	retMember, err := a.memberCtrl.CreateMember(c, postMember)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if perror.Cause(err) == herrors.ErrNameConflict {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, postMember)
			response.AbortWithRPCError(c, rpcerror.ConflictError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, retMember)
}

func (a *API) UpdateMember(c *gin.Context) {
	op := "member: update"
	memberIDStr := c.Param(_paramMemberID)
	uintID, err := strconv.ParseUint(memberIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("memberid error,%v", err))
		return
	}
	var updateMember *member.UpdateMember
	if err = c.ShouldBindJSON(&updateMember); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("%v", err))
		return
	}

	if uint(uintID) != updateMember.ID {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			"id not match")
		return
	}

	if err := a.validRole(c, updateMember.Role); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			err.Error())
		return
	}

	retMember, err := a.memberCtrl.UpdateMember(c, updateMember.ID, updateMember.Role)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, updateMember)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, updateMember)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, retMember)
}

func (a *API) DeleteMember(c *gin.Context) {
	op := "member: delete"
	memberIDStr := c.Param(_paramMemberID)
	uintID, err := strconv.ParseUint(memberIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("memberid error,%v", err))
		return
	}
	err = a.memberCtrl.RemoveMember(c, uint(uintID))
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Warningf("err = %+v", err)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Warningf("err = %+v", err)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}

func (a *API) ListGroupMember(c *gin.Context) {
	resourceIDStr := c.Param(_paramGroupID)

	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("%v", err))
		return
	}
	a.listMember(c, resourceID, membermodels.TypeGroup)
}

func (a *API) ListApplicationMember(c *gin.Context) {
	resourceIDStr := c.Param(_paramApplicationID)

	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("%v", err))
		return
	}
	a.listMember(c, resourceID, membermodels.TypeApplication)
}

func (a *API) ListApplicationClusterMember(c *gin.Context) {
	resourceIDStr := c.Param(_paramApplicationClusterID)

	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("%v", err))
		return
	}
	a.listMember(c, resourceID, membermodels.TypeApplicationCluster)
}

func (a *API) ListTemplateMember(c *gin.Context) {
	resourceIDStr := c.Param(_paramTemplateID)

	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("%v", err))
		return
	}
	a.listMember(c, resourceID, membermodels.TypeTemplate)
}

func (a *API) listMember(c *gin.Context, resourceID uint64, resourceType membermodels.ResourceType) {
	querySelf, err := strconv.ParseBool(c.DefaultQuery(_querySelf, "false"))
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("%v", err))
		return
	}

	emails, emailOK := c.GetQueryArray(_queryEmail)
	directMemberOnly, err := strconv.ParseBool(c.DefaultQuery(_queryDirectMemberOnly, "false"))
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("%v", err))
		return
	}

	membersResp := response.DataWithTotal{}
	if querySelf {
		memberInfo, err := a.memberCtrl.GetMemberOfResource(c, string(resourceType), uint(resourceID))
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		if memberInfo != nil {
			membersResp.Items = []member.Member{*memberInfo}
			membersResp.Total = 1
		}
	} else {
		var ctx context.Context
		ctx = context.WithValue(c, memberctx.MemberDirectMemberOnly, directMemberOnly)
		if emailOK {
			ctx = context.WithValue(ctx, memberctx.MemberQueryOnCondition, true)
			ctx = context.WithValue(ctx, memberctx.MemberEmails, emails)
		}
		members, err := a.memberCtrl.ListMember(ctx, string(resourceType), uint(resourceID))
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		membersResp.Items = members
		membersResp.Total = int64(len(members))
	}
	response.SuccessWithData(c, membersResp)
}

func (a *API) validRole(ctx context.Context, role string) error {
	_, err := a.roleService.GetRole(ctx, role)
	if err != nil {
		return err
	}
	return nil
}

// validatePostMember validate postMember body according to resourceType
func (a *API) validatePostMember(ctx context.Context,
	postMember *member.PostMember) error {
	if err := validMemberType(postMember.MemberType); err != nil {
		return err
	}

	return a.validRole(ctx, postMember.Role)
}

func validMemberType(memberType membermodels.MemberType) error {
	switch memberType {
	case membermodels.MemberUser:
	case membermodels.MemberGroup:
		return fmt.Errorf("this type of member is not supported yet")
	default:
		return fmt.Errorf("invalid memberType")
	}
	return nil
}
