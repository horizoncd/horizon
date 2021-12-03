package authenticate

type Key struct {
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
}

type Keys []*Key

type KeysConfig map[string]Keys
