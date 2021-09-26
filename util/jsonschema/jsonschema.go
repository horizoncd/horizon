package jsonschema

import (
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// Validate json by jsonschema.
// schema and document support 3 types: string, []byte, map[string]interface{}
func Validate(schema, document interface{}) error {
	var schemaLoader, documentLoader gojsonschema.JSONLoader

	switch schema.(type) {
	case string:
		schemaLoader = gojsonschema.NewStringLoader(schema.(string))
	case []byte:
		schemaLoader = gojsonschema.NewBytesLoader(schema.([]byte))
	case map[string]interface{}:
		schemaLoader = gojsonschema.NewGoLoader(schema.(map[string]interface{}))
	default:
		return fmt.Errorf("unsported type: %T for schema", schema)
	}

	switch document.(type) {
	case string:
		documentLoader = gojsonschema.NewStringLoader(document.(string))
	case []byte:
		documentLoader = gojsonschema.NewBytesLoader(document.([]byte))
	case map[string]interface{}:
		documentLoader = gojsonschema.NewGoLoader(document.(map[string]interface{}))
	default:
		return fmt.Errorf("unsported type: %T for document", document)
	}

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	} else {
		errMsg := ""
		for index, err := range result.Errors() {
			errMsg += fmt.Sprintf("[%d] %v. ", index, err)
		}
		return errors.New(errMsg)
	}
}
