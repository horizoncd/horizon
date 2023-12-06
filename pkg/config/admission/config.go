package admission

import (
	"strings"

	"github.com/horizoncd/horizon/pkg/admission/models"
)

type FailurePolicy string

func (f FailurePolicy) Eq(other FailurePolicy) bool {
	return strings.EqualFold(string(f), string(other))
}

const (
	FailurePolicyIgnore FailurePolicy = "ignore"
	FailurePolicyFail   FailurePolicy = "fail"
)

type ClientConfig struct {
	URL      string `yaml:"url"`
	CABundle string `yaml:"caBundle"`
	Insecure bool   `yaml:"insecure"`
}

type Rule struct {
	Resources  []string           `yaml:"resources"`
	Operations []models.Operation `yaml:"operations"`
	Versions   []string           `yaml:"versions"`
}

type Webhook struct {
	Name           string        `yaml:"name"`
	Kind           models.Kind   `yaml:"kind"`
	FailurePolicy  FailurePolicy `yaml:"failurePolicy"`
	TimeoutSeconds int32         `yaml:"timeoutSeconds"`
	Rules          []Rule        `yaml:"rules"`
	ClientConfig   ClientConfig  `yaml:"clientConfig"`
}

type Admission struct {
	Webhooks []Webhook `yaml:"webhooks"`
}
