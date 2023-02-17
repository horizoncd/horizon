package models

type CollectionResourceType string

const (
	ResourceCluster = CollectionResourceType("cluster")
)

type Collection struct {
	ID uint `gorm:"primarykey"`

	ResourceID   uint `gorm:"column:resource_id"`
	ResourceType CollectionResourceType
	UserID       uint `gorm:"column:user_id"`
}
