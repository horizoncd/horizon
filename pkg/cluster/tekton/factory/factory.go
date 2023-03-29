package factory

import (
	"sync"

	"github.com/horizoncd/horizon/lib/s3"
	"github.com/horizoncd/horizon/pkg/cluster/tekton"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/collector"
	tektonconfig "github.com/horizoncd/horizon/pkg/config/tekton"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	"github.com/horizoncd/horizon/pkg/util/errors"
)

const (
	_default   = "default"
	_s3Storage = "s3"
)

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
		var c collector.Interface
		if tektonConfig.LogStorage.Type == _s3Storage {
			s3Driver, err := s3.NewDriver(s3.Params{
				AccessKey:        tektonConfig.LogStorage.AccessKey,
				SecretKey:        tektonConfig.LogStorage.SecretKey,
				Region:           tektonConfig.LogStorage.Region,
				Endpoint:         tektonConfig.LogStorage.Endpoint,
				Bucket:           tektonConfig.LogStorage.Bucket,
				DisableSSL:       tektonConfig.LogStorage.DisableSSL,
				SkipVerify:       tektonConfig.LogStorage.SkipVerify,
				S3ForcePathStyle: tektonConfig.LogStorage.S3ForcePathStyle,
				ContentType:      "text/plain",
			})
			if err != nil {
				return nil, errors.E(op, err)
			}
			c = collector.NewS3Collector(s3Driver, t)
		} else {
			c = collector.NewDummyCollector(t)
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
