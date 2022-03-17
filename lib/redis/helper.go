package redis

import (
	"context"
	"fmt"
	"time"

	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"github.com/gomodule/redigo/redis"
)

var (
	ErrNotFound = perror.New("key not found")
)

type Interface interface {
	Ping(ctx context.Context) error
	Get(ctx context.Context, key string, value interface{}) error
	Save(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error
	Delete(ctx context.Context, key string) error
}

type Helper struct {
	pool *redis.Pool
	opts *Options
}

func NewHelper(redisURL, poolName string, param *PoolParam, opts *Options) (*Helper, error) {
	pool, err := GetRedisPool(poolName, redisURL, param)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		return nil, perror.New("opts cannot be nil")
	}
	return &Helper{
		pool: pool,
		opts: opts,
	}, nil
}

func NewHelperWithPool(pool *redis.Pool, opts *Options) (*Helper, error) {
	if pool == nil {
		return nil, perror.New("pool cannot be nil")
	}
	if opts == nil {
		return nil, perror.New("opts cannot be nil")
	}
	return &Helper{
		pool: pool,
		opts: opts,
	}, nil
}

func (h *Helper) Ping(ctx context.Context) (err error) {
	const op = "redis helper: ping"
	defer wlog.Start(ctx, op).StopPrint()

	_, err = h.do(ctx, "PING")
	return err
}

func (h *Helper) Get(ctx context.Context, key string, value interface{}) (err error) {
	const op = "redis helper: get"
	defer wlog.Start(ctx, op).StopPrint()

	data, err := redis.Bytes(h.do(ctx, "GET", h.opts.Key(key)))
	if err != nil {
		// convert internal or Timeout error to be ErrNotFound
		// so that the caller can continue working without breaking
		return ErrNotFound
	}

	if err := h.opts.Codec.Decode(data, value); err != nil {
		return fmt.Errorf("failed to decode redis data value to dest, key %s, error: %v", key, err)
	}

	return nil
}

func (h *Helper) Save(ctx context.Context, key string, value interface{}, expiration ...time.Duration) (err error) {
	const op = "redis helper: save"
	defer wlog.Start(ctx, op).StopPrint()

	data, err := h.opts.Codec.Encode(value)
	if err != nil {
		return fmt.Errorf("failed to encode value, key %s, error: %v", key, err)
	}

	args := []interface{}{h.opts.Key(key), data}

	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	} else if h.opts.Expiration > 0 {
		exp = h.opts.Expiration
	}

	if exp > 0 {
		args = append(args, "EX", int64(exp/time.Second))
	}

	_, err = h.do(ctx, "SET", args...)
	return err
}

func (h *Helper) Delete(ctx context.Context, key string) (err error) {
	const op = "redis helper: delete"
	defer wlog.Start(ctx, op).StopPrint()

	_, err = h.do(ctx, "DEL", h.opts.Key(key))
	return err
}

func (h *Helper) do(ctx context.Context, command string, args ...interface{}) (reply interface{}, err error) {
	conn := h.pool.Get()
	defer func() { _ = conn.Close() }()

	return conn.Do(command, args...)
}
