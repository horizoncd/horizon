package jsonschema

import (
	"encoding/json"
	"fmt"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

var (
	additionalProperties = "additionalProperties"
	properties           = "properties"
)

// Validate json by jsonschema.
// schema and document support 3 types: string, []byte, map[string]interface{}
func Validate(schema, document interface{}) error {
	// add "additionalProperties": false
	// change schema type to Golang map
	var schemaMap map[string]interface{}
	switch schema := schema.(type) {
	case string:
		err := json.Unmarshal([]byte(schema), &schemaMap)
		if err != nil {
			return perror.Wrap(herrors.ErrParamInvalid,
				fmt.Sprintf("unsported type: %T for schema", schema))
		}
	case []byte:
		err := json.Unmarshal(schema, &schemaMap)
		if err != nil {
			return perror.Wrap(herrors.ErrParamInvalid,
				fmt.Sprintf("unsported type: %T for schema", schema))
		}
	case map[string]interface{}:
		schemaMap = schema
	default:
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("unsported type: %T for schema", schema))
	}
	addAdditionalPropertiesField(schemaMap)

	var schemaLoader, documentLoader gojsonschema.JSONLoader
	schemaLoader = gojsonschema.NewGoLoader(schemaMap)

	switch document := document.(type) {
	case string:
		documentLoader = gojsonschema.NewStringLoader(document)
	case []byte:
		documentLoader = gojsonschema.NewBytesLoader(document)
	case map[string]interface{}:
		documentLoader = gojsonschema.NewGoLoader(document)
	default:
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("unsported type: %T for document", document))
	}

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	if result.Valid() {
		return nil
	}
	errMsg := ""
	for index, err := range result.Errors() {
		errMsg += fmt.Sprintf("[%d] %v. ", index, err)
	}
	return perror.Wrap(herrors.ErrParamInvalid, errMsg)
}

// addAdditionalPropertiesField add "additionalProperties": false to the jsonschema.
// no additional properties will be allowed.
func addAdditionalPropertiesField(m map[string]interface{}) map[string]interface{} {
	_, ok := m[properties]
	if ok {
		m[additionalProperties] = false
	}

	for _, v := range m {
		v1, ok := v.(map[string]interface{})
		if ok {
			addAdditionalPropertiesField(v1)
		}
	}

	return m
}
