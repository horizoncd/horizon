package cmdb

type Config struct {
	Enabled    bool   `yaml:"enabled"`
	URL        string `yaml:"url"`
	ClientID   string `yaml:"clientID"`
	SecretCode string `yaml:"secretCode"`
	ParentID   int    `yaml:"parentID"`
}
