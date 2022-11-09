package templaterepo

type Repo struct {
	Kind     string `yaml:"kind"`
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Token    string `yaml:"token"`
	Insecure bool   `yaml:"insecure"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
	CAFile   string `yaml:"caFile"`
	RepoName string `yaml:"repoName"`
}
