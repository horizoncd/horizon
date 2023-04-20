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

package manager

import (
	"context"

	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/core/middleware/requestid"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/event/dao"
	"github.com/horizoncd/horizon/pkg/event/models"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

type Manager interface {
	CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error)
	ListEvents(ctx context.Context, offset, limit int) ([]*models.Event, error)
	ListEventsByRange(ctx context.Context, start, end uint) ([]*models.Event, error)
	CreateOrUpdateCursor(ctx context.Context,
		eventIndex *models.EventCursor) (*models.EventCursor, error)
	GetCursor(ctx context.Context) (*models.EventCursor, error)
	GetEvent(ctx context.Context, id uint) (*models.Event, error)
	ListSupportEvents() map[string]string
}

type manager struct {
	dao dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}

func (m *manager) CreateEvent(ctx context.Context,
	event *models.Event) (*models.Event, error) {
	const op = "event manager: create event"
	defer wlog.Start(ctx, op).StopPrint()

	if event.ReqID == "" {
		rid, err := requestid.FromContext(ctx)
		if err != nil {
			log.Warningf(ctx, "failed to get request id, err: %s", err.Error())
		}
		event.ReqID = rid
	}
	e, err := m.dao.CreateEvent(ctx, event)
	if err != nil {
		return nil, herrors.NewErrCreateFailed(herrors.EventInDB, err.Error())
	}

	return e, nil
}

func (m *manager) CreateOrUpdateCursor(ctx context.Context,
	eventCursor *models.EventCursor) (*models.EventCursor, error) {
	const op = "event manager: create or update cursor"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.CreateOrUpdateCursor(ctx, eventCursor)
}

func (m *manager) GetCursor(ctx context.Context) (*models.EventCursor, error) {
	const op = "event manager: get cursor"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.GetCursor(ctx)
}

func (m *manager) ListEvents(ctx context.Context, offset, limit int) ([]*models.Event, error) {
	const op = "event manager: list events"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListEvents(ctx, &q.Query{
		Keywords: q.KeyWords{
			common.Offset: offset,
			common.Limit:  limit,
		},
	})
}

func (m *manager) ListEventsByRange(ctx context.Context, start, end uint) ([]*models.Event, error) {
	const op = "event manager: list events by range"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListEvents(ctx, &q.Query{
		Keywords: q.KeyWords{
			common.StartID: start,
			common.EndID:   end,
		},
	})
}

func (m *manager) GetEvent(ctx context.Context, id uint) (*models.Event, error) {
	const op = "event manager: get event"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.GetEvent(ctx, id)
}

var supportedEvents = map[string]string{
	models.ApplicationCreated:     "New application has been created",
	models.ApplicationDeleted:     "Application has been deleted",
	models.ApplicationTransfered:  "Application has been transferred to another group",
	models.ApplicationUpdated:     "Application has been updated",
	models.ClusterCreated:         "New cluster has been created",
	models.ClusterDeleted:         "Cluster has been deleted",
	models.ClusterUpdated:         "Cluster has been updated",
	models.ClusterBuildDeployed:   "Cluster has completed a build task and triggered a deploy task",
	models.ClusterDeployed:        "Cluster has triggered a deploying task",
	models.ClusterRollbacked:      "Cluster has triggered a rollback task",
	models.ClusterFreed:           "Cluster has been freed",
	models.ClusterRestarted:       "Cluster has been restarted",
	models.ClusterPodsRescheduled: "Pods has been deleted to reschedule",
}

func (m *manager) ListSupportEvents() map[string]string {
	return supportedEvents
}
