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

package common

import (
	"context"

	herror "github.com/horizoncd/horizon/core/errors"
	hctx "github.com/horizoncd/horizon/pkg/context"
)

func WithContextJWTTokenString(ctx context.Context, tokenStr string) context.Context {
	return context.WithValue(ctx, hctx.JWTTokenString, tokenStr)
}

func JWTTokenStringFromContext(ctx context.Context) (string, error) {
	str, ok := ctx.Value(hctx.JWTTokenString).(string)
	if !ok {
		return "", herror.ErrFailedToGetJWTToken
	}
	return str, nil
}
