package user

import (
	"context"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/user/manager"
	linkmanager "g.hz.netease.com/horizon/pkg/userlink/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	// GetUserByEmail get user by email
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, query *q.Query) (int64, []*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
	UpdateByID(c context.Context, id uint, u *UpdateUserRequest) (*User, error)
	ListUserLinks(ctx context.Context, uid uint) ([]*Link, error)
	DeleteLinksByID(c context.Context, id uint) error
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
	const op = "user controller: get user by id"
	defer wlog.Start(ctx, op).StopPrint()

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
	if currentUser.GetID() != id || !currentUser.IsAdmin() {
		return nil, perror.Wrapf(herrors.ErrNoPrivilege,
			"you can not update user %d", id)
	}
	userInDB, err := c.userMgr.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if u.IsBanned != nil {
		userInDB.Banned = *u.IsBanned
	} else if u.IsAdmin != nil {
		if !userInDB.Admin {
			return nil, perror.Wrap(herrors.ErrNoPrivilege,
				"you have no privilege to update user's admin permission")
		}
		userInDB.Admin = *u.IsAdmin
	} else {
		return ofUser(userInDB), nil
	}

	updatedUserInDB, err := c.userMgr.UpdateByID(ctx, id, userInDB)
	if err != nil {
		return nil, err
	}

	return ofUser(updatedUserInDB), nil
}

func (c *controller) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const op = "user controller: get user by email"
	defer wlog.Start(ctx, op).StopPrint()
	user, err := c.userMgr.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return ofUser(user), nil
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

	if !user.IsAdmin() && link.UserID != user.GetID() {
		return perror.Wrap(herrors.ErrNoPrivilege, "could not delete link\n"+
			"should be admin or link owner")
	}

	return c.linksMgr.DeleteByID(ctx, id)
}
