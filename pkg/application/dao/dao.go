package dao

import (
	"context"
	"fmt"
	"time"

	corecommon "github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/application/models"
	"github.com/horizoncd/horizon/pkg/common"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/rbac/role"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	"gorm.io/gorm"
)

const (
	KeyTemplate        = "template"
	KeyTemplateRelease = "templateRelease"
)

type DAO interface {
	GetByID(ctx context.Context, id uint, includeSoftDelete bool) (*models.Application, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*models.Application, error)
	GetByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Application, error)
	GetByName(ctx context.Context, name string) (*models.Application, error)
	GetByNamesUnderGroup(ctx context.Context, groupID uint, names []string) ([]*models.Application, error)
	// GetByNameFuzzily get applications that fuzzily matching the given name
	GetByNameFuzzily(ctx context.Context, name string, includeSoftDelete bool) ([]*models.Application, error)
	// CountByGroupID get the count of the records matching the given groupID
	CountByGroupID(ctx context.Context, groupID uint) (int64, error)
	Create(ctx context.Context, application *models.Application,
		extraMembers map[*usermodels.User]string) (*models.Application, error)
	UpdateByID(ctx context.Context, id uint, application *models.Application) (*models.Application, error)
	DeleteByID(ctx context.Context, id uint) error
	TransferByID(ctx context.Context, id uint, groupID uint) error
	List(ctx context.Context, groupIDs []uint, query *q.Query) (int, []*models.Application, error)
}

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

type dao struct{ db *gorm.DB }

func (d *dao) CountByGroupID(ctx context.Context, groupID uint) (int64, error) {
	var count int64
	result := d.db.WithContext(ctx).Raw(common.ApplicationCountByGroupID, groupID).Scan(&count)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return 0, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return count, result.Error
}

func (d *dao) GetByNameFuzzily(ctx context.Context, name string,
	includeSoftDelete bool) ([]*models.Application, error) {
	var applications []*models.Application

	statement := d.db.Unscoped().WithContext(ctx).Where("name like ?", fmt.Sprintf("%%%s%%", name))
	if !includeSoftDelete {
		statement.Where("deleted_ts = 0")
	}
	result := statement.Find(&applications)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return applications, result.Error
}

func (d *dao) GetByID(ctx context.Context, id uint, includeSoftDelete bool) (*models.Application, error) {
	var application models.Application
	statement := d.db.Unscoped().WithContext(ctx).Where("id = ?", id)
	if !includeSoftDelete {
		statement.Where("deleted_ts = 0")
	}
	result := statement.First(&application)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return &application, result.Error
}

func (d *dao) GetByIDs(ctx context.Context, ids []uint) ([]*models.Application, error) {
	var applications []*models.Application
	result := d.db.WithContext(ctx).Raw(common.ApplicationQueryByIDs, ids).Scan(&applications)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return applications, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return applications, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return applications, nil
}

func (d *dao) GetByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Application, error) {
	var applications []*models.Application
	result := d.db.WithContext(ctx).Raw(common.ApplicationQueryByGroupIDs, groupIDs).Scan(&applications)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return applications, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return applications, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return applications, result.Error
}

func (d *dao) GetByName(ctx context.Context, name string) (*models.Application, error) {
	var application models.Application
	result := d.db.WithContext(ctx).Raw(common.ApplicationQueryByName, name).First(&application)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &application, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return &application, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return &application, result.Error
}

func (d *dao) GetByNamesUnderGroup(ctx context.Context, groupID uint, names []string) ([]*models.Application, error) {
	var applications []*models.Application
	result := d.db.WithContext(ctx).Raw(common.ApplicationQueryByNamesUnderGroup, groupID, names).Scan(&applications)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return applications, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return applications, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return applications, result.Error
}

func (d *dao) Create(ctx context.Context, application *models.Application,
	extraMembers map[*usermodels.User]string) (*models.Application, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// TODO: check the group exist

		if err := tx.Create(application).Error; err != nil {
			return herrors.NewErrInsertFailed(herrors.ApplicationInDB, err.Error())
		}
		// insert records to member table
		members := make([]*membermodels.Member, 0)

		// the owner who created this application
		members = append(members, &membermodels.Member{
			ResourceType: membermodels.TypeApplication,
			ResourceID:   application.ID,
			Role:         role.Owner,
			MemberType:   membermodels.MemberUser,
			MemberNameID: application.CreatedBy,
			GrantedBy:    application.UpdatedBy,
		})

		// the extra members
		for extraMember, roleOfMember := range extraMembers {
			if extraMember.ID == application.CreatedBy {
				continue
			}
			members = append(members, &membermodels.Member{
				ResourceType: membermodels.TypeApplication,
				ResourceID:   application.ID,
				Role:         roleOfMember,
				MemberType:   membermodels.MemberUser,
				MemberNameID: extraMember.ID,
				GrantedBy:    application.CreatedBy,
			})
		}

		result := tx.Create(members)
		if result.Error != nil {
			return herrors.NewErrInsertFailed(herrors.ApplicationInDB, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return herrors.NewErrInsertFailed(herrors.ApplicationInDB, "create member error")
		}
		return nil
	})
	return application, err
}

func (d *dao) UpdateByID(ctx context.Context, id uint, application *models.Application) (*models.Application, error) {
	var applicationInDB models.Application
	if err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. get application in db first
		result := tx.Raw(common.ApplicationQueryByID, id).Scan(&applicationInDB)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
			}
			return herrors.NewErrUpdateFailed(herrors.ApplicationInDB, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return herrors.NewErrNotFound(herrors.ApplicationInDB, "rows affected = 0")
		}
		// 2. update value
		applicationInDB.Description = application.Description
		applicationInDB.Priority = application.Priority
		applicationInDB.GitURL = application.GitURL
		applicationInDB.GitSubfolder = application.GitSubfolder
		applicationInDB.GitRefType = application.GitRefType
		applicationInDB.GitRef = application.GitRef
		applicationInDB.Template = application.Template
		applicationInDB.TemplateRelease = application.TemplateRelease
		// 3. save application after updated
		tx.Save(&applicationInDB)

		return nil
	}); err != nil {
		return nil, err
	}
	return &applicationInDB, nil
}

func (d *dao) DeleteByID(ctx context.Context, id uint) error {
	currentUser, err := corecommon.UserFromContext(ctx)
	if err != nil {
		return err
	}

	result := d.db.WithContext(ctx).Exec(common.ApplicationDeleteByID, time.Now().Unix(), currentUser.GetID(), id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return herrors.NewErrDeleteFailed(herrors.ApplicationInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return herrors.NewErrNotFound(herrors.ApplicationInDB, "application not found")
	}
	return nil
}

func (d *dao) TransferByID(ctx context.Context, id uint, groupID uint) error {
	currentUser, err := corecommon.UserFromContext(ctx)
	if err != nil {
		return err
	}
	err = d.db.Transaction(func(tx *gorm.DB) error {
		var group groupmodels.Group
		result := tx.Raw(common.GroupQueryByID, groupID).Scan(&group)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return herrors.NewErrNotFound(herrors.GroupInDB, "group not found")
		}

		result = tx.Exec(common.ApplicationTransferByID, groupID, currentUser.GetID(), id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return herrors.NewErrNotFound(herrors.ApplicationInDB, "application not found")
		}
		return nil
	})

	return err
}

func (d *dao) List(ctx context.Context, groupIDs []uint,
	query *q.Query) (int, []*models.Application, error) {
	var (
		applications   []*models.Application
		total          int64
		finalStatement *gorm.DB
	)

	genSQL := func() *gorm.DB {
		statement := d.db.WithContext(ctx).Table("tb_application as a").Select("a.*")
		for k, v := range query.Keywords {
			switch k {
			case corecommon.ApplicationQueryName:
				statement = statement.Where("a.name like ?", fmt.Sprintf("%%%v%%", v))
			case corecommon.ApplicationQueryByTemplate:
				statement = statement.Where("a.template = ?", v)
			case corecommon.ApplicationQueryByRelease:
				statement = statement.Where("a.template_release = ?", v)
			}
		}
		statement = statement.Where("a.deleted_ts = 0")
		return statement
	}

	if query != nil {
		// basic filter
		finalStatement = genSQL()
		// add groupID filter
		if len(groupIDs) != 0 {
			finalStatement = finalStatement.Where("group_id in ?", groupIDs)
		}

		// union user direct authorized entity
		if userID, ok := query.Keywords[corecommon.ApplicationQueryByUser]; ok {
			statementUser := genSQL().
				Joins("join tb_member as m on m.resource_id = a.id").
				Where("m.resource_type = ?", corecommon.ResourceApplication).
				Where("m.member_type = '0'").
				Where("m.membername_id = ?", userID).
				Where("m.deleted_ts = 0")
			if len(groupIDs) != 0 {
				finalStatement = d.db.Raw("? union ?", finalStatement, statementUser)
			} else {
				// user does not belong to any group
				finalStatement = statementUser
			}
		}
	}

	res := d.db.Raw("select count(distinct id) from (?) as apps", finalStatement).Scan(&total)

	if res.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.ApplicationInDB, res.Error.Error())
	}

	finalStatement = d.db.Raw("select distinct * from (?) as apps order by updated_at desc limit ? offset ?",
		finalStatement, query.Limit(), query.Offset())
	res = finalStatement.Scan(&applications)
	if res.Error != nil {
		return 0, nil, herrors.NewErrGetFailed(herrors.ApplicationInDB, res.Error.Error())
	}

	return int(total), applications, nil
}
