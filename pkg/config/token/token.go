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

package token

import "time"

type Config struct {
	// JwtSigningKey is used to sign JWT tokens
	JwtSigningKey string `yaml:"jwtSigningKey"`
	// CallbackTokenExpireIn is the expiration time of token for tekton callback
	CallbackTokenExpireIn time.Duration `yaml:"callbackTokenExpireIn"`
}
