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

package pipelinerun

import "time"

type GetDiffResponse struct {
	CodeInfo   *CodeInfo   `json:"codeInfo"`
	ConfigDiff *ConfigDiff `json:"configDiff"`
}

type CodeInfo struct {
	// deploy branch info
	Branch string `json:"branch,omitempty"`
	// deploy tag info
	Tag string `json:"tag,omitempty"`
	// branch commit
	CommitID string `json:"commitID"`
	// commit message
	CommitMsg string `json:"commitMsg"`
	// code history link
	Link string `json:"link"`
}

type ConfigDiff struct {
	From string `json:"from"`
	To   string `json:"to"`
	Diff string `json:"diff"`
}

type BuildDeployRequestGit struct {
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Commit string `json:"commit"`
}

type CreatePrMessageRequest struct {
	Content string `json:"string"`
}

type User struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	UserType string `json:"userType,omitempty"`
}

type PrMessage struct {
	Content   string    `json:"content"`
	CreatedBy User      `json:"createdBy"`
	UpdatedBy User      `json:"updatedBy"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateOrUpdateCheckRunRequest struct {
	Name       string `json:"name"`
	CheckID    uint   `json:"checkId"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	ExternalID string `json:"externalId"`
	DetailURL  string `json:"detailUrl"`
}
