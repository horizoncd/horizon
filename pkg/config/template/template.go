package template

type UpgradeMapper map[string]TargetTemplate

type TargetTemplate struct {
	Name        string      `json:"name" yaml:"name"`
	Release     string      `json:"release" yaml:"release"`
	BuildConfig BuildConfig `json:"buildConfig" yaml:"buildConfig"`
}

type BuildConfig struct {
	Language    string `json:"language" yaml:"language"`
	Environment string `json:"environment" yaml:"environment"`
}
