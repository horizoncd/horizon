package models

import "strings"

type Kind string

func (k Kind) String() string {
	return strings.ToLower(string(k))
}

func (k Kind) Eq(other Kind) bool {
	return strings.EqualFold(string(k), string(other))
}

type Operation string

func (o Operation) Eq(other Operation) bool {
	return strings.EqualFold(string(o), string(other))
}

const (
	MatchAll string = "*"

	KindValidating Kind = "validating"
)
