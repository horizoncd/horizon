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
	Name        string       `yaml:"name" json:"name"`
	Desc        string       `yaml:"desc" json:"desc"`
	PolicyRules []PolicyRule `yaml:"rules" json:"rules"`
}

type PolicyRule struct {
	Verbs           []string `yaml:"verbs" json:"verbs"`
	APIGroups       []string `yaml:"apiGroups" json:"apiGroups"`
	Resources       []string `yaml:"resources" json:"resources"`
	Scopes          []string `yaml:"scopes" json:"scopes"`
	NonResourceURLs []string `yaml:"nonResourceURLs" json:"nonResourceURLs"`
}
