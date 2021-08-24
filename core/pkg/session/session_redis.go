package session

import (
	"time"

	"g.hz.netease.com/horizon/pkg/redis"
)

const expiration = 2 * time.Hour

type RedisSession struct {
	redisHelper *redis.Helper
}

func NewRedisSession(redisHelper *redis.Helper) *RedisSession {
	return &RedisSession{redisHelper: redisHelper}
}

func (r *RedisSession) GetSession(sessionID string) (*Session, error) {
	var session *Session
	if err := r.redisHelper.Get(sessionID, &session); err != nil {
		// 找不到session，认为正常
		if err == redis.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}

func (r *RedisSession) SetSession(sessionID string, session *Session) error {
	return r.redisHelper.Save(sessionID, session, expiration)
}

func (r *RedisSession) DeleteSession(sessionID string) error {
	return r.redisHelper.Delete(sessionID)
}
