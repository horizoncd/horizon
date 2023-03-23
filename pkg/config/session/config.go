package session

type Config struct {
	MaxAge    uint32 `yaml:"maxAge"`
	StoreType string `yaml:"storeType"`
}
