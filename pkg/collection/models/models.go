package models

type Collection struct {
	ID uint `gorm:"primarykey"`

	ResourceID   uint `gorm:"column:resource_id"`
	ResourceType string
	UserID       uint `gorm:"column:user_id"`
}
