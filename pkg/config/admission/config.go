package admission

import (
	"strings"
	"time"

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
	Kind          models.Kind   `yaml:"kind"`
	FailurePolicy FailurePolicy `yaml:"failurePolicy"`
	Timeout       time.Duration `yaml:"timeout"`
	Rules         []Rule        `yaml:"rules"`
	ClientConfig  ClientConfig  `yaml:"clientConfig"`
}

type Admission struct {
	Webhooks []Webhook `yaml:"webhooks"`
}
