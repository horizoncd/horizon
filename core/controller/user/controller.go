package user

import (
	"context"

	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	// GetUserByEmail get user by email
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, query *q.Query) (int64, []*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
	UpdateByID(c context.Context, id uint, u *UpdateUserRequest) (*User, error)
}

type controller struct {
	userMgr manager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		userMgr: param.UserManager,
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
	userInDB, err := c.userMgr.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if u.IsBanned != nil {
		userInDB.Banned = *u.IsBanned
	} else if u.IsAdmin != nil {
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
