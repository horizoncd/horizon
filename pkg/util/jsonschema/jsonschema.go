package jsonschema

import (
	"fmt"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// Validate json by jsonschema.
// schema and document support 3 types: string, []byte, map[string]interface{}
func Validate(schema, document interface{}) error {
	var schemaLoader, documentLoader gojsonschema.JSONLoader

	switch schema := schema.(type) {
	case string:
		schemaLoader = gojsonschema.NewStringLoader(schema)
	case []byte:
		schemaLoader = gojsonschema.NewBytesLoader(schema)
	case map[string]interface{}:
		schemaLoader = gojsonschema.NewGoLoader(schema)
	default:
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("unsported type: %T for schema", schema))
	}

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
