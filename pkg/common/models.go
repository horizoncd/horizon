package common

type Manifest struct {
	// TODO(encode the template info into manifest),currently only the Version
	Version string `yaml:"version" json:"version"`
}
