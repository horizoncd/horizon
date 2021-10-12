package dao

import (
	"context"
	"net/http"

	user2 "g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"gorm.io/gorm"
)

// _defaultQuery default query params
var _defaultQuery = &q.Query{
	// PageNumber start with 1
	PageNumber: 1,
	PageSize:   20,
}

type DAO interface {
	Create(ctx context.Context, member *models.Member) (*models.Member, error)
	Get(ctx context.Context, resourceType models.ResourceType, resourceID uint,
		memberType models.MemberType, memberInfo string) (*models.Member, error)
	Delete(ctx context.Context, memberID uint) error
	UpdateByID(ctx context.Context, id uint, member *models.Member) (*models.Member, error)
	ListDirectMember(ctx context.Context, resourceType models.ResourceType,
		resourceID uint) ([]models.Member, error)
}

func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(ctx context.Context, member *models.Member) (*models.Member, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(member)
	return member, result.Error
}

func (d *dao) Get(ctx context.Context, resourceType models.ResourceType, resourceID uint,
	memberType models.MemberType, memberInfo string) (*models.Member, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var member models.Member
	result := db.Raw(common.MemberSingleQuery, resourceType, resourceID, memberType, memberInfo).Scan(&member)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &member, nil
}

func (d *dao) UpdateByID(ctx context.Context, id uint, member *models.Member) (*models.Member, error) {
	const op = "member dao: update by ID"
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := user2.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var memberInDB models.Member
	if err := db.Transaction(func(tx *gorm.DB) error {
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

func (d *dao) Delete(ctx context.Context, memberID uint) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Exec(common.MemberSingleDelete, memberID)
	return result.Error
}

func (d *dao) ListDirectMember(ctx context.Context, resourceType models.ResourceType,
	resourceID uint) ([]models.Member, error) {

	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var members []models.Member
	result := db.Raw(common.MemberSelectAll, resourceType, resourceID).Scan(&members)
	if result.Error != nil {
		return nil, result.Error
	}
	return members, nil
}
