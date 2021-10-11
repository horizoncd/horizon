package user

import (
	"context"

	"g.hz.netease.com/horizon/pkg/dao/common"
	"g.hz.netease.com/horizon/pkg/lib/orm"
	"g.hz.netease.com/horizon/pkg/lib/q"
)

// _defaultQuery default query params
var _defaultQuery = &q.Query{
	// PageNumber start with 1
	PageNumber: 1,
	// PageSize default pageSize is 20
	PageSize: 20,
}

type DAO interface {
	// Create user
	Create(ctx context.Context, user *User) (*User, error)
	// GetByOIDCMeta get user by oidcID and oidcType
	GetByOIDCMeta(ctx context.Context, oidcID, oidcType string) (*User, error)
	// SearchUser search user with a given filter, search for name/full_name/email.
	SearchUser(ctx context.Context, filter string, query *q.Query) (int, []User, error)
}

// newDAO returns an instance of the default DAO
func newDAO() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(ctx context.Context, user *User) (*User, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(user)
	return user, result.Error
}

func (d *dao) GetByOIDCMeta(ctx context.Context, oidcID, oidcType string) (*User, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var user User
	result := db.Raw(common.UserQueryByOIDC, oidcID, oidcType).Scan(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}

func (d *dao) SearchUser(ctx context.Context, filter string, query *q.Query) (int, []User, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, nil, err
	}

	if query == nil {
		query = _defaultQuery
	}

	if query.PageNumber < 1 {
		query.PageNumber = _defaultQuery.PageNumber
	}

	if query.PageSize == 0 {
		query.PageSize = _defaultQuery.PageSize
	}

	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize

	like := "%" + filter + "%"
	var users []User
	result := db.Raw(common.UserSearch, like, like, like, limit, offset).Scan(&users)
	if result.Error != nil {
		return 0, nil, result.Error
	}

	var count int
	result = db.Raw(common.UserSearchCount, like, like, like).Scan(&count)
	if result.Error != nil {
		return 0, nil, result.Error
	}

	return count, users, nil
}
