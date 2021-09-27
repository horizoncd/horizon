package user

import (
	"context"

	usermiddle "g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/user"
	"g.hz.netease.com/horizon/util/errors"
	"g.hz.netease.com/horizon/util/wlog"
)

var (
	Ctl = NewController()
)

type Controller interface {
	// GetName returns the current username that uniquely identifies this user.
	GetName(ctx context.Context) (string, error)
	// GetID returns the current user id that uniquely identifies this user.
	GetID(ctx context.Context) (int, error)
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

func (c *controller) GetName(ctx context.Context) (string, error) {
	u, err := usermiddle.FromContext(ctx)
	if err != nil {
		return "", err
	}
	return u.Name, nil
}

func (c *controller) GetID(ctx context.Context) (int, error) {
	u, err := usermiddle.FromContext(ctx)
	if err != nil {
		return -1, err
	}
	return int(u.ID), nil
}

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
