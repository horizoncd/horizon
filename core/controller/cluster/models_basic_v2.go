package cluster

import (
	"time"

	controllertag "g.hz.netease.com/horizon/core/controller/tag"
	codemodels "g.hz.netease.com/horizon/pkg/cluster/code"
)

type CreateClusterRequestV2 struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Priority    string               `json:"priority"`
	Git         *codemodels.Git      `json:"git"`
	Tags        []*controllertag.Tag `json:"tags"`

	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`

	// TODO(tom): just for internal usage
	ExtraMembers map[string]string `json:"extraMembers"`
}

type UpdateClusterRequestV2 struct {
	// basic infos
	Description string               `json:"description"`
	Priority    string               `json:"priority"`
	Tags        []*controllertag.Tag `json:"tags"`

	// env and region info (can only be modified after cluster freed)
	Environment string `json:"environment"`
	Region      string `json:"region"`

	// source info
	Git *codemodels.Git `json:"git"`

	// git config info
	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`
}

type GetClusterResponseV2 struct {
	// basic infos
	ID              uint                 `json:"id"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	Priority        string               `json:"priority"`
	Scope           *Scope               `json:"scope"`
	FullPath        string               `json:"fullPath"`
	ApplicationName string               `json:"applicationName"`
	ApplicationID   string               `json:"applicationID"`
	Tags            []*controllertag.Tag `json:"tags"`

	// source info
	Git *codemodels.Git `json:"git"`

	// git config info
	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`
	Manifest       map[string]interface{}   `json:"manifest"`

	// status and update info
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy *User     `json:"createdBy,omitempty"`
	UpdatedBy *User     `json:"updatedBy,omitempty"`
}
