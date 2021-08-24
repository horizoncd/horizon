package redis


import (
	"time"
)

// Option function to set the options of the redis
type Option func(*Options)

// Options options of the redis
type Options struct {
	Codec      Codec         // the codec for redis data
	Expiration time.Duration // the default expiration for the redis
	Prefix     string        // the prefix for all the keys in the redis data
}

// Key returns the real cache key
func (opts *Options) Key(key string) string {
	return opts.Prefix + key
}

func NewOptionsWithDefaultCodec(opt ...Option) *Options {
	opts := Options{
		Codec: DefaultCodec,
	}

	for _, o := range opt {
		o(&opts)
	}

	return &opts
}

// Expiration sets the default expiration
func Expiration(d time.Duration) Option {
	return func(o *Options) {
		o.Expiration = d
	}
}

// Prefix sets the prefix
func Prefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}
