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
	"strings"

	"github.com/horizoncd/horizon/pkg/server/global"
)

const RootGroupID = 0

type Group struct {
	global.Model
	Name            string
	Path            string
	VisibilityLevel string
	Description     string
	ParentID        uint
	TraversalIDs    string
	RegionSelector  string
	CreatedBy       uint
	UpdatedBy       uint
}

type GroupRegionSelectors struct {
	*Group
	RegionSelectors RegionSelectors
}

type GroupOrApplication struct {
	global.Model
	Name        string
	Path        string
	Description string
	Type        string
}

type Groups []*Group

// Len the length of the groups
func (g Groups) Len() int {
	return len(g)
}

// Less sort groups by the size of the traversalIDs array after split by ','
func (g Groups) Less(i, j int) bool {
	return len(strings.Split(g[i].TraversalIDs, ",")) < len(strings.Split(g[j].TraversalIDs, ","))
}

// Swap the two group
func (g Groups) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

type RegionSelector struct {
	Key      string   `json:"key"`
	Values   []string `json:"values"`
	Operator string   `json:"operator" default:"in"` // not used currently
}

type RegionSelectors []*RegionSelector
