package region

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	reader := bytes.NewReader([]byte(`
defaultRegions:
  dev,test,reg,perf,beta: hz-test
  pre,online: hz
`))

	config, err := LoadRegionConfig(reader)
	assert.Nil(t, err)

	b, err := json.MarshalIndent(config, "", "  ")
	assert.Nil(t, err)
	t.Logf("%v", string(b))

	reader = bytes.NewReader([]byte(`
defaultRegions:
dev,test,reg,perf,beta: hz-test
  pre,online: hz

`))
	_, err = LoadRegionConfig(reader)
	assert.NotNil(t, err)
}
