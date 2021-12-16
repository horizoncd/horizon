package template

import (
	"bytes"
	"encoding/json"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestYamlTemplate(t *testing.T) { // nolint
	yamlTextTemplate := `
SyncDomain:
   Description: The URL to access the service
   Value:  {{ .horizon.cluster }}.{{ .env.ingressDomain }}
   Value1:  {{ .horizon.cluster1 }}.{{ .env.ingressDomain1 }}
AsyncDomain:
   Description: The URL to access the service
   Value:  {{ .horizon.cluster }}-async.{{ .env.ingressDomain }}
`

	vals := make(map[string]interface{})
	horizonVals := make(map[string]interface{})
	envVals := make(map[string]interface{})

	horizonVals["cluster"] = "serverless-demo"
	envVals["ingressDomain"] = "serverless.horizon.com"

	vals["horizon"] = horizonVals
	vals["env"] = envVals

	var b bytes.Buffer
	template2 := template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(yamlTextTemplate))
	err := template2.ExecuteTemplate(&b, "", vals)
	assert.Nil(t, err)
	t.Logf("the bytes is %s", b.String())

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(b.Bytes(), &m)
	assert.Nil(t, err)
	t.Logf("the unmarshal struct is %+v", m)

	mjson := convert(m)
	jsonByte, err := json.MarshalIndent(&mjson, "", " ")
	assert.Nil(t, err)
	t.Logf("the json struct is %s", string(jsonByte))
}

func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}
