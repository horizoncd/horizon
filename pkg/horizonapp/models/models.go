package models

type HorizonApp struct {
	ID      uint
	Name    string
	Desc    string
	HomeURL string
	WebHook string

	OauthAppID uint
}

type Permissions struct {
	ID          uint
	Permissions string
}
