package user

import (
	"context"

	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/user"
	"g.hz.netease.com/horizon/util/errors"
	"g.hz.netease.com/horizon/util/wlog"
)

var (
	Ctl = NewController()
)

type Controller interface {
	// SearchUser search for user
	SearchUser(ctx context.Context, filter string, query *q.Query) (int, []*SearchUserResponse, error)
}

type controller struct {
	userMgr user.Manager
}

func NewController() Controller {
	return &controller{
		userMgr: user.Mgr,
	}
}

var _ Controller = (*controller)(nil)

func (c *controller) SearchUser(ctx context.Context,
	filter string, query *q.Query) (_ int, _ []*SearchUserResponse, err error) {
	const op = "user controller: search user"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	count, users, err := c.userMgr.SearchUser(ctx, filter, query)
	if err != nil {
		return 0, nil, errors.E(op, err)
	}
	return count, ofUsers(users), nil
}
