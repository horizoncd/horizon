package dao

import (
	"context"
	"net/http"
	"time"

	common2 "g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/pkg/common"
	memberctx "g.hz.netease.com/horizon/pkg/member/context"
	"g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, member *models.Member) (*models.Member, error)
	Get(ctx context.Context, resourceType models.ResourceType, resourceID uint,
		memberType models.MemberType, memberInfo uint) (*models.Member, error)
	GetByID(ctx context.Context, memberID uint) (*models.Member, error)
	Delete(ctx context.Context, memberID uint) error
	UpdateByID(ctx context.Context, memberID uint, role string) (*models.Member, error)
	ListDirectMember(ctx context.Context, resourceType models.ResourceType,
		resourceID uint) ([]models.Member, error)
	ListDirectMemberOnCondition(ctx context.Context, resourceType models.ResourceType,
		resourceID uint) ([]models.Member, error)
	ListResourceOfMemberInfo(ctx context.Context, resourceType models.ResourceType, memberInfo uint) ([]uint, error)
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

type dao struct{ db *gorm.DB }

func (d *dao) Create(ctx context.Context, member *models.Member) (*models.Member, error) {
	result := d.db.WithContext(ctx).Create(member)
	return member, result.Error
}

func (d *dao) Get(ctx context.Context, resourceType models.ResourceType, resourceID uint,
	memberType models.MemberType, memberInfo uint) (*models.Member, error) {
	var member models.Member
	result := d.db.WithContext(ctx).Raw(common.MemberSingleQuery, resourceType, resourceID,
		memberType, memberInfo).Scan(&member)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &member, nil
}

func (d *dao) GetByID(ctx context.Context, memberID uint) (*models.Member, error) {
	var member models.Member
	result := d.db.WithContext(ctx).Raw(common.MemberQueryByID, memberID).Scan(&member)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &member, nil
}

func (d *dao) UpdateByID(ctx context.Context, id uint, role string) (*models.Member, error) {
	const op = "member dao: update by ID"

	currentUser, err := common2.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var memberInDB models.Member
	if err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. get member in db first
		result := tx.Raw(common.MemberQueryByID, id).Scan(&memberInDB)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.E(op, http.StatusNotFound)
		}

		// 2. update value
		memberInDB.Role = role
		memberInDB.GrantedBy = currentUser.GetID()

		// 3. save member after updated
		tx.Save(&memberInDB)
		return nil
	}); err != nil {
		return nil, err
	}

	return &memberInDB, nil
}

func (d *dao) Delete(ctx context.Context, memberID uint) error {
	result := d.db.WithContext(ctx).Exec(common.MemberSingleDelete, time.Now().Unix(), memberID)
	return result.Error
}

func (d *dao) ListDirectMember(ctx context.Context, resourceType models.ResourceType,
	resourceID uint) ([]models.Member, error) {
	var members []models.Member
	result := d.db.WithContext(ctx).Raw(common.MemberSelectAll, resourceType, resourceID).Scan(&members)
	if result.Error != nil {
		return nil, result.Error
	}
	return members, nil
}

func (d *dao) ListDirectMemberOnCondition(ctx context.Context, resourceType models.ResourceType,
	resourceID uint) ([]models.Member, error) {
	var members []models.Member
	if emails, ok := ctx.Value(memberctx.ContextEmails).([]string); ok {
		result := d.db.WithContext(ctx).Raw(common.MemberSelectByUserEmails, resourceType, resourceID, emails).Scan(&members)
		if result.Error != nil {
			return nil, result.Error
		}
	}
	return members, nil
}

func (d *dao) ListResourceOfMemberInfo(ctx context.Context,
	resourceType models.ResourceType, memberInfo uint) ([]uint, error) {
	var resources []uint
	result := d.db.WithContext(ctx).Raw(common.MemberListResource, resourceType, memberInfo).Scan(&resources)
	if result.Error != nil {
		return nil, result.Error
	}
	return resources, nil
}
