package common

type Resource struct {
	ResourceID uint   `json:"resource_id" yaml:"resourceID"`
	Type       string `gorm:"column:resource_type" json:"resource_type" yaml:"resourceType"`
}
