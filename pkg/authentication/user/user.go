package user

import "fmt"

// User describes a user that has been authenticated to the system
type User interface {
	GetName() string
	GetFullName() string
	GetID() uint
	GetEmail() string
	String() string
	IsAdmin() bool
}

type DefaultInfo struct {
	Name     string
	FullName string
	ID       uint
	Email    string
	Admin    bool
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

func (d *DefaultInfo) GetEmail() string {
	return d.Email
}

func (d *DefaultInfo) String() string {
	return fmt.Sprintf("%s(%d)", d.Name, d.ID)
}

func (d *DefaultInfo) IsAdmin() bool {
	return d.Admin
}
