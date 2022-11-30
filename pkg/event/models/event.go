package models

import (
	"time"

	"g.hz.netease.com/horizon/core/common"
)

type EventAction string
type EventResourceType string

const (
	Any string = "*"
	// resource type
	Group       EventResourceType = common.ResourceGroup
	Application EventResourceType = common.ResourceApplication
	Cluster     EventResourceType = common.ResourceCluster

	// common actions
	Created EventAction = "created"
	Deleted EventAction = "deleted"

	// cluster actions
	Transferred EventAction = "transferred"

	// cluster actions
	BuildDeployed EventAction = "buildDeployed"
	Deployed      EventAction = "deployed"
	Rollbacked    EventAction = "rollbacked"
	Freed         EventAction = "freed"
)

type EventSummary struct {
	ResourceType EventResourceType
	ResourceID   uint
	Action       EventAction
}

type Event struct {
	EventSummary
	ID        uint
	ReqID     string
	CreatedAt time.Time
	CreatedBy uint
}

type EventCursor struct {
	ID        uint
	Position  uint
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ActionWithDescription struct {
	Name        EventAction
	Description string
}
