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

package cd

import "context"

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/cd/lagacycd_mock.go -package=mock_cd -aux_files=github.com/horizoncd/horizon/pkg/cd=cd.go
type LegacyCD interface {
	CD

	GetClusterStateV1(ctx context.Context, params *GetClusterStateParams) (*ClusterState, error)
}
