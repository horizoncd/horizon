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

package permission

import (
	"context"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
)

func OnlySelfAndAdmin(ctx context.Context, self uint) error {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	if !currentUser.IsAdmin() && currentUser.GetID() != self {
		return perror.Wrap(herrors.ErrForbidden, "you can only access resources of yourself")
	}
	return nil
}
