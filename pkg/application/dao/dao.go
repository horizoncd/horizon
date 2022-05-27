package dao

import (
	"context"
	"fmt"
	"time"

	common2 "g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/common"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	"gorm.io/gorm"
)

type DAO interface {
	GetByID(ctx context.Context, id uint) (*models.Application, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*models.Application, error)
	GetByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Application, error)
	GetByName(ctx context.Context, name string) (*models.Application, error)
	GetByNamesUnderGroup(ctx context.Context, groupID uint, names []string) ([]*models.Application, error)
	// GetByNameFuzzily get applications that fuzzily matching the given name
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error)
	// GetByNameFuzzilyByPagination get applications that fuzzily matching the given name
	GetByNameFuzzilyByPagination(ctx context.Context, name string, query q.Query) (int, []*models.Application, error)
	// CountByGroupID get the count of the records matching the given groupID
	CountByGroupID(ctx context.Context, groupID uint) (int64, error)
	Create(ctx context.Context, application *models.Application,
		extraMembers map[*usermodels.User]string) (*models.Application, error)
	UpdateByID(ctx context.Context, id uint, application *models.Application) (*models.Application, error)
	DeleteByID(ctx context.Context, id uint) error
	TransferByID(ctx context.Context, id uint, groupID uint) error
	ListUserAuthorizedByNameFuzzily(ctx context.Context,
		name string, groupIDs []uint, userInfo uint, query *q.Query) (int, []*models.Application, error)
}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) CountByGroupID(ctx context.Context, groupID uint) (int64, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	var count int64
	result := db.Raw(common.ApplicationCountByGroupID, groupID).Scan(&count)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return 0, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return count, result.Error
}

func (d *dao) GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var applications []*models.Application
	result := db.Raw(common.ApplicationQueryByFuzzily, fmt.Sprintf("%%%s%%", name)).Scan(&applications)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return applications, result.Error
}

func (d *dao) GetByNameFuzzilyByPagination(ctx context.Context, name string, query q.Query) (int,
	[]*models.Application, error) {
	var (
		applications []*models.Application
		total        int
	)

	db, err := orm.FromContext(ctx)
	if err != nil {
		return total, nil, err
	}

	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize

	result := db.Raw(common.ApplicationQueryByFuzzilyAndPagination, fmt.Sprintf("%%%s%%", name), limit, offset).
		Scan(&applications)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return total, applications, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return total, applications, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	result = db.Raw(common.ApplicationQueryByFuzzilyCount, fmt.Sprintf("%%%s%%", name)).Scan(&total)

	return total, applications, result.Error
}

func (d *dao) GetByID(ctx context.Context, id uint) (*models.Application, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var application models.Application
	result := db.Raw(common.ApplicationQueryByID, id).First(&application)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return &application, result.Error
}

func (d *dao) GetByIDs(ctx context.Context, ids []uint) ([]*models.Application, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var applications []*models.Application
	result := db.Raw(common.ApplicationQueryByIDs, ids).Scan(&applications)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return applications, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return applications, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return applications, nil
}

func (d *dao) GetByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Application, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var applications []*models.Application
	result := db.Raw(common.ApplicationQueryByGroupIDs, groupIDs).Scan(&applications)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return applications, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return applications, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return applications, result.Error
}

func (d *dao) GetByName(ctx context.Context, name string) (*models.Application, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var application models.Application
	result := db.Raw(common.ApplicationQueryByName, name).First(&application)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &application, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return &application, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return &application, result.Error
}

func (d *dao) GetByNamesUnderGroup(ctx context.Context, groupID uint, names []string) ([]*models.Application, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var applications []*models.Application
	result := db.Raw(common.ApplicationQueryByNamesUnderGroup, groupID, names).Scan(&applications)

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
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
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
			return herrors.NewErrInsertFailed(herrors.ApplicationInDB, err.Error())
		}
		if result.RowsAffected == 0 {
			return herrors.NewErrInsertFailed(herrors.ApplicationInDB, "create member error")
		}
		return nil
	})
	return application, err
}

func (d *dao) UpdateByID(ctx context.Context, id uint, application *models.Application) (*models.Application, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var applicationInDB models.Application
	if err := db.Transaction(func(tx *gorm.DB) error {
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
		applicationInDB.GitBranch = application.GitBranch
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
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	currentUser, err := common2.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Exec(common.ApplicationDeleteByID, time.Now().Unix(), currentUser.GetID(), id)
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
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	currentUser, err := common2.FromContext(ctx)
	if err != nil {
		return err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
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

func (d *dao) ListUserAuthorizedByNameFuzzily(ctx context.Context,
	name string, groupIDs []uint, userInfo uint, query *q.Query) (int, []*models.Application, error) {
	var (
		applications []*models.Application
		total        int
	)

	db, err := orm.FromContext(ctx)
	if err != nil {
		return total, nil, err
	}

	offset := (query.PageNumber - 1) * query.PageSize
	limit := query.PageSize
	like := "%" + name + "%"

	result := db.Raw(common.ApplicationQueryByUserAndNameFuzzily, userInfo, like, groupIDs, like, limit, offset).
		Scan(&applications)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, nil, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return 0, nil, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	result = db.Raw(common.ApplicationCountByUserAndNameFuzzily, userInfo, like, groupIDs, like).Scan(&total)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, nil, herrors.NewErrNotFound(herrors.ApplicationInDB, result.Error.Error())
		}
		return 0, nil, herrors.NewErrGetFailed(herrors.ApplicationInDB, result.Error.Error())
	}

	return total, applications, nil
}
