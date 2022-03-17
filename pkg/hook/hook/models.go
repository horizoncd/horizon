package hook

import (
	"context"
	"time"
)

type EventType string

const (
	CreateApplication EventType = "CreateApplication"
	DeleteApplication EventType = "DeleteApplication"
	CreateCluster     EventType = "CreateCluster"
	DeleteCluster     EventType = "DeleteCluster"
)

var (
	DefaultDelay = 1 * time.Second
	MaxDelay     = 1000 * time.Second
)

type Event struct {
	EventType EventType
	Event     interface{}
}

type EventCtx struct {
	EventType   EventType
	Event       interface{}
	Ctx         context.Context
	FailedTimes uint
}
