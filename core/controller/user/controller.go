package user

import (
	"context"

	"g.hz.netease.com/horizon/lib/q"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	// SearchUser search for user
	SearchUser(ctx context.Context, filter string, query *q.Query) (int, []*SearchUserResponse, error)
	// GetUserByEmail get user by email
	GetUserByEmail(ctx context.Context, email string) (userauth.User, error)
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

func (c *controller) SearchUser(ctx context.Context,
	filter string, query *q.Query) (_ int, _ []*SearchUserResponse, err error) {
	const op = "user controller: search user"
	defer wlog.Start(ctx, op).StopPrint()

	count, users, err := c.userMgr.SearchUser(ctx, filter, query)
	if err != nil {
		return 0, nil, errors.E(op, err)
	}
	return count, ofUsers(users), nil
}

func (c *controller) GetUserByEmail(ctx context.Context, email string) (userauth.User, error) {
	const op = "user controller: get user by email"
	defer wlog.Start(ctx, op).StopPrint()
	user, err := c.userMgr.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return ofUser(user), nil
}
