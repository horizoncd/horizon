package session

import (
	"context"
	"time"

	"g.hz.netease.com/horizon/lib/log"
	"g.hz.netease.com/horizon/lib/redis"
)

const expiration = 2 * time.Hour

type RedisSession struct {
	redisHelper *redis.Helper
}

func NewRedisSession(redisHelper *redis.Helper) *RedisSession {
	return &RedisSession{redisHelper: redisHelper}
}

func (r *RedisSession) GetSession(ctx context.Context, sessionID string) (_ *Session, err error) {
	defer log.TRACE.Debug(ctx)(func() error { return err })
	logger := log.GetLogger(ctx)
	logger.Debugf("get session with sessionID %s", sessionID)

	var session *Session
	if err = r.redisHelper.Get(ctx, sessionID, &session); err != nil {
		// 找不到session，认为正常
		if err == redis.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}

func (r *RedisSession) SetSession(ctx context.Context, sessionID string, session *Session) (err error) {
	defer log.TRACE.Debug(ctx)(func() error { return err })
	logger := log.GetLogger(ctx)
	logger.Debugf("set session with sessionID %s, session: %v", sessionID, session)

	return r.redisHelper.Save(ctx, sessionID, session, expiration)
}

func (r *RedisSession) DeleteSession(ctx context.Context, sessionID string) (err error) {
	defer log.TRACE.Debug(ctx)(func() error { return err })
	logger := log.GetLogger(ctx)
	logger.Debugf("delete session with sessionID %s", sessionID)

	return r.redisHelper.Delete(ctx, sessionID)
}
