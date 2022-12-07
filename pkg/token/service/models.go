package service

type TokenType uint8

const (
	NeverExpire     = "never"
	ExpiresAtFormat = "2006-01-02"

	TypeUserAccessToken     TokenType = 1
	TypeInternalAccessToken TokenType = 2
)
