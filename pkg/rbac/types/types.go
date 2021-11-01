package types

// attention: rbac is refers to the kubernetes rbac
// copy core struct and logics from the kubernetes code
// and do same modify
const (
	APIGroupAll    = "*"
	ResourceAll    = "*"
	VerbAll        = "*"
	ScopeAll       = "*"
	NonResourceAll = "*"
)

type Role struct {
	Name        string       `yaml:"name"`
	PolicyRules []PolicyRule `yaml:"rules"`
}

type PolicyRule struct {
	Verbs           []string `yaml:"verbs"`
	APIGroups       []string `yaml:"apiGroups"`
	Resources       []string `yaml:"resources"`
	Scopes          []string `yaml:"scopes"`
	NonResourceURLs []string `yaml:"nonResourceURLs"`
}
