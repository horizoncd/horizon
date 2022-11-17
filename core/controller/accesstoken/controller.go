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
	CreateResourceAccessToken(ctx context.Context, request CreateResourceAccessTokenRequest,
		resourceType string, resourceID uint) (*CreateResourceAccessTokenResponse, error)
	CreatePersonalAccessToken(ctx context.Context,
		request CreatePersonalAccessTokenRequest) (*CreatePersonalAccessTokenResponse, error)
	ListPersonalAccessTokens(ctx context.Context,
		query *q.Query) (accessTokens []PersonalAccessToken, total int, err error)
	ListResourceAccessTokens(ctx context.Context, resourceType string,
		reosurceID uint, query *q.Query) (accessTokens []ResourceAccessToken, total int, err error)
	RevokePersonalAccessToken(ctx context.Context, id uint) error
	RevokeResourceAccessToken(ctx context.Context, id uint) error
}

type controller struct {
	userMgr        usermanager.Manager
	accessTokenMgr accesstokenmanager.Manager
	oauthMgr       oauthmanager.Manager
	memberSvc      memberservice.Service
	memberMgr      membermanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		userMgr:        param.UserManager,
		accessTokenMgr: param.AccessTokenManager,
		oauthMgr:       param.OauthManager,
		memberSvc:      param.MemberService,
		memberMgr:      param.MemberManager,
	}
}

func (c *controller) CreateResourceAccessToken(ctx context.Context, request CreateResourceAccessTokenRequest,
	resourceType string, resourceID uint) (*CreateResourceAccessTokenResponse, error) {
	var (
		userID uint
	)

	// resource access token need robot user & member
	robot := generateRobot(request.Name, resourceType, resourceID)
	robot, err := c.userMgr.Create(ctx, robot)
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

	token, err := generateToken(request.Name, request.ExpiresAt, userID, request.Scopes)
	if err != nil {
		return nil, err
	}
	token, err = c.accessTokenMgr.CreateAccessToken(ctx, token)
	if err != nil {
		return nil, err
	}

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	resp := &CreateResourceAccessTokenResponse{
		ResourceAccessToken: ResourceAccessToken{
			CreateResourceAccessTokenRequest: CreateResourceAccessTokenRequest{
				CreatePersonalAccessTokenRequest: CreatePersonalAccessTokenRequest{
					Name:      token.Name,
					Scopes:    request.Scopes,
					ExpiresAt: parseExpiredAt(token.CreatedAt, token.ExpiresIn),
				},
				Role: request.Role,
			},
			CreatedAt: token.CreatedAt,
			CreatedBy: &usermodels.UserBasic{
				ID:    currentUser.GetID(),
				Name:  currentUser.GetName(),
				Email: currentUser.GetEmail(),
			},
			ID: token.ID,
		},
		Token: token.Code,
	}

	return resp, nil
}

func (c *controller) CreatePersonalAccessToken(ctx context.Context,
	request CreatePersonalAccessTokenRequest) (*CreatePersonalAccessTokenResponse, error) {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	token, err := generateToken(request.Name, request.ExpiresAt,
		currentUser.GetID(), request.Scopes)
	if err != nil {
		return nil, err
	}

	token, err = c.accessTokenMgr.CreateAccessToken(ctx, token)
	if err != nil {
		return nil, err
	}

	resp := &CreatePersonalAccessTokenResponse{
		PersonalAccessToken: PersonalAccessToken{
			CreatePersonalAccessTokenRequest: CreatePersonalAccessTokenRequest{
				Name:      token.Name,
				Scopes:    request.Scopes,
				ExpiresAt: parseExpiredAt(token.CreatedAt, token.ExpiresIn),
			},
			CreatedAt: token.CreatedAt,
			CreatedBy: &usermodels.UserBasic{
				ID:    currentUser.GetID(),
				Name:  currentUser.GetName(),
				Email: currentUser.GetEmail(),
			},
			ID: token.ID,
		},
		Token: token.Code,
	}

	return resp, nil
}

func (c *controller) ListPersonalAccessTokens(ctx context.Context,
	query *q.Query) (accessTokens []PersonalAccessToken, total int, err error) {
	var (
		tokens []*models.AccessToken
	)

	tokens, total, err = c.accessTokenMgr.ListPersonalAccessTokens(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	for _, token := range tokens {
		creator, err := c.userMgr.GetUserByID(ctx, token.CreatedBy)
		if err != nil {
			return nil, 0, err
		}
		accessTokens = append(accessTokens, PersonalAccessToken{
			CreatePersonalAccessTokenRequest: CreatePersonalAccessTokenRequest{
				Name:      token.Name,
				Scopes:    strings.Split(token.Scope, " "),
				ExpiresAt: parseExpiredAt(token.CreatedAt, token.ExpiresIn),
			},
			CreatedAt: token.CreatedAt,
			CreatedBy: &usermodels.UserBasic{
				ID:    creator.ID,
				Name:  creator.Name,
				Email: creator.Email,
			},
			ID: token.ID,
		})
	}

	return accessTokens, total, err
}

func (c *controller) ListResourceAccessTokens(ctx context.Context, resourceType string,
	resourceID uint, query *q.Query) (accessTokens []ResourceAccessToken, total int, err error) {
	var (
		tokens []*models.AccessToken
	)
	tokens, total, err = c.accessTokenMgr.ListAccessTokensByResource(ctx, resourceType, resourceID, query)
	if err != nil {
		return nil, 0, err
	}

	for _, token := range tokens {
		creator, err := c.userMgr.GetUserByID(ctx, token.CreatedBy)
		if err != nil {
			return nil, 0, err
		}
		accessTokens = append(accessTokens, ResourceAccessToken{
			CreateResourceAccessTokenRequest: CreateResourceAccessTokenRequest{
				CreatePersonalAccessTokenRequest: CreatePersonalAccessTokenRequest{
					Name:      token.Name,
					Scopes:    strings.Split(token.Scope, " "),
					ExpiresAt: parseExpiredAt(token.CreatedAt, token.ExpiresIn),
				},
				Role: token.Role,
			},
			CreatedAt: token.CreatedAt,
			CreatedBy: &usermodels.UserBasic{
				ID:    creator.ID,
				Name:  creator.Name,
				Email: creator.Email,
			},
			ID: token.ID,
		})
	}

	return accessTokens, total, err
}

func (c *controller) RevokePersonalAccessToken(ctx context.Context, id uint) error {
	token, err := c.accessTokenMgr.GetAccessToken(ctx, id)
	if err != nil {
		return err
	}

	checkPermission := func() error {
		currentUser, err := common.UserFromContext(ctx)
		if err != nil {
			return err
		}
		if currentUser.IsAdmin() {
			return nil
		}
		if currentUser.GetID() != token.UserID {
			return perror.Wrap(herror.ErrForbidden, "you could not revoke access tokens created by others")
		}
		return nil
	}

	// 1. check if current user can revoke this access token
	if err := checkPermission(); err != nil {
		return err
	}

	// 2. delete token
	return c.accessTokenMgr.DeleteAccessToken(ctx, id)
}

func (c *controller) RevokeResourceAccessToken(ctx context.Context, id uint) error {
	token, err := c.accessTokenMgr.GetAccessToken(ctx, id)
	if err != nil {
		return err
	}

	user, err := c.userMgr.GetUserByID(ctx, token.UserID)
	if err != nil {
		return err
	}
	if user.UserType != usermodels.UserTypeRobot {
		return perror.Wrap(herror.ErrParamInvalid, "this is not a resource token")
	}

	checkPermission := func() error {
		// list members of the robot user
		members, err := c.memberMgr.ListMembersByUserID(ctx, user.ID)
		if err != nil {
			return err
		}
		// check if current user can delete all the members
		for _, member := range members {
			if err := c.memberSvc.CheckIfPermissionEqualOrHigher(ctx, member.Role, string(member.ResourceType),
				member.ResourceID); err != nil {
				return err
			}
		}
		return nil
	}

	cleanRelatedResources := func() error {
		// 2. delete member
		err = c.memberMgr.DeleteMemberByMemberNameID(ctx, token.UserID)
		if err != nil {
			return err
		}
		// 3. delete user
		return c.userMgr.DeleteUser(ctx, token.UserID)
	}

	// 1. check if current user can revoke this access token
	if err := checkPermission(); err != nil {
		return err
	}

	// 2. delete token
	if err := c.accessTokenMgr.DeleteAccessToken(ctx, id); err != nil {
		return err
	}

	// 3. delete related resources
	return cleanRelatedResources()
}

func generateRobot(token, resourceType string, resourceID uint) *usermodels.User {
	fullName := fmt.Sprintf("%s_%d_robot_%s", resourceType, resourceID, uuid.New())
	name := token
	email := fmt.Sprintf("%s%s", fullName, RobotEmailSuffix)
	return &usermodels.User{
		Name:     name,
		FullName: fullName,
		Email:    email,
		UserType: usermodels.UserTypeRobot,
	}
}

func generateCode(userID uint) string {
	buf := bytes.NewBufferString(time.Now().String())
	buf.WriteString(strconv.Itoa(int(userID)))
	code := base64.URLEncoding.EncodeToString([]byte(uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
	code = generate.AccessTokenPrefix + strings.ToUpper(strings.TrimRight(code, "="))
	return code
}

func generateToken(name, expiresAtStr string, userID uint, scopes []string) (*oauthmodels.Token, error) {
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
		Name:      name,
		Code:      generateCode(userID),
		Scope:     strings.Join(scopes, " "),
		CreatedAt: createdAt,
		ExpiresIn: expiredIn,
		UserID:    userID,
	}, nil
}

func parseExpiredAt(startTime time.Time, expiresIn time.Duration) string {
	expiredAt := NeverExpire
	if expiresIn > 0 {
		expiredAt = startTime.Add(expiresIn).Format(ExpiresAtFormat)
	}
	return expiredAt
}
