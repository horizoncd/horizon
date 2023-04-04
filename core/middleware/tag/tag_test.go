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

package tag

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	f := Middleware()
	expected := map[string][]string{
		"a": {"b"},
		"c": {"d"},
	}

	// %3d -> "="
	// %2c -> ","
	req, err := http.NewRequest(http.MethodGet, "/?tagSelector=a%3db%2cc%3dd", nil)
	assert.Nil(t, err)
	ctx := &gin.Context{}
	ctx.Request = req
	f(ctx)

	tssi, ok := ctx.Get(common.TagSelector)
	assert.True(t, ok)

	tss, ok := tssi.([]models.TagSelector)
	assert.True(t, ok)

	for _, ts := range tss {
		v, ok := expected[ts.Key]
		assert.True(t, ok)
		assert.Equal(t, v, ts.Values.List())
	}
}
