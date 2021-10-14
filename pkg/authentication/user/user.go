package user

// User describes a user that has been authenticated to the system
type User interface {
	GetName() string
	GetID() uint
	GetFullName() string
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
