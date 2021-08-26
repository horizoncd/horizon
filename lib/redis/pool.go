package redis

import (
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	knownPool sync.Map
	m         sync.Mutex
)

// PoolParam redis pool params
type PoolParam struct {
	PoolMaxIdle     int
	PoolMaxActive   int
	PoolIdleTimeout time.Duration

	DialConnectionTimeout time.Duration
	DialReadTimeout       time.Duration
	DialWriteTimeout      time.Duration
}

// GetRedisPool get a named redis pool
// supported rawURL
// redis://user:pass@redis_host:port/db
func GetRedisPool(name string, rawURL string, param *PoolParam) (*redis.Pool, error) {
	if p, ok := knownPool.Load(name); ok {
		return p.(*redis.Pool), nil
	}
	m.Lock()
	defer m.Unlock()
	// load again in case multi threads
	if p, ok := knownPool.Load(name); ok {
		return p.(*redis.Pool), nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("bad redis url: %s, %s, %s", name, rawURL, err)
	}

	if param == nil {
		param = &PoolParam{
			PoolMaxIdle:           3,
			PoolMaxActive:         10,
			PoolIdleTimeout:       1 * time.Minute,
			DialConnectionTimeout: 3 * time.Second,
			DialReadTimeout:       3 * time.Second,
			DialWriteTimeout:      3 * time.Second,
		}
	}
	if t := u.Query().Get("idle_timeout_seconds"); t != "" {
		if tt, e := strconv.Atoi(t); e == nil {
			param.PoolIdleTimeout = time.Second * time.Duration(tt)
		}
	}

	// log.Debug("get redis pool:", name, rawURL)
	if u.Scheme == "redis" {
		pool := &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return redis.DialURL(rawURL)
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
			MaxIdle:     param.PoolMaxIdle,
			MaxActive:   param.PoolMaxActive,
			IdleTimeout: param.PoolIdleTimeout,
			Wait:        true,
		}
		knownPool.Store(name, pool)
		return pool, nil
	} else {
		return nil, fmt.Errorf("bad redis url: not support scheme %s", u.Scheme)
	}
}
