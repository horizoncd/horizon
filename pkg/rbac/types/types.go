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

package types

// attention: rbac is refers to the kubernetes rbac
// copy core struct and logics from the kubernetes code
// and do same modify
const (
	APIGroupAll    = "*"
	ResourceAll    = "*"
	VerbAll        = "*"
	ScopeAll       = "*"
	NonResourceAll = "*"
)

type Role struct {
	Name        string       `yaml:"name" json:"name"`
	Desc        string       `yaml:"desc" json:"desc"`
	PolicyRules []PolicyRule `yaml:"rules" json:"rules"`
}

type PolicyRule struct {
	Verbs           []string `yaml:"verbs" json:"verbs"`
	APIGroups       []string `yaml:"apiGroups" json:"apiGroups"`
	Resources       []string `yaml:"resources" json:"resources"`
	Scopes          []string `yaml:"scopes" json:"scopes"`
	NonResourceURLs []string `yaml:"nonResourceURLs" json:"nonResourceURLs"`
}
