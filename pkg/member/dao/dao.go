package dao

import (
	"context"
	"net/http"
	"time"

	common2 "github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/common"
	memberctx "github.com/horizoncd/horizon/pkg/context"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/util/errors"
	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, member *models.Member) (*models.Member, error)
	Get(ctx context.Context, resourceType models.ResourceType, resourceID uint,
		memberType models.MemberType, memberInfo uint) (*models.Member, error)
	GetByID(ctx context.Context, memberID uint) (*models.Member, error)
	Delete(ctx context.Context, memberID uint) error
	HardDelete(ctx context.Context, resourceType string, resourceID uint) error
	DeleteByMemberNameID(ctx context.Context, memberNameID uint) error
	UpdateByID(ctx context.Context, memberID uint, role string) (*models.Member, error)
	ListDirectMember(ctx context.Context, resourceType models.ResourceType,
		resourceID uint) ([]models.Member, error)
	ListDirectMemberOnCondition(ctx context.Context, resourceType models.ResourceType,
		resourceID uint) ([]models.Member, error)
	ListResourceOfMemberInfo(ctx context.Context, resourceType models.ResourceType, memberInfo uint) ([]uint, error)
	ListResourceOfMemberInfoByRole(ctx context.Context,
		resourceType models.ResourceType, info uint, role string) ([]uint, error)
	ListMembersByUserID(ctx context.Context, userID uint) ([]models.Member, error)
}

var (
	model = &models.Member{}
)

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

func (d *dao) DeleteByMemberNameID(ctx context.Context, memberNameID uint) error {
	result := d.db.WithContext(ctx).Exec(common.MemberHardDeleteByMemberNameID, memberNameID)
	return result.Error
}

func (d *dao) HardDelete(ctx context.Context, resourceType string, resourceID uint) error {
	result := d.db.WithContext(ctx).Exec(common.MemberHardDeleteByResourceTypeID, resourceType, resourceID)
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
	if emails, ok := ctx.Value(memberctx.MemberEmails).([]string); ok {
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

func (d *dao) ListResourceOfMemberInfoByRole(ctx context.Context,
	resourceType models.ResourceType, info uint, role string) ([]uint, error) {
	members := make([]uint, 0)
	res := d.db.Model(model).WithContext(ctx).Select("resource_id").Where(&models.Member{
		ResourceType: resourceType,
		Role:         role,
		MemberNameID: info,
	}).Find(&members)
	if res.Error != nil {
		return nil, perror.Wrapf(herrors.NewErrGetFailed(herrors.MemberInfoInDB, res.Error.Error()),
			"failed to get members:\n"+
				"resourceType = %v\n member id = %v\n role = %v", resourceType, info, role)
	}
	return members, nil
}

func (d *dao) ListMembersByUserID(ctx context.Context, userID uint) ([]models.Member, error) {
	var members []models.Member
	result := d.db.Model(model).WithContext(ctx).
		Where("membername_id = ?", userID).
		Where("deleted_ts = 0").
		Scan(&members)
	if result.Error != nil {
		return nil, result.Error
	}
	return members, nil
}
