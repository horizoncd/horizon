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

package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
)

type Environment struct {
	global.Model

	Name        string
	DisplayName string
	CreatedBy   uint
	UpdatedBy   uint
}

type EnvironmentList []*Environment

func (e EnvironmentList) Len() int {
	return len(e)
}

func (e EnvironmentList) Less(i, j int) bool {
	const pre = "pre"
	const online = "online"
	if e[i].Name == online {
		return false
	}
	if e[j].Name == online {
		return true
	}
	if e[i].Name == pre {
		return false
	}
	if e[j].Name == pre {
		return true
	}
	return e[i].Name < e[j].Name
}

func (e EnvironmentList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
