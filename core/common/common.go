package common

type Resource struct {
	ID   uint   `json:"resource_id" yaml:"resourceID"`
	Type string `json:"resource_type" yaml:"resourceType"`
}
