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

package user

import (
	"fmt"
	"strconv"
)

// User describes a user that has been authenticated to the system
type User interface {
	GetName() string
	GetFullName() string
	GetID() uint
	GetEmail() string
	String() string
	IsAdmin() bool

	GetStrID() string
}

type DefaultInfo struct {
	Name     string
	FullName string
	ID       uint
	Email    string
	Admin    bool
}

func (d *DefaultInfo) GetName() string {
	return d.Name
}

func (d *DefaultInfo) GetID() uint {
	return d.ID
}

func (d *DefaultInfo) GetFullName() string {
	return d.FullName
}

func (d *DefaultInfo) GetEmail() string {
	return d.Email
}

func (d *DefaultInfo) String() string {
	return fmt.Sprintf("%s(%d)", d.Name, d.ID)
}

func (d *DefaultInfo) IsAdmin() bool {
	return d.Admin
}

func (d *DefaultInfo) GetStrID() string {
	return strconv.FormatUint(uint64(d.GetID()), 10)
}
