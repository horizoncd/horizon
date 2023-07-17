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

package service

import (
	"context"

	groupmanager "github.com/horizoncd/horizon/pkg/group/manager"
	"github.com/horizoncd/horizon/pkg/group/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
)

const (
	// ChildTypeGroup used to indicate the 'Child' is a group
	ChildTypeGroup = "group"
	// ChildTypeApplication ...
	ChildTypeApplication = "application"
	// ChildTypeCluster ...
	ChildTypeCluster = "cluster"

	ChildTypeTemplate = "template"

	ChildTypeRelease = "release"
	// RootGroupID id of the root group, which is not actually exists in the group table
	RootGroupID = 0
)

type Service interface {
	// GetChildByID get a child by id
	GetChildByID(ctx context.Context, id uint) (*Child, error)
	// GetChildrenByIDs returns children map according to group ids
	GetChildrenByIDs(ctx context.Context, ids []uint) (map[uint]*Child, error)
}

type service struct {
	groupManager groupmanager.Manager
}

func (s service) GetChildByID(ctx context.Context, id uint) (*Child, error) {
	group, err := s.groupManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	groups, err := s.groupManager.GetByIDs(ctx, groupmanager.FormatIDsFromTraversalIDs(group.TraversalIDs))
	if err != nil {
		return nil, err
	}

	full := GenerateFullFromGroups(groups)

	return ConvertGroupToChild(group, full), nil
}

func (s service) GetChildrenByIDs(ctx context.Context, ids []uint) (map[uint]*Child, error) {
	var groupIDs []uint
	// childrenMap store result
	childrenMap := map[uint]*Child{}
	// groupMap store all queried groups
	groupMap := map[uint]*models.Group{}

	// 1.query groups
	groups, err := s.groupManager.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	// 2.query parent groups by traversal id, and store in map
	for _, group := range groups {
		for _, groupID := range groupmanager.FormatIDsFromTraversalIDs(group.TraversalIDs) {
			groupMap[groupID] = nil
		}
	}
	for groupID := range groupMap {
		groupIDs = append(groupIDs, groupID)
	}
	parentGroups, err := s.groupManager.GetByIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	for i, group := range parentGroups {
		groupMap[group.ID] = parentGroups[i]
	}

	// 3.convert to children map
	for _, group := range groups {
		parentGroups = []*models.Group{}
		for _, id := range groupmanager.FormatIDsFromTraversalIDs(group.TraversalIDs) {
			parentGroups = append(parentGroups, groupMap[id])
		}
		full := GenerateFullFromGroups(parentGroups)
		childrenMap[group.ID] = ConvertGroupToChild(group, full)
	}

	return childrenMap, nil
}

func NewService(manager *managerparam.Manager) Service {
	return &service{
		groupManager: manager.GroupMgr,
	}
}
