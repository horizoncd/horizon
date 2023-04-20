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

package cluster

type BuildDeployRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Git         *BuildDeployRequestGit `json:"git"`
}

type BuildDeployRequestGit struct {
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Commit string `json:"commit"`
}

type BuildDeployResponse struct {
	PipelinerunID uint `json:"pipelinerunID"`
}

type GetDiffResponse struct {
	CodeInfo   *CodeInfo `json:"codeInfo"`
	ConfigDiff string    `json:"configDiff"`
}

type CodeInfo struct {
	// deploy branch info
	Branch string `json:"branch,omitempty"`
	// deploy tag info
	Tag string `json:"tag,omitempty"`
	// current branch commit
	CommitID string `json:"commitID"`
	// commit message
	CommitMsg string `json:"commitMsg"`
	// code history link
	Link string `json:"link"`
}
