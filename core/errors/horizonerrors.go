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
	Cluster                   = sourceType{name: "Cluster"}
	ClusterInDB               = sourceType{name: "ClusterInDB"}
	ClusterStateInArgo        = sourceType{name: "ClusterStateInArgo"}
	ClusterTagInDB            = sourceType{name: "ClusterTagInDB"}
	Application               = sourceType{name: "Application"}
	ApplicationInArgo         = sourceType{name: "ApplicationInArgo"}
	ApplicationResourceInArgo = sourceType{name: "ApplicationResourceInArgo"}
	ApplicationInDB           = sourceType{name: "ApplicationInDB"}
	EnvironmentRegion         = sourceType{name: "EnvironmentRegion"}
	EnvironmentRegionInDB     = sourceType{name: "EnvironmentRegionInDB"}
	Region                    = sourceType{name: "Region"}
	RegionInDB                = sourceType{name: "RegionInDB"}
	Group                     = sourceType{name: "Group"}
	GroupInDB                 = sourceType{name: "GroupInDB"}
	K8SCluster                = sourceType{name: "K8SCluster"}
	K8SClusterInDB            = sourceType{name: "K8SClusterInDB"}
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
	source sourceType
}

func NewErrGetFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrGetFailed{
		source: source,
	}, msg)
}

func (e *HorizonErrGetFailed) Error() string {
	return fmt.Sprintf("%s get failed", e.source.name)
}

type HorizonErrDeleteFailed struct {
	source sourceType
}

func NewErrDeleteFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrDeleteFailed{
		source: source,
	}, msg)
}

func (e *HorizonErrDeleteFailed) Error() string {
	return fmt.Sprintf("%s delete failed", e.source.name)
}

type HorizonErrUpdateFailed struct {
	source sourceType
}

func NewErrUpdateFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrUpdateFailed{
		source: source,
	}, msg)
}

func (e *HorizonErrUpdateFailed) Error() string {
	return fmt.Sprintf("%s update failed", e.source.name)
}

type HorizonErrInsertFailed struct {
	source sourceType
}

func NewErrInsertFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrInsertFailed{
		source: source,
	}, msg)
}

func (e *HorizonErrInsertFailed) Error() string {
	return fmt.Sprintf("%s insert failed", e.source.name)
}

type HorizonErrCreateFailed struct {
	source sourceType
}

func NewErrCreateFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrCreateFailed{
		source: source,
	}, msg)
}

func (e *HorizonErrCreateFailed) Error() string {
	return fmt.Sprintf("%s create failed", e.source.name)
}

type HorizonErrListFailed struct {
	source sourceType
}

func NewErrListFailed(source sourceType, msg string) error {
	return errors.Wrap(&HorizonErrListFailed{
		source: source,
	}, msg)
}

func (e *HorizonErrListFailed) Error() string {
	return fmt.Sprintf("%s list failed", e.source.name)
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

	ErrK8SClusterNotList = errors.New("k8s cluster could not list")
	ErrHarborNotList     = errors.New("harbor could not list")

	ErrKubeDynamicCliResponseNotOK = errors.New("response for kube dynamic cli is not 200 OK")
	ErrKubeExecFailed              = errors.New("kube exec failed")

	// S3
	ErrS3SignFailed   = errors.New("s3 sign failed")
	ErrS3PutObjFailed = errors.New("s3 put obj failed")
	ErrS3GetObjFailed = errors.New("s3 get obj failed")

	ErrGitlabResourceNotFound = errors.New("gitlab resource not found")
	ErrGitlabInternal         = errors.New("gitlab internal")

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
