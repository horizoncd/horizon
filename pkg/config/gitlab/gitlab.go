package gitlab

const (
	HTTPURLSchema = "http"
	SSHURLSchema  = "ssh"
)

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
	URLSchema     string `yaml:"urlSchema"`
}
