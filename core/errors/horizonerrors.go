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

package errors

import (
	"fmt"

	"github.com/horizoncd/horizon/pkg/errors"
)

type sourceType struct {
	name string
}

var (
	TLS                       = sourceType{name: "TLS"}
	GitlabClient              = sourceType{name: "GitlabClient"}
	GitlabResource            = sourceType{name: "GitlabResource"}
	GithubResource            = sourceType{name: "GithubResource"}
	ClusterInDB               = sourceType{name: "ClusterInDB"}
	CollectionInDB            = sourceType{name: "CollectionInDB"}
	ClusterStateInArgo        = sourceType{name: "ClusterStateInArgo"}
	TagInDB                   = sourceType{name: "TagInDB"}
	ApplicationInArgo         = sourceType{name: "ApplicationInArgo"}
	ApplicationResourceInArgo = sourceType{name: "ApplicationResourceInArgo"}
	ApplicationInDB           = sourceType{name: "ApplicationInDB"}
	ApplicationRegionInDB     = sourceType{name: "ApplicationRegionInDB"}
	EnvironmentRegionInDB     = sourceType{name: "EnvironmentRegionInDB"}
	EnvironmentInDB           = sourceType{name: "EnvironmentInDB"}
	RegionInDB                = sourceType{name: "RegionInDB"}
	GroupInDB                 = sourceType{name: "GroupInDB"}
	K8SClient                 = sourceType{name: "K8SClient"}
	RegistryInDB              = sourceType{name: "RegistryInDB"}
	Pipelinerun               = sourceType{name: "Pipelinerun"}
	PipelinerunInTekton       = sourceType{name: "PipelinerunInTekton"}
	PipelinerunInDB           = sourceType{name: "PipelinerunInDB"}
	PipelineInDB              = sourceType{name: "PipelineInDB"}
	TaskInDB                  = sourceType{name: "TaskInDB"}
	StepInDB                  = sourceType{name: "StepInDB"}
	UserInDB                  = sourceType{name: "UserInDB"}
	UserLinkInDB              = sourceType{name: "UserLinkInDB"}
	TemplateSchemaTagInDB     = sourceType{name: "TemplateSchemaTagInDB"}
	TemplateInDB              = sourceType{name: "TemplateInDB"}
	TemplateReleaseInDB       = sourceType{name: "TemplateReleaseInDB"}
	TemplateReleaseInRepo     = sourceType{name: "TempalteReleaseInRepo"}
	MemberInfoInDB            = sourceType{name: "MemberInfoInDB"}
	ApplicationManifestInArgo = sourceType{name: "ApplicationManifestInArgo"}
	PodsInK8S                 = sourceType{name: "PodsInK8S"}
	PodLogsInK8S              = sourceType{name: "PodLogsInK8S"}
	ReplicasSetInK8S          = sourceType{name: "ReplicasSetInK8S"}
	DeploymentInK8S           = sourceType{name: "DeploymentInK8S"}
	ResourceInK8S             = sourceType{name: "ResourceInK8S"}
	PodEventInK8S             = sourceType{name: "PodEventInK8S"}
	KubeConfigInK8S           = sourceType{name: "KubeConfigK8S"}
	GroupFullPath             = sourceType{name: "GroupFullPath"}
	IdentityProviderInDB      = sourceType{name: "IdentityProviderInDB"}
	EventInDB                 = sourceType{name: "EventInDB"}
	EventCursorInDB           = sourceType{name: "EventCursorInDB"}
	WebhookInDB               = sourceType{name: "WebhookInDB"}
	WebhookLogInDB            = sourceType{name: "WebhookLogInDB"}
	MetatagInDB               = sourceType{name: "MetatagInDB"}

	// S3
	PipelinerunLog = sourceType{name: "PipelinerunLog"}
	PipelinerunObj = sourceType{name: "PipelinerunObj"}

	ArgoCD = sourceType{name: "ArgoCD"}

	Tekton          = sourceType{name: "Tekton"}
	TektonClient    = sourceType{name: "TektonClient"}
	TektonCollector = sourceType{name: "TektonCollector"}

	HelmRepo  = sourceType{name: "HelmRepo"}
	OAuthInDB = sourceType{name: "OauthAppClient"}
	TokenInDB = sourceType{name: "TokenInDB"}
	ChartFile = sourceType{name: "ChartFile"}

	// identity provider
	Oauth2Token           = sourceType{name: "Oauth2Token"}
	ProviderFromDiscovery = sourceType{name: "ProviderFromDiscovery"}

	StepInWorkload = sourceType{name: "StepInWorkload"}

	EnvValueInGit = sourceType{name: "EnvValueInGit"}
)

type HorizonErrNotFound struct {
	Source sourceType
}

func NewErrNotFound(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrNotFound{
		Source: source,
	}, msg)
}

func (e *HorizonErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", e.Source.name)
}

type HorizonErrGetFailed struct {
	Source sourceType
}

func NewErrGetFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrGetFailed{
		Source: source,
	}, msg)
}

func (e *HorizonErrGetFailed) Error() string {
	return fmt.Sprintf("%s get failed", e.Source.name)
}

type HorizonErrDeleteFailed struct {
	Source sourceType
}

func NewErrDeleteFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrDeleteFailed{
		Source: source,
	}, msg)
}

func (e *HorizonErrDeleteFailed) Error() string {
	return fmt.Sprintf("%s delete failed", e.Source.name)
}

type HorizonErrUpdateFailed struct {
	Source sourceType
}

func NewErrUpdateFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrUpdateFailed{
		Source: source,
	}, msg)
}

func (e *HorizonErrUpdateFailed) Error() string {
	return fmt.Sprintf("%s update failed", e.Source.name)
}

type HorizonErrInsertFailed struct {
	Source sourceType
}

func NewErrInsertFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrInsertFailed{
		Source: source,
	}, msg)
}

func (e *HorizonErrInsertFailed) Error() string {
	return fmt.Sprintf("%s insert failed", e.Source.name)
}

type HorizonErrCreateFailed struct {
	Source sourceType
}

func NewErrCreateFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrCreateFailed{
		Source: source,
	}, msg)
}

func (e *HorizonErrCreateFailed) Error() string {
	return fmt.Sprintf("%s create failed", e.Source.name)
}

type HorizonErrListFailed struct {
	Source sourceType
}

func NewErrListFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrListFailed{
		Source: source,
	}, msg)
}

func (e *HorizonErrListFailed) Error() string {
	return fmt.Sprintf("%s list failed", e.Source.name)
}

var (
	// universal
	ErrWriteFailed      = errors.New("write failed")
	ErrReadFailed       = errors.New("read failed")
	ErrNameConflict     = errors.New("name conflict")
	ErrPathConflict     = errors.New("path conflict")
	ErrPairConflict     = errors.New("entity pair conflict")
	ErrSubResourceExist = errors.New("sub resource exist")
	ErrNoPrivilege      = errors.New("no privilege")
	ErrParamInvalid     = errors.New("parameter is invalid")
	ErrForbidden        = errors.New("forbidden")
	ErrDeadlineExceeded = errors.New("time limit exceeded")
	ErrGenerateRandomID = errors.New("failed to generate random id")
	ErrDisabled         = errors.New("entity is disabled")
	ErrDuplicatedKey    = errors.New("duplicated keys")
	// ErrInternal = errors.New("internal error")

	// http
	ErrHTTPRespNotAsExpected = errors.New("http response is not as expected")
	ErrHTTPRequestFailed     = errors.New("http request failed")

	// cluster
	ErrClusterNoChange         = errors.New("no change to cluster")
	ErrShouldBuildDeployFirst  = errors.New("clusters with build config should build and deploy first")
	ErrBuildDeployNotSupported = errors.New("builddeploy is not supported for this cluster")

	// pipelinerun

	// context
	ErrFailedToGetORM       = errors.New("cannot get the ORM from context")
	ErrFailedToGetUser      = errors.New("cannot get user from context")
	ErrFailedToGetRequestID = errors.New("cannot get the requestID from context")
	ErrFailedToGetJWTToken  = errors.New("cannot get the jwt token from context")

	ErrKubeDynamicCliResponseNotOK = errors.New("response for kube dynamic cli is not 200 OK")
	ErrKubeExecFailed              = errors.New("kube exec failed")

	// S3
	ErrS3SignFailed   = errors.New("s3 sign failed")
	ErrS3PutObjFailed = errors.New("s3 put obj failed")
	ErrS3GetObjFailed = errors.New("s3 get obj failed")

	ErrGitlabInternal              = errors.New("gitlab internal")
	ErrGitlabMRNotReady            = errors.New("gitlab mr is not ready and cannot be merged")
	ErrGitlabResourceNotFound      = errors.New("gitlab resource not found")
	ErrGitLabDefaultBranchNotMatch = errors.New("gitlab default branch do not match")

	// git
	ErrBranchAndCommitEmpty      = errors.New("branch and commit cannot be empty at the same time")
	ErrGitlabInterfaceCallFailed = errors.New("failed to call gitlab interface")

	// pipeline
	ErrPipelineOutputEmpty = errors.New("pipeline output is empty")

	// restart
	ErrRestartFileEmpty = errors.New("restart file is empty")

	// tekton
	ErrTektonInternal = errors.New("tekton internal error")

	// helm
	ErrLoadChartArchive = errors.New("failed to load archive")

	// group
	// ErrHasChildren used when delete a group which still has some children
	ErrGroupHasChildren = errors.New("children exist, cannot be deleted")
	// ErrConflictWithApplication conflict with the application
	ErrGroupConflictWithApplication = errors.New("name or path is in conflict with application")

	// ErrOAuthSecretNotFound oauth clientid secret was not valid
	ErrOAuthSecretNotValid = errors.New("secret not valid")

	// ErrOAuthCodeExpired oauth authorization code  or access token expire
	ErrOAuthCodeExpired               = errors.New("code expired")
	ErrOAuthAccessTokenExpired        = errors.New("access token expired")
	ErrOAuthRefreshTokenExpired       = errors.New("refresh token expired")
	ErrOAuthReqNotValid               = errors.New("oauth request not valid")
	ErrOAuthAuthorizationCodeNotExist = errors.New("authorization code not exist")
	ErrOAuthRefreshTokenNotExist      = errors.New("refresh token not exist")
	ErrOAuthInternal                  = errors.New("oauth internal error")
	ErrAuthorizationHeaderNotFound    = errors.New("AuthorizationHeader not found")
	ErrOAuthTokenFormatError          = errors.New("Oauth token format error")
	ErrOAuthNotGroupOwnerType         = errors.New("not group oauth app")

	// ErrRegistryUsedByRegions used when deleting a registry that is still used by regions
	ErrRegistryUsedByRegions = errors.New("cannot delete a registry when used by regions")

	// ErrRegionUsedByClusters used when deleting a region that is still used by clusters
	ErrRegionUsedByClusters        = errors.New("cannot delete a region when used by clusters")
	ErrPipelineOutPut              = errors.New("pipeline output is not valid")
	ErrTemplateParamInvalid        = errors.New("parameters of template are invalid")
	ErrTemplateReleaseParamInvalid = errors.New("parameters of release are invalid")
	ErrAPIServerResponseNotOK      = errors.New("response for api-server is not 200 OK")
	ErrListGrafanaDashboard        = errors.New("List grafana dashboards error")

	// event
	ErrEventHandlerAlreadyExist = errors.New("event handler already exist")

	ErrSessionNotFound   = errors.New("session not found")
	ErrSessionSaveFailed = errors.New("failed to save session")

	ErrMethodNotImplemented = errors.New("method not implemented")
	ErrTopResourceNotFound  = errors.New("top resource in resource tree not found")

	ErrNotSupport = errors.New("not support")

	// token
	ErrTokenInvalid = errors.New("token is invalid")
)
