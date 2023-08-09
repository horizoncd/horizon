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
	KindValidating  Kind      = "validating"
	KindMutating    Kind      = "mutating"
	OperationCreate Operation = "CREATE"
	OperationUpdate Operation = "UPDATE"
	OperationDelete Operation = "DELETE"
	OperationAll    Operation = "*"

	PatchTypeJSONPatch = "JSONPatch"
)
