package models

import (
	"time"
)

type EventType string
type EventResourceType string

const (
	Any                   string = "*"
	ApplicationCreated    string = "applications_created"
	ApplicationDeleted    string = "applications_deleted"
	ApplicationTransfered string = "applications_transferred"
	ClusterCreated        string = "clusters_created"
	ClusterDeleted        string = "clusters_deleted"
	ClusterBuildDeployed  string = "clusters_builddeployed"
	ClusterDeployed       string = "clusters_deployed"
	ClusterRollbacked     string = "clusters_rollbacked"
	ClusterFreed          string = "clusters_freed"
)

type EventSummary struct {
	ResourceType string
	ResourceID   uint
	EventType    string
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
	Name        EventType `json:"name"`
	Description string    `json:"description"`
}
