package oidc

type Config struct {
	UserHeader     string `yaml:"userHeader"`
	FullNameHeader string `yaml:"fullNameHeader"`
	EmailHeader    string `yaml:"emailHeader"`
	OIDCIDHeader   string `yaml:"oidcIDHeader"`
	OIDCTypeHeader string `yaml:"oidcTypeHeader"`
}
