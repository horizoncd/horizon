package models

import (
	"strings"

	"g.hz.netease.com/horizon/pkg/server/global"
)

type Group struct {
	global.Model
	Name               string
	Path               string
	VisibilityLevel    string
	Description        string
	ParentID           uint
	TraversalIDs       string
	KubernetesSelector string
	CreatedBy          uint
	UpdatedBy          uint
}

type GroupRegionSelectors struct {
	*Group
	RegionSelectors KubernetesSelectors
}

type GroupOrApplication struct {
	global.Model
	Name        string
	Path        string
	Description string
	Type        string
}

type Groups []*Group

// Len the length of the groups
func (g Groups) Len() int {
	return len(g)
}

// Less sort groups by the size of the traversalIDs array after split by ','
func (g Groups) Less(i, j int) bool {
	return len(strings.Split(g[i].TraversalIDs, ",")) < len(strings.Split(g[j].TraversalIDs, ","))
}

// Swap the two group
func (g Groups) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

type KubernetesSelector struct {
	Key      string   `json:"key"`
	Values   []string `json:"values"`
	Operator string   `json:"operator" default:"in"` // not used currently
}

type KubernetesSelectors []*KubernetesSelector
