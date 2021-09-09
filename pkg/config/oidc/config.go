package oidc

type Config struct {
	UserHeader     string `yaml:"userHeader"`
	EmailHeader    string `yaml:"emailHeader"`
	OIDCIDHeader   string `yaml:"oidcIDHeader"`
	OIDCTypeHeader string `yaml:"oidcTypeHeader"`
}
