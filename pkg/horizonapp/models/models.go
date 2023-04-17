package models

// OwnerType represents the type of owner for a HorizonApp.
type OwnerType uint8

const (
	GroupOwnerType OwnerType = 1
)

// HorizonApp represents an application in the Horizon system.
type HorizonApp struct {
	ID      uint
	Name    string
	Desc    string
	HomeURL string
	WebHook string

	OauthAppID uint

	OwnerType OwnerType
	OwnerID   uint
}

// Permissions represents a set of permissions.
type Permissions struct {
	ID          uint
	Permissions string
}
