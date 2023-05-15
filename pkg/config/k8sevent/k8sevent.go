package k8sevent

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Reason struct {
	Reason   string   `yaml:"reason"`
	Messages []string `yaml:"messages"`
}

type Rule struct {
	schema.GroupVersionKind `yaml:",inline"`
	Reasons                 []Reason `yaml:"reasons"`
}

type Config struct {
	Rules []Rule `yaml:"rules"`
}
