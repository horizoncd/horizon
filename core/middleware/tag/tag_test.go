package tag

import (
	"net/http"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/pkg/tag/models"
	"github.com/gin-gonic/gin"
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
