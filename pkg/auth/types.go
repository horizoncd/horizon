package auth

import (
	"context"

	"g.hz.netease.com/horizon/pkg/authentication/user"
)

// attention: authorization is refers to the kubernetes rbac
// we just copy core struct and logics from the kubernetes code
// and do same modify

// Attributes is an interface used by an Authorizer to get information about a request
// that is used to make an authorization decision.
type Attributes interface {
	// GetUser return the user.Info object to authorize
	GetUser() user.User

	// GetVerb returns the kube verb associated with API requests
	// (this includes get, list,  create, update, patch, delete),
	// or the lowercased HTTP verb associated with non-API requests
	// (this includes get, put, post, patch, and delete)
	GetVerb() string

	// IsReadOnly true represent the request has no side effects.
	IsReadOnly() bool

	// GetScope return the scope of the rest resource
	GetScope() string

	// GetResource The kind of object, if a request is for a REST object.
	GetResource() string

	// GetSubResource returns the subresource being requested, if present
	GetSubResource() string

	// GetName returns the name of the object as parsed off the request.
	GetName() string

	// GetAPIGroup  The group of the resource, if a request is for a REST object.
	GetAPIGroup() string

	// GetAPIVersion  returns the version of the group requested,
	// if a request is for a REST object.
	GetAPIVersion() string

	// IsResourceRequest  returns true  for requests to API resources /apis/core/groups
	// and false for non-resource endpoints. /healthz
	IsResourceRequest() bool

	// GetPath returns the path of the request
	GetPath() string
}

type Authorizer interface {
	Authorize(ctx context.Context, a Attributes) (Decision, reason string, err error)
}

type Decision int

const (
	// DecisionDeny means that an authorizer decided to deny the action
	DecisionDeny Decision = iota
	// DecisionAllow means that an authorizer  decided to allow the action.
	DecisionAllow
)

// AttributesRecord implements Attributes interface.
type AttributesRecord struct {
	User            user.User
	Verb            string
	APIGroup        string
	APIVersion      string
	Resource        string
	SubResource     string
	Name            string
	Scope           string
	ResourceRequest bool
	Path            string
}

func (a AttributesRecord) GetUser() user.User {
	return a.User
}

func (a AttributesRecord) GetVerb() string {
	return a.Verb
}

func (a AttributesRecord) GetScope() string {
	return a.Scope
}

func (a AttributesRecord) GetAPIVersion() string {
	return a.APIVersion
}

func (a AttributesRecord) GetAPIGroup() string {
	return a.APIGroup
}

func (a AttributesRecord) IsReadOnly() bool {
	return a.Verb == "get" || a.Verb == "list"
}

func (a AttributesRecord) GetResource() string {
	return a.Resource
}

func (a AttributesRecord) GetSubResource() string {
	return a.SubResource
}

func (a AttributesRecord) GetName() string {
	return a.Name
}

func (a AttributesRecord) IsResourceRequest() bool {
	return a.ResourceRequest
}

func (a AttributesRecord) GetPath() string {
	return a.Path
}
