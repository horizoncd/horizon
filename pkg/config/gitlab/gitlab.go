package gitlab

// Mapper gitlab mapper
type Mapper map[string]*Gitlab

type Gitlab struct {
	HTTPURL string `yaml:"httpURL"`
	SSHURL  string `yaml:"sshURL"`
	Token   string `yaml:"token"`
}

// GitopsRepoConfig gitops repo config
type GitopsRepoConfig struct {
	RootGroupPath string `yaml:"rootGroupPath"`
}
