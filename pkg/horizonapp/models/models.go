package models

type OwnerType uint8

const (
	GroupOwnerType OwnerType = 1
)

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

type Permissions struct {
	ID          uint
	Permissions string
}
