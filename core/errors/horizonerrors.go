package errors

import (
	"fmt"

	"g.hz.netease.com/horizon/pkg/errors"
)

type sourceType struct {
	name string
}

var (
	GitlabResource            = sourceType{name: "GitlabResource"}
	ClusterInDB               = sourceType{name: "ClusterInDB"}
	ClusterStateInArgo        = sourceType{name: "ClusterStateInArgo"}
	ClusterTagInDB            = sourceType{name: "ClusterTagInDB"}
	ApplicationInArgo         = sourceType{name: "ApplicationInArgo"}
	ApplicationResourceInArgo = sourceType{name: "ApplicationResourceInArgo"}
	ApplicationInDB           = sourceType{name: "ApplicationInDB"}
	EnvironmentRegionInDB     = sourceType{name: "EnvironmentRegionInDB"}
	RegionInDB                = sourceType{name: "RegionInDB"}
	GroupInDB                 = sourceType{name: "GroupInDB"}
	K8SClient                 = sourceType{name: "K8SClient"}
	Harbor                    = sourceType{name: "Harbor"}
	HarborInDB                = sourceType{name: "HarborInDB"}
	Pipelinerun               = sourceType{name: "Pipelinerun"}
	PipelinerunInTekton       = sourceType{name: "PipelinerunInTekton"}
	PipelinerunInDB           = sourceType{name: "PipelinerunInDB"}
	UserInDB                  = sourceType{name: "UserInDB"}
	TemplateSchemaTagInDB     = sourceType{name: "TemplateSchemaTagInDB"}
	TemplateReleaseInDB       = sourceType{name: "TemplateReleaseInDB"}
	ApplicationManifestInArgo = sourceType{name: "ApplicationManifestInArgo"}
	PodsInK8S                 = sourceType{name: "PodsInK8S"}
	ReplicasSetInK8S          = sourceType{name: "ReplicasSetInK8S"}
	DeploymentInK8S           = sourceType{name: "DeploymentInK8S"}
	PodEventInK8S             = sourceType{name: "PodEventInK8S"}
	KubeConfigInK8S           = sourceType{name: "KubeConfigK8S"}

	// S3
	PipelinerunLog = sourceType{name: "PipelinerunLog"}
	PipelinerunObj = sourceType{name: "PipelinerunObj"}

	ArgoCD = sourceType{name: "ArgoCD"}

	Tekton          = sourceType{name: "Tekton"}
	TektonClient    = sourceType{name: "TektonClient"}
	TektonCollector = sourceType{name: "TektonCollector"}

	HelmRepo = sourceType{name: "HelmRepo"}
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
	ErrParamInvalid     = errors.New("parameter is invalid")
	ErrDeadlineExceeded = errors.New("time limit exceeded")
	ErrFailedToRollback = errors.New("failed to rollback")
	ErrGenerateRandomID = errors.New("failed to generate random id")
	// ErrInternal = errors.New("internal error")

	// http
	ErrHTTPRespNotAsExpected = errors.New("http response is not as expected")
	ErrHTTPRequestFailed     = errors.New("http request failed")

	// cluster
	ErrClusterNoChange = errors.New("no change to cluster")

	// pipelinerun

	// context
	ErrFailedToGetORM       = errors.New("cannot get the ORM from context")
	ErrFailedToGetUser      = errors.New("cannot get user from context")
	ErrFailedToGetRequestID = errors.New("cannot get the requestID from context")

	ErrHarborNotList = errors.New("harbor could not list")

	ErrKubeDynamicCliResponseNotOK = errors.New("response for kube dynamic cli is not 200 OK")
	ErrKubeExecFailed              = errors.New("kube exec failed")

	// S3
	ErrS3SignFailed   = errors.New("s3 sign failed")
	ErrS3PutObjFailed = errors.New("s3 put obj failed")
	ErrS3GetObjFailed = errors.New("s3 get obj failed")

	ErrGitlabResourceNotFound = errors.New("gitlab resource not found")
	ErrGitlabInternal         = errors.New("gitlab internal")
	ErrGitlabMRNotReady       = errors.New("gitlab mr is not ready and cannot be merged")

	// git
	ErrBranchAndCommitEmpty = errors.New("branch and commit cannot be empty at the same time")

	// pipeline
	ErrPipelineOutputEmpty = errors.New("pipeline output is empty")

	// tekton
	ErrTektonInternal = errors.New("tekton internal error")

	// grafana
	ErrGrafanaNotSupport = errors.New("grafana not support")

	// group
	// ErrHasChildren used when delete a group which still has some children
	ErrGroupHasChildren = errors.New("children exist, cannot be deleted")
	// ErrConflictWithApplication conflict with the application
	ErrGroupConflictWithApplication = errors.New("name or path is in conflict with application")
)
