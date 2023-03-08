package accesstoken

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	tokenmanager "github.com/horizoncd/horizon/pkg/token/manager"
	tokenservice "github.com/horizoncd/horizon/pkg/token/service"

	"github.com/horizoncd/horizon/core/common"
	herror "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	accesstokenmanager "github.com/horizoncd/horizon/pkg/accesstoken/manager"
	"github.com/horizoncd/horizon/pkg/accesstoken/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
	membermanager "github.com/horizoncd/horizon/pkg/member"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	memberservice "github.com/horizoncd/horizon/pkg/member/service"
	"github.com/horizoncd/horizon/pkg/param"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
)

type Controller interface {
	CreateResourceAccessToken(ctx context.Context, request CreateResourceAccessTokenRequest,
		resourceType string, resourceID uint) (*CreateResourceAccessTokenResponse, error)
	CreatePersonalAccessToken(ctx context.Context,
		request CreatePersonalAccessTokenRequest) (*CreatePersonalAccessTokenResponse, error)
	ListPersonalAccessTokens(ctx context.Context,
		query *q.Query) (accessTokens []PersonalAccessToken, total int, err error)
	ListResourceAccessTokens(ctx context.Context, resourceType string,
		resourceID uint, query *q.Query) (accessTokens []ResourceAccessToken, total int, err error)
	RevokePersonalAccessToken(ctx context.Context, id uint) error
	RevokeResourceAccessToken(ctx context.Context, id uint) error
}

type controller struct {
	userMgr        usermanager.Manager
	accessTokenMgr accesstokenmanager.Manager
	tokenMgr       tokenmanager.Manager
	tokenSvc       tokenservice.Service
	memberSvc      memberservice.Service
	memberMgr      membermanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		userMgr:        param.UserManager,
		accessTokenMgr: param.AccessTokenManager,
		tokenMgr:       param.TokenManager,
		tokenSvc:       param.TokenSvc,
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

	token, err := c.tokenSvc.CreateAccessToken(ctx, request.Name,
		request.ExpiresAt, userID, request.Scopes)
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

	token, err := c.tokenSvc.CreateAccessToken(ctx, request.Name, request.ExpiresAt,
		currentUser.GetID(), request.Scopes)
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
	token, err := c.tokenMgr.LoadTokenByID(ctx, id)
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
	return c.tokenMgr.RevokeTokenByID(ctx, id)
}

func (c *controller) RevokeResourceAccessToken(ctx context.Context, id uint) error {
	token, err := c.tokenMgr.LoadTokenByID(ctx, id)
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
			if err := c.memberSvc.RequirePermissionEqualOrHigher(ctx, member.Role, string(member.ResourceType),
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
	if err := c.tokenMgr.RevokeTokenByID(ctx, id); err != nil {
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

func parseExpiredAt(startTime time.Time, expiresIn time.Duration) string {
	expiredAt := NeverExpire
	if expiresIn > 0 {
		expiredAt = startTime.Add(expiresIn).Format(ExpiresAtFormat)
	}
	return expiredAt
}
