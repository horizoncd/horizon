package templaterepo

type Repo struct {
	Host      string `yaml:"host"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Token     string `yaml:"token"`
	Scheme    string `yaml:"scheme"`
	PlainHTTP bool   `yaml:"plainHTTP"`
	Insecure  bool   `yaml:"insecure"`
	CertFile  string `yaml:"certFile"`
	KeyFile   string `yaml:"keyFile"`
	CAFile    string `yaml:"caFile"`
	RepoName  string `yaml:"repoName"`
}
