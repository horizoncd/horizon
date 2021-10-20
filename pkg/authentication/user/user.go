package user

import "fmt"

// User describes a user that has been authenticated to the system
type User interface {
	GetName() string
	GetID() uint
	GetFullName() string
	String() string
}

type DefaultInfo struct {
	Name     string
	FullName string
	ID       uint
}

func (d *DefaultInfo) GetName() string {
	return d.Name
}

func (d *DefaultInfo) GetID() uint {
	return d.ID
}

func (d *DefaultInfo) GetFullName() string {
	return d.FullName
}

func (d *DefaultInfo) String() string {
	return fmt.Sprintf("%s(%d)", d.Name, d.ID)
}
