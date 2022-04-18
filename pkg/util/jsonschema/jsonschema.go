package jsonschema

import (
	"encoding/json"
	"fmt"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	v5jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

var (
	unevaluatedProperties = "unevaluatedProperties"
	properties            = "properties"
)

// Validate json by jsonschema.
// schema and document support 2 types: string, map[string]interface{}
func Validate(schema, document interface{}) error {
	// add "unevaluatedProperties": false
	// change schema type to Golang map
	var schemaMap map[string]interface{}
	switch schema := schema.(type) {
	case string:
		err := json.Unmarshal([]byte(schema), &schemaMap)
		if err != nil {
			return perror.Wrap(herrors.ErrParamInvalid,
				fmt.Sprintf("json unmarshal error, schema: %s, error: %s", schema, err.Error()))
		}
	case map[string]interface{}:
		schemaMap = schema
	default:
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("unsported type: %T for schema", schema))
	}
	addUnevaluatedPropertiesField(schemaMap)

	var v interface{}
	switch document := document.(type) {
	case string:
		if err := json.Unmarshal([]byte(document), &v); err != nil {
			return perror.Wrap(herrors.ErrParamInvalid,
				fmt.Sprintf("json unmarshal error, document: %s, error: %s", document, err.Error()))
		}
	case map[string]interface{}:
		doc, err := json.Marshal(document)
		if err != nil {
			return perror.Wrap(herrors.ErrParamInvalid,
				fmt.Sprintf("json marshal error, document: %s, error: %s", document, err.Error()))
		}
		if err := json.Unmarshal(doc, &v); err != nil {
			return perror.Wrap(herrors.ErrParamInvalid,
				fmt.Sprintf("json unmarshal error, document: %s, error: %s", document, err.Error()))
		}
	default:
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("unsported type: %T for document", document))
	}

	schemaStr, err := json.Marshal(schemaMap)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("json marshal error, document: %s, error: %s", document, err.Error()))
	}
	sch, err := v5jsonschema.CompileString("schema.json", string(schemaStr))
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("jsonschema compilestring error, schema: %s, error: %s", schemaStr, err.Error()))
	}
	if err = sch.Validate(v); err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	return nil
}

// addUnevaluatedPropertiesField add "unevaluatedProperties": false to the jsonschema
// which means no additional properties will be allowed.
func addUnevaluatedPropertiesField(m map[string]interface{}) map[string]interface{} {
	_, propertiesExist := m[properties]
	_, unevaluatedPropertiesExist := m[unevaluatedProperties]
	// ignore when schema has already set unevaluatedProperties field
	if propertiesExist && !unevaluatedPropertiesExist {
		m[unevaluatedProperties] = false
	}

	for _, v := range m {
		v1, ok := v.(map[string]interface{})
		if ok {
			addUnevaluatedPropertiesField(v1)
		}
	}

	return m
}
