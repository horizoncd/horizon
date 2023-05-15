// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"time"
)

type EventType string
type EventResourceType string

const (
	Any                    string = "*"
	ApplicationCreated     string = "applications_created"
	ApplicationDeleted     string = "applications_deleted"
	ApplicationUpdated     string = "applications_updated"
	ApplicationTransfered  string = "applications_transferred"
	ClusterCreated         string = "clusters_created"
	ClusterDeleted         string = "clusters_deleted"
	ClusterBuildDeployed   string = "clusters_builddeployed"
	ClusterDeployed        string = "clusters_deployed"
	ClusterRollbacked      string = "clusters_rollbacked"
	ClusterRestarted       string = "clsuters_restarted"
	ClusterPodsRescheduled string = "clusters_rescheduled"
	ClusterUpdated         string = "clusters_updated"
	ClusterFreed           string = "clusters_freed"
	ClusterKubernetesEvent string = "clusters_kubernetes_event"
	ClusterAction                 = "clusters_action"
	// TODO: add group events
)

type EventSummary struct {
	ResourceType string
	ResourceID   uint
	EventType    string
	Extra        *string `gorm:"default:''"`
}

type Event struct {
	EventSummary
	ID        uint
	ReqID     string
	CreatedAt time.Time
	CreatedBy uint
}

type EventCursorType string

const (
	// CursorHorizon is the cursor for taking out events from db
	CursorHorizon EventCursorType = "horizon"
	// CursorRegion is the cursor for recording the resourceVersion of the last written event from kubernetes
	CursorRegion EventCursorType = "region"
)

type EventCursor struct {
	ID        uint
	Position  uint
	Type      EventCursorType `gorm:"default:'horizon';uniqueIndex:idx_type_region_id"`
	RegionID  uint            `gorm:"column:region_id;uniqueIndex:idx_type_region_id"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ActionWithDescription struct {
	Name        EventType `json:"name"`
	Description string    `json:"description"`
}
