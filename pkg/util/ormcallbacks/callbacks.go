// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package callbacks

import (
	"github.com/horizoncd/horizon/core/common"
	"gorm.io/gorm"
)

const (
	_createdBy = "created_by"
	_updatedBy = "updated_by"
)

// addCreatedByUpdatedByForCreateCallback will set `created_by` and `updated_by` when creating records if fields exist
func addCreatedByUpdatedByForCreateCallback(db *gorm.DB) {
	currentUser, err := common.UserFromContext(db.Statement.Context)
	if err != nil {
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

// addUpdatedByForUpdateDeleteCallback will set `updated_by` when updating or deleting records if fields exist
func addUpdatedByForUpdateDeleteCallback(db *gorm.DB) {
	currentUser, err := common.UserFromContext(db.Statement.Context)
	if err != nil {
		return
	}
	field := db.Statement.Schema.LookUpField(_updatedBy)
	if field != nil {
		db.Statement.SetColumn(_updatedBy, currentUser.GetID())
	}
}

func RegisterCustomCallbacks(db *gorm.DB) {
	_ = db.Callback().Create().After("gorm:before_create").Before("gorm:create").
		Register("add_created_by", addCreatedByUpdatedByForCreateCallback)

	_ = db.Callback().Update().After("gorm:before_update").Before("gorm:update").
		Register("add_updated_by", addUpdatedByForUpdateDeleteCallback)

	_ = db.Callback().Delete().After("gorm:before_delete").Before("gorm:delete").
		Register("add_updated_by", addUpdatedByForUpdateDeleteCallback)
}
