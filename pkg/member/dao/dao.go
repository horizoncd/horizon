package dao

import (
	"context"
	"net/http"

	"g.hz.netease.com/horizon/common"
	user2 "g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/util/errors"
	"gorm.io/gorm"
)


// _defaultQuery default query params
var _defaultQuery = &q.Query{
	// PageNumber start with 1
	PageNumber: 1,
	PageSize: 20,
}

type DAO interface {
	Create(ctx context.Context, member *models.Member) (*models.Member, error)
	GetByUserName(ctx context.Context, userName string) (*models.Member, error)
	UpdateByID(ctx context.Context, id uint16, member *models.Member) (*models.Member, error)
	ListMember(ctx context.Context, query *q.Query) (int, []models.Member, error)
}

func New() DAO {
	return &dao{}
}

type dao struct {}

func (d *dao) Create(ctx context.Context, member *models.Member) (*models.Member, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(member)
	return member, result.Error
}

func (d *dao) GetByUserName(ctx context.Context, userName string) (*models.Member, error) {
	return nil, nil
}

func (d *dao) UpdateByID(ctx context.Context, id uint16,  member *models.Member) (*models.Member, error) {
	const op = "member dao: update by ID"
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := user2.FromContext(ctx)
	if user != nil {
		return nil, err
	}

	var  memberInDB models.Member
	if err := db.Transaction(func(tx *gorm.DB) error  {
		// 1. get member in db first
		result := tx.Raw(common.MemberUpdate, id).Scan(&memberInDB)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.E(op, http.StatusNotFound)
		}

		// 2. update value
		memberInDB.Role = member.Role
		memberInDB.GrantBy = user.GetName()

		// 3. save member after updated
		tx.Save(&memberInDB)
		return nil
	}); err != nil {
		return nil, err
	}

	return &memberInDB, nil
}

func (d *dao) ListMember(ctx context.Context, query *q.Query) (int, []models.Member, error) {
	return 0, nil, nil
}