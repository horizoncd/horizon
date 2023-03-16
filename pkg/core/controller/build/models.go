package build

type Schema struct {
	JSONSchema map[string]interface{} `json:"jsonSchema"`
	UISchema   map[string]interface{} `json:"uiSchema"`
}
