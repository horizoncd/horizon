package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/user/models"
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
	Create(ctx context.Context, user *models.User) (*models.User, error)
	// GetByOIDCMeta get user by oidcType and email
	GetByOIDCMeta(ctx context.Context, oidcType, email string) (*models.User, error)
	// SearchUser search user with a given filter, search for name/full_name/email.
	SearchUser(ctx context.Context, filter string, query *q.Query) (int, []models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	ListByEmail(ctx context.Context, emails []string) ([]*models.User, error)
	GetByIDs(ctx context.Context, userID []uint) ([]models.User, error)
}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(ctx context.Context, user *models.User) (*models.User, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(user)
	return user, result.Error
}

func (d *dao) GetByIDs(ctx context.Context, userID []uint) ([]models.User, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	var users []models.User
	result := db.Raw(common.UserGetByID, userID).Scan(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return users, nil
}

func (d *dao) GetByOIDCMeta(ctx context.Context, oidcType, email string) (*models.User, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	result := db.Raw(common.UserQueryByOIDC, oidcType, email).Scan(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}

func (d *dao) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	result := db.Raw(common.UserQueryByEmail, email).Scan(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}

func (d *dao) ListByEmail(ctx context.Context, emails []string) ([]*models.User, error) {
	if len(emails) == 0 {
		return nil, nil
	}

	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var users []*models.User
	result := db.Raw(common.UserListByEmail, emails).Scan(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return users, nil
}

func (d *dao) SearchUser(ctx context.Context, filter string, query *q.Query) (int, []models.User, error) {
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
	var users []models.User
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
