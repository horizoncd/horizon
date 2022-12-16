package user

import (
	"context"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/user/manager"
	"github.com/horizoncd/horizon/pkg/user/models"
	linkmanager "github.com/horizoncd/horizon/pkg/userlink/manager"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

type Controller interface {
	List(ctx context.Context, query *q.Query) (int64, []*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
	UpdateByID(c context.Context, id uint, u *UpdateUserRequest) (*User, error)
	ListUserLinks(ctx context.Context, uid uint) ([]*Link, error)
	DeleteLinksByID(c context.Context, id uint) error
	// LoginWithPasswd checks inputed email & password
	LoginWithPasswd(ctx context.Context, request *LoginRequest) (*models.User, error)
}

type controller struct {
	userMgr  manager.Manager
	linksMgr linkmanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		userMgr:  param.UserManager,
		linksMgr: param.UserLinksManager,
	}
}

var _ Controller = (*controller)(nil)

func (c *controller) List(ctx context.Context, query *q.Query) (int64, []*User, error) {
	const op = "user controller: list user"
	defer wlog.Start(ctx, op).StopPrint()

	total, users, err := c.userMgr.List(ctx, query)
	if err != nil {
		return 0, nil, err
	}

	return total, ofUsers(users), nil
}

func (c *controller) GetByID(ctx context.Context, id uint) (*User, error) {
	user, err := c.userMgr.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ofUser(user), nil
}

func (c *controller) UpdateByID(ctx context.Context,
	id uint, u *UpdateUserRequest) (*User, error) {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if !currentUser.IsAdmin() {
		return nil, perror.Wrap(herrors.ErrForbidden,
			"you have no privilege")
	}
	userInDB, err := c.userMgr.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if u.IsAdmin == nil && u.IsBanned == nil {
		return ofUser(userInDB), nil
	}

	if u.IsBanned != nil {
		userInDB.Banned = *u.IsBanned
	}
	if u.IsAdmin != nil {
		userInDB.Admin = *u.IsAdmin
	}

	updatedUserInDB, err := c.userMgr.UpdateByID(ctx, id, userInDB)
	if err != nil {
		return nil, err
	}

	return ofUser(updatedUserInDB), nil
}

func (c *controller) ListUserLinks(ctx context.Context, uid uint) ([]*Link, error) {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin() && uid != user.GetID() {
		return nil, perror.Wrap(herrors.ErrNoPrivilege, "could not list link\n"+
			"should be admin or link owner")
	}
	links, err := c.linksMgr.ListByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return ofUserLinks(links), nil
}

func (c *controller) DeleteLinksByID(ctx context.Context, id uint) error {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}
	link, err := c.linksMgr.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !link.Deletable {
		return perror.Wrapf(
			herrors.ErrForbidden,
			"links with id = %d can not be deleted", id)
	}

	if !user.IsAdmin() && link.UserID != user.GetID() {
		return perror.Wrap(herrors.ErrNoPrivilege, "could not delete link\n"+
			"should be admin or link owner")
	}

	return c.linksMgr.DeleteByID(ctx, id)
}

func (c *controller) LoginWithPasswd(ctx context.Context, request *LoginRequest) (*models.User, error) {
	users, err := c.userMgr.ListByEmail(ctx, []string{request.Email})
	if err != nil {
		return nil, nil
	}
	if len(users) < 1 {
		return nil, perror.Wrapf(herrors.ErrForbidden,
			"there's no account with email = %v found", request.Email)
	}
	if len(users) > 1 {
		return nil, perror.Wrapf(herrors.ErrDuplicatedKey,
			"there's more than one account with email = %v", request.Email)
	}
	user := users[0]
	if user.Password != request.Password {
		return nil, nil
	}
	return user, nil
}
