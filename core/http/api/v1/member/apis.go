package member

import (
	"context"
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/member"
	memberctx "g.hz.netease.com/horizon/pkg/member/context"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/response"

	"github.com/gin-gonic/gin"
)

const (
	_paramGroupID              = "groupID"
	_paramApplicationID        = "applicationID"
	_paramApplicationClusterID = "clusterID"
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
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, retMember)
}

func (a *API) CreateApplicationMember(c *gin.Context) {
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
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, retMember)
}

func (a *API) CreateApplicationClusterMember(c *gin.Context) {
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
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, retMember)
}

func (a *API) UpdateMember(c *gin.Context) {
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
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, retMember)
}

func (a *API) DeleteMember(c *gin.Context) {
	memberIDStr := c.Param(_paramMemberID)
	uintID, err := strconv.ParseUint(memberIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam,
			fmt.Sprintf("memberid error,%v", err))
		return
	}
	err = a.memberCtrl.RemoveMember(c, uint(uintID))
	if err != nil {
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
		ctx = context.WithValue(c, memberctx.ContextDirectMemberOnly, directMemberOnly)
		if emailOK {
			ctx = context.WithValue(ctx, memberctx.ContextQueryOnCondition, true)
			ctx = context.WithValue(ctx, memberctx.ContextEmails, emails)
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
