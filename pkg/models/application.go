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

type Priority string

const (
	P0 Priority = "P0"
	P1 Priority = "P1"
	P2 Priority = "P2"
	P3 Priority = "P3"
)

type Application struct {
	global.Model
	GroupID         uint
	Name            string
	Description     string
	Priority        Priority
	GitURL          string
	GitSubfolder    string
	GitRef          string
	GitRefType      string
	Image           string
	Template        string
	TemplateRelease string
	CreatedBy       uint
	UpdatedBy       uint
}

// ApplicationDetail contains the fullPath
type ApplicationDetail struct {
	Application
	FullPath string
	FullName string
}
