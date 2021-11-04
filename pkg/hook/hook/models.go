package hook

type EventType string

const (
	CreateApplication EventType = "CreateApplication"
	DeleteApplication EventType = "DeleteApplication"
	CreateCluster     EventType = "CreateCluster"
	DeleteCluster     EventType = "DeleteCluster"
)

type Event struct {
	EventType EventType
	Event     interface{}
}
