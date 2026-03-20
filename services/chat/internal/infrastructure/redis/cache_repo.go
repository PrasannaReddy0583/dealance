package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	onlineTTL    = 5 * time.Minute
	onlinePrefix = "chat:online:"
	channelPrefix = "chat:conv:"
)

type CacheRepo struct{ rdb *redis.Client }
func NewCacheRepo(rdb *redis.Client) *CacheRepo { return &CacheRepo{rdb: rdb} }

func (r *CacheRepo) SetUserOnline(ctx context.Context, userID string) error {
	return r.rdb.Set(ctx, onlinePrefix+userID, "1", onlineTTL).Err()
}
func (r *CacheRepo) SetUserOffline(ctx context.Context, userID string) error {
	return r.rdb.Del(ctx, onlinePrefix+userID).Err()
}
func (r *CacheRepo) IsOnline(ctx context.Context, userID string) (bool, error) {
	val, err := r.rdb.Exists(ctx, onlinePrefix+userID).Result()
	return val > 0, err
}
func (r *CacheRepo) PublishMessage(ctx context.Context, convID string, payload []byte) error {
	return r.rdb.Publish(ctx, channelPrefix+convID, payload).Err()
}
