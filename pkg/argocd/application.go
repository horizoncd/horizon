package argocd

type Application struct {
	APIVersion string              `json:"apiVersion" yaml:"apiVersion"`
	Kind       string              `json:"kind" yaml:"kind"`
	Metadata   ApplicationMetadata `json:"metadata" yaml:"metadata"`
	Spec       ApplicationSpec     `json:"spec" yaml:"spec"`
}

type ApplicationMetadata struct {
	Finalizers []string `json:"finalizers" yaml:"finalizers"`
	Name       string   `json:"name" yaml:"name"`
	Namespace  string   `json:"namespace" yaml:"namespace"`
}

// ApplicationSpec represents desired application state.
// Contains link to repository with application definition and additional parameters link definition revision.
type ApplicationSpec struct {
	// Source is a reference to the location ksonnet application definition
	Source ApplicationSource `json:"source" yaml:"source"`
	// Destination overrides the kubernetes server and namespace defined in the environment ksonnet app.yaml
	Destination ApplicationDestination `json:"destination" yaml:"destination"`
	// Project is a application project name. Empty name means that application belongs to 'default' project.
	Project string `json:"project" yaml:"project"`
	// SyncPolicy controls when a sync will be performed
	SyncPolicy *SyncPolicy `json:"syncPolicy" yaml:"syncPolicy,omitempty"`
}

// ApplicationSource contains information about github repository,
// path within repository and target application environment.
type ApplicationSource struct {
	// RepoURL is the repository GitlabHTTPS of the application manifests
	RepoURL string `json:"repoURL" yaml:"repoURL"`
	// Path is a directory path within the Git repository
	Path string `json:"path" yaml:"path,omitempty"`
	// TargetRevision defines the commit, tag, or branch in which to sync the application to.
	// If omitted, will sync to HEAD
	TargetRevision string `json:"targetRevision" yaml:"targetRevision,omitempty"`
	// Helm holds helm specific options
	Helm *ApplicationSourceHelm `json:"helm" yaml:"helm,omitempty"`
}

// ApplicationSourceHelm holds helm specific options
type ApplicationSourceHelm struct {
	// ValuesFiles is a list of Helm value files to use when generating a template
	ValueFiles []string `json:"valueFiles" yaml:"valueFiles,omitempty"`
}

// SyncPolicy controls when a sync will be performed in response to updates in git
type SyncPolicy struct {
	// Options allow you to specify whole app sync-options
	SyncOptions SyncOptions `json:"syncOptions" yaml:"syncOptions,omitempty"`
}

// ApplicationDestination contains deployment destination information
type ApplicationDestination struct {
	// Server overrides the environment server value in the ksonnet app.yaml
	Server string `json:"server" yaml:"server,omitempty"`
	// Namespace overrides the environment namespace value in the ksonnet app.yaml
	Namespace string `json:"namespace" yaml:"namespace,omitempty"`
}

type SyncOptions []string
