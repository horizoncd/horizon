package factory

import (
	"fmt"
	"sync"

	"g.hz.netease.com/horizon/lib/s3"
	"g.hz.netease.com/horizon/pkg/cluster/tekton"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/collector"
	tektonconfig "g.hz.netease.com/horizon/pkg/config/tekton"
	"g.hz.netease.com/horizon/pkg/util/errors"
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
		s3Driver, err := s3.NewDriver(s3.Params{
			AccessKey:        tektonConfig.S3.AccessKey,
			SecretKey:        tektonConfig.S3.SecretKey,
			Region:           tektonConfig.S3.Region,
			Endpoint:         tektonConfig.S3.Endpoint,
			Bucket:           tektonConfig.S3.Bucket,
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
	var ret interface{}
	var ok bool
	if ret, ok = f.cache.Load(environment); !ok {
		return nil, fmt.Errorf("the tekton for environment: %v is not found", environment)
	}
	return ret.(*tektonCache).tekton, nil
}

func (f factory) GetTektonCollector(environment string) (collector.Interface, error) {
	var ret interface{}
	var ok bool
	if ret, ok = f.cache.Load(environment); !ok {
		return nil, fmt.Errorf("the tektonCollector for environment: %v is not found", environment)
	}
	return ret.(*tektonCache).tektonCollector, nil
}
