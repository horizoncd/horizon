package git

type Repo struct {
	Kind  string `yaml:"kind"`
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}
