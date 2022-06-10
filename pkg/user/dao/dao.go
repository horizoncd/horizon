package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/user/models"
	"gorm.io/gorm"
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
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

type dao struct{ db *gorm.DB }

func (d *dao) Create(ctx context.Context, user *models.User) (*models.User, error) {

	result := d.db.WithContext(ctx).WithContext(ctx).Create(user)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.UserInDB, result.Error.Error())
	}

	return user, result.Error
}

func (d *dao) GetByIDs(ctx context.Context, userID []uint) ([]models.User, error) {

	var users []models.User
	result := d.db.WithContext(ctx).Raw(common.UserGetByID, userID).Scan(&users)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.UserInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return users, nil
}

func (d *dao) GetByOIDCMeta(ctx context.Context, oidcType, email string) (*models.User, error) {

	var user models.User
	result := d.db.WithContext(ctx).Raw(common.UserQueryByOIDC, oidcType, email).Scan(&user)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.UserInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}

func (d *dao) GetByEmail(ctx context.Context, email string) (*models.User, error) {

	var user models.User
	result := d.db.WithContext(ctx).Raw(common.UserQueryByEmail, email).Scan(&user)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.UserInDB, result.Error.Error())
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

	var users []*models.User
	result := d.db.WithContext(ctx).Raw(common.UserListByEmail, emails).Scan(&users)
	if result.Error != nil {
		return nil, herrors.NewErrListFailed(herrors.UserInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return users, nil
}

func (d *dao) SearchUser(ctx context.Context, filter string, query *q.Query) (int, []models.User, error) {

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
	result := d.db.WithContext(ctx).Raw(common.UserSearch, like, like, like, limit, offset).Scan(&users)
	if result.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.UserInDB, result.Error.Error())
	}

	var count int
	result = d.db.WithContext(ctx).Raw(common.UserSearchCount, like, like, like).Scan(&count)
	if result.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.UserInDB, result.Error.Error())
	}

	return count, users, nil
}
