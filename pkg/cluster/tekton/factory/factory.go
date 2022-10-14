package factory

import (
	"sync"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/s3"
	"g.hz.netease.com/horizon/pkg/cluster/tekton"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/collector"
	tektonconfig "g.hz.netease.com/horizon/pkg/config/tekton"
	"g.hz.netease.com/horizon/pkg/util/errors"
)

const _default = "default"

type Factory interface {
	GetTekton(environment string) (tekton.Interface, error)
	GetTektonCollector(environment string) (collector.Interface, error)
}

type factory struct {
	cache *sync.Map
}

type tektonCache struct {
	tekton          tekton.Interface
	tektonCollector collector.Interface
}

func NewFactory(tektonMapper tektonconfig.Mapper) (Factory, error) {
	const op = "new tekton factory"

	cache := &sync.Map{}
	for env, tektonConfig := range tektonMapper {
		t, err := tekton.NewTekton(tektonConfig)
		if err != nil {
			return nil, errors.E(op, err)
		}
		s3Driver, err := s3.NewDriver(s3.Params{
			AccessKey:        tektonConfig.S3.AccessKey,
			SecretKey:        tektonConfig.S3.SecretKey,
			Region:           tektonConfig.S3.Region,
			Endpoint:         tektonConfig.S3.Endpoint,
			Bucket:           tektonConfig.S3.Bucket,
			DisableSSL:       tektonConfig.S3.DisableSSL,
			SkipVerify:       tektonConfig.S3.SkipVerify,
			S3ForcePathStyle: tektonConfig.S3.S3ForcePathStyle,
			ContentType:      "text/plain",
		})
		c := collector.NewS3Collector(s3Driver, t)
		if err != nil {
			return nil, errors.E(op, err)
		}
		cache.Store(env, &tektonCache{
			tekton:          t,
			tektonCollector: c,
		})
	}
	return &factory{
		cache: cache,
	}, nil
}

func (f factory) GetTekton(environment string) (tekton.Interface, error) {
	cache, err := f.GetFromCache(environment)
	if err != nil {
		return nil, err
	}
	return cache.tekton, nil
}

func (f factory) GetTektonCollector(environment string) (collector.Interface, error) {
	cache, err := f.GetFromCache(environment)
	if err != nil {
		return nil, err
	}
	return cache.tektonCollector, nil
}

func (f factory) GetFromCache(environment string) (*tektonCache, error) {
	var ret interface{}
	var ok bool
	if ret, ok = f.cache.Load(environment); !ok {
		// check and use default tekton
		if ret, ok = f.cache.Load(_default); !ok {
			return nil, herrors.NewErrNotFound(herrors.Tekton, "default tekton not found")
		}
	}
	return ret.(*tektonCache), nil
}
