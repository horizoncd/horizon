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

package service

import "time"

// Child subResource of a group, including group & application
type Child struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Path            string    `json:"path"`
	VisibilityLevel string    `json:"visibilityLevel"`
	Description     string    `json:"description"`
	ParentID        uint      `json:"parentID"`
	TraversalIDs    string    `json:"traversalIDs"`
	UpdatedAt       time.Time `json:"updatedAt"`
	FullName        string    `json:"fullName"`
	FullPath        string    `json:"fullPath"`
	Type            string    `json:"type"`
	ChildrenCount   int       `json:"childrenCount"`
	Children        []*Child  `json:"children"`
}

// Full is the fullName&fullPath for a group/application/applicationInstance
type Full struct {
	FullName string `json:"fullName"`
	FullPath string `json:"fullPath"`
}
