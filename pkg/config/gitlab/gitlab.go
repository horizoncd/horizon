package gitlab

// Mapper gitlab mapper
type Mapper map[string]*Gitlab

type Gitlab struct {
	HTTPURL string `yaml:"httpURL"`
	SSHURL  string `yaml:"sshURL"`
	Token   string `yaml:"token"`
}

// RepoConfig gitlab repo config
type RepoConfig struct {
	Application *Repo `yaml:"application"`
	Cluster     *Repo `yaml:"cluster"`
}

type Repo struct {
	Parent          *Parent `yaml:"parent"`
	RecyclingParent *Parent `yaml:"recyclingParent"`
}

type Parent struct {
	Path string `yaml:"path"`
	ID   int    `yaml:"id"`
}
