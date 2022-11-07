package accesstoken

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"g.hz.netease.com/horizon/core/common"
	herror "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	accesstokenmanager "g.hz.netease.com/horizon/pkg/accesstoken/manager"
	"g.hz.netease.com/horizon/pkg/accesstoken/models"
	perror "g.hz.netease.com/horizon/pkg/errors"
	membermanager "g.hz.netease.com/horizon/pkg/member"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/oauth/generate"
	oauthmanager "g.hz.netease.com/horizon/pkg/oauth/manager"
	oauthmodels "g.hz.netease.com/horizon/pkg/oauth/models"
	"g.hz.netease.com/horizon/pkg/param"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
)

type Controller interface {
	CreateAccessToken(ctx context.Context, request CreateAccessTokenRequest) (*CreateAccessTokenResponse, error)
	RevokeAccessToken(ctx context.Context, id uint) error
	ListTokens(ctx context.Context, opts *Resource, query *q.Query) (accessTokens []AccessToken, total int, err error)
}

type controller struct {
	userMgr        usermanager.Manager
	accessTokenMgr accesstokenmanager.Manager
	oauthMgr       oauthmanager.Manager
	memberSvc      memberservice.Service
	memberMgr      membermanager.Manager
}

var _ Controller = (*controller)(nil)

func NewController(param *param.Param) Controller {
	return &controller{
		userMgr:        param.UserManager,
		accessTokenMgr: param.AccessTokenManager,
		oauthMgr:       param.OauthManager,
		memberSvc:      param.MemberService,
		memberMgr:      param.MemberManager,
	}
}
func (c *controller) CreateAccessToken(ctx context.Context,
	request CreateAccessTokenRequest) (*CreateAccessTokenResponse, error) {
	var (
		userID uint
	)

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// resource access token need robot user & member
	if request.Resource != nil {
		resourceType := request.Resource.ResourceType
		resourceID := request.Resource.ResourceID
		_, total, err := c.accessTokenMgr.ListAccessTokensOfResource(ctx, resourceType, resourceID, nil)
		if err != nil {
			return nil, err
		}

		robot := generateRobot(request.Name, resourceType, resourceID, total+1)
		robot, err = c.userMgr.Create(ctx, robot)
		if err != nil {
			return nil, err
		}

		_, err = c.memberSvc.CreateMember(ctx, memberservice.PostMember{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			MemberInfo:   robot.ID,
			MemberType:   membermodels.MemberUser,
			Role:         request.Role,
		})
		if err != nil {
			return nil, err
		}
		userID = robot.ID
	} else {
		userID = currentUser.GetID()
	}

	token, err := generateToken(request.Name, request.ExpiresAt, userID, request.Scopes)
	if err != nil {
		return nil, err
	}
	token, err = c.accessTokenMgr.CreateAccessToken(ctx, token)
	if err != nil {
		return nil, err
	}

	creator, err := c.userMgr.GetUserByID(ctx, token.CreatedBy)
	if err != nil {
		return nil, err
	}

	expiredAt := NeverExpire
	if token.ExpiresIn > 0 {
		expiredAt = token.CreatedAt.Add(token.ExpiresIn).Format(ExpiresAtFormat)
	}

	resp := &CreateAccessTokenResponse{
		AccessToken: AccessToken{
			CreateAccessTokenRequest: CreateAccessTokenRequest{
				Name:      token.Name,
				Scopes:    request.Scopes,
				ExpiresAt: expiredAt,
			},
			CreatedAt: token.CreatedAt,
			CreatedBy: &User{
				ID:    creator.ID,
				Name:  creator.Name,
				Email: creator.Email,
			},
			ID: token.ID,
		},
		Token: token.Code,
	}

	if request.Resource != nil {
		resp.Role = request.Role
	}

	return resp, nil
}

func (c *controller) RevokeAccessToken(ctx context.Context, id uint) error {
	token, err := c.accessTokenMgr.GetAccessToken(ctx, id)
	if err != nil {
		return err
	}
	userID, err := func() (uint64, error) {
		userIDStr := token.UserOrRobotIdentity
		return strconv.ParseUint(userIDStr, 10, 0)
	}()
	if err != nil {
		return err
	}
	user, err := c.userMgr.GetUserByID(ctx, uint(userID))
	if err != nil {
		return err
	}
	if user.UserType == usermodels.UserTypeRobot {
		// 1. delete member
		err = c.memberMgr.DeleteMemberByMemberNameID(ctx, uint(userID))
		if err != nil {
			return err
		}
		// 2. delete user
		err = c.userMgr.DeleteUser(ctx, uint(userID))
		if err != nil {
			return err
		}
	}
	// 3. delete token
	return c.accessTokenMgr.DeleteAccessToken(ctx, id)
}

func (c *controller) ListTokens(ctx context.Context, opts *Resource,
	query *q.Query) (accessTokens []AccessToken, total int, err error) {
	var (
		tokens []*models.AccessToken
	)
	if opts != nil {
		tokens, total, err = c.accessTokenMgr.ListAccessTokensOfResource(ctx, opts.ResourceType, opts.ResourceID, query)
		if err != nil {
			return nil, 0, err
		}
	} else {
		tokens, total, err = c.accessTokenMgr.ListOwnAccessTokens(ctx, query)
		if err != nil {
			return nil, 0, err
		}
	}

	for _, token := range tokens {
		expiredAt := NeverExpire
		if token.ExpiresIn > 0 {
			expiredAt = token.CreatedAt.Add(token.ExpiresIn).Format(ExpiresAtFormat)
		}
		creator, err := c.userMgr.GetUserByID(ctx, token.CreatedBy)
		if err != nil {
			return nil, 0, err
		}
		accessTokens = append(accessTokens, AccessToken{
			CreateAccessTokenRequest: CreateAccessTokenRequest{
				Name:      token.Name,
				Role:      token.Role,
				Scopes:    strings.Split(token.Scope, ","),
				ExpiresAt: expiredAt,
			},
			CreatedAt: token.CreatedAt,
			CreatedBy: &User{
				ID:    creator.ID,
				Name:  creator.Name,
				Email: creator.Email,
			},
			ID: token.ID,
		})
	}

	return accessTokens, total, err
}

func generateRobot(token, resourceType string, resourceID uint, cnt int) *usermodels.User {
	fullName := fmt.Sprintf("%s_%d_robot%d", resourceType, resourceID, cnt)
	name := token
	email := fmt.Sprintf("%s%s", fullName, RobotEmailSuffix)
	return &usermodels.User{
		Name:     name,
		FullName: fullName,
		Email:    email,
		UserType: usermodels.UserTypeRobot,
	}
}

func generateToken(name, expiresAtStr string, userID uint, scopes []string) (*oauthmodels.Token, error) {
	buf := bytes.NewBufferString(time.Now().String())
	code := base64.URLEncoding.EncodeToString([]byte(uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
	code = generate.AccessTokenPrefix + strings.ToUpper(strings.TrimRight(code, "="))
	createdAt := time.Now()
	expiredIn := time.Duration(0)
	if expiresAtStr != NeverExpire {
		expiredAt, err := time.Parse(ExpiresAtFormat, expiresAtStr)
		if err != nil {
			return nil, perror.Wrapf(herror.ErrParamInvalid, "invalid expiration time, error: %s", err.Error())
		}
		if !expiredAt.After(createdAt) {
			return nil, perror.Wrap(herror.ErrParamInvalid, "expiration time must be later than current time")
		}
		expiredIn = expiredAt.Sub(createdAt)
	}
	return &oauthmodels.Token{
		Name:                name,
		Code:                code,
		Scope:               strings.Join(scopes, ","),
		CreatedAt:           createdAt,
		ExpiresIn:           expiredIn,
		UserOrRobotIdentity: strconv.Itoa(int(userID)),
	}, nil
}
