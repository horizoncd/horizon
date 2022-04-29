package callbacks

import (
	"context"

	"g.hz.netease.com/horizon/core/middleware/user"
	"gorm.io/gorm"
)

const (
	_createdBy = "created_by"
	_updatedBy = "updated_by"
)

// AddCreatedByUpdatedByForCreateCallback will set `created_by` and `updated_by` when creating records if fields exist
func AddCreatedByUpdatedByForCreateCallback(ctx context.Context, db *gorm.DB) {
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		db.Error = err
		return
	}

	field := db.Statement.Schema.LookUpField(_createdBy)
	if field != nil {
		db.Statement.SetColumn(_createdBy, currentUser.GetID(), true)
	}
	field = db.Statement.Schema.LookUpField(_updatedBy)
	if field != nil {
		db.Statement.SetColumn(_updatedBy, currentUser.GetID(), true)
	}
}

// AddUpdatedByForUpdateDeleteCallback will set `updated_by` when updating or deleting records if fields exist
func AddUpdatedByForUpdateDeleteCallback(ctx context.Context, db *gorm.DB) {
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		db.Error = err
		return
	}

	field := db.Statement.Schema.LookUpField(_updatedBy)
	if field != nil {
		db.Statement.SetColumn(_updatedBy, currentUser.GetID())
	}
}

func RegisterCustomCallbacks(ctx context.Context, db *gorm.DB) {
	_ = db.Callback().Create().Before("gorm:create").After("gorm:before_create").
		Register("add_created", func(d *gorm.DB) {
			AddCreatedByUpdatedByForCreateCallback(ctx, d)
		})

	_ = db.Callback().Update().Before("gorm:update").After("gorm:before_update").
		Register("add_updated_by", func(d *gorm.DB) {
			AddUpdatedByForUpdateDeleteCallback(ctx, d)
		})

	_ = db.Callback().Delete().Before("gorm:delete").After("gorm:before_delete").
		Register("add_updated_by", func(d *gorm.DB) {
			AddUpdatedByForUpdateDeleteCallback(ctx, d)
		})
}
