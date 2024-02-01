package service

import (
	"context"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/event/manager"
	"github.com/horizoncd/horizon/pkg/event/models"
	membermanager "github.com/horizoncd/horizon/pkg/member/manager"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/util/log"
)

type Service interface {
	// CreateEventIgnoreError creates an event and ignore the error
	CreateEventIgnoreError(ctx context.Context, resourceType string,
		resourceID uint, eventType string, extra *string) []*models.Event
	// CreateEventsIgnoreError creates events and ignore the error
	CreateEventsIgnoreError(ctx context.Context, events ...*models.Event) []*models.Event
	// RecordMemberCreatedEvent records members_created event for the given resource that is created
	RecordMemberCreatedEvent(ctx context.Context, resourceType string, resourceID uint) []*models.Event
}

type service struct {
	eventMgr  manager.Manager
	memberMgr membermanager.Manager
}

func New(manager *managerparam.Manager) Service {
	return &service{
		eventMgr:  manager.EventMgr,
		memberMgr: manager.MemberMgr,
	}
}

func (s *service) CreateEventIgnoreError(ctx context.Context, resourceType string,
	resourceID uint, eventType string, extra *string) []*models.Event {
	events, err := s.eventMgr.CreateEvent(ctx, &models.Event{
		EventSummary: models.EventSummary{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			EventType:    eventType,
			Extra:        extra,
		},
	})
	if err != nil {
		log.Warningf(ctx, "failed to create event, err: %s", err.Error())
		return nil
	}
	return events
}

func (s *service) CreateEventsIgnoreError(ctx context.Context, events ...*models.Event) []*models.Event {
	events, err := s.eventMgr.CreateEvent(ctx, events...)
	if err != nil {
		log.Warningf(ctx, "failed to create event, err: %s", err.Error())
		return nil
	}
	return events
}

func (s *service) RecordMemberCreatedEvent(ctx context.Context, resourceType string,
	resourceID uint) []*models.Event {
	var members []membermodels.Member
	var err error
	switch resourceType {
	case common.ResourceApplication:
		members, err = s.memberMgr.ListDirectMember(ctx, membermodels.TypeApplication, resourceID)
	case common.ResourceCluster:
		members, err = s.memberMgr.ListDirectMember(ctx, membermodels.TypeApplicationCluster, resourceID)
	default:
		log.Warningf(ctx, "unsupported resource type: %s", resourceType)
		return nil
	}
	if err != nil {
		log.Warningf(ctx, "failed to list members of resource, err: %s", err.Error())
		return nil
	}
	events := make([]*models.Event, 0, len(members))
	for _, m := range members {
		events = append(events, &models.Event{
			EventSummary: models.EventSummary{
				ResourceType: common.ResourceMember,
				ResourceID:   m.ID,
				EventType:    models.MemberCreated,
			},
		})
	}
	if len(events) > 0 {
		retEvents, err := s.eventMgr.CreateEvent(ctx, events...)
		if err != nil {
			log.Warningf(ctx, "failed to create event, err: %s", err.Error())
		}
		return retEvents
	}
	return nil
}
