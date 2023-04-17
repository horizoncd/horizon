package gitlab

// GitopsRepoConfig gitops repo config.
type GitopsRepoConfig struct {
	URL           string `yaml:"url"`
	Token         string `yaml:"token"`
	RootGroupPath string `yaml:"rootGroupPath"`
}
