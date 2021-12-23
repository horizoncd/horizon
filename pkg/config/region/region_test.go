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

groupRegions:
  dev,test,reg,perf,beta:
    1,2: hz-test
    3,4: hz-test
  pre,online:
    1,2: hz
    3,4: hz

applicationRegions:
  dev,test,reg,perf,beta:
    app1,app2: hz-test
    app3,app4: eks-test
  pre,online:
    app1,app2: hz
    app3,app4: eks
`))

	config, err := LoadRegionConfig(reader)
	assert.Nil(t, err)

	b, err := json.MarshalIndent(config, "", "  ")
	assert.Nil(t, err)
	t.Logf("%v", string(b))

	assert.Equal(t, "hz-test", config.ApplicationRegions["beta"]["app2"])
	assert.Equal(t, "eks-test", config.ApplicationRegions["beta"]["app3"])
	assert.Equal(t, "hz", config.ApplicationRegions["pre"]["app2"])
	assert.Equal(t, "eks", config.ApplicationRegions["online"]["app3"])
	assert.Equal(t, "hz-test", config.GroupRegions["test"]["1"])
	assert.Equal(t, "hz-test", config.GroupRegions["test"]["4"])
	assert.Equal(t, "hz", config.GroupRegions["pre"]["2"])
	assert.Equal(t, "hz", config.GroupRegions["online"]["3"])

	reader = bytes.NewReader([]byte(`
defaultRegions:
dev,test,reg,perf,beta: hz-test
  pre,online: hz

applicationRegions:
  dev,test,reg,perf,beta:
    app1,app2: hz-test
    app3,app4: eks-test
  pre,online:
    app1,app2: hz
    app3,app4: eks
`))
	_, err = LoadRegionConfig(reader)
	assert.NotNil(t, err)
}
