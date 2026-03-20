package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/dealance/services/startup/internal/domain/entity"
)

const (
	startupCacheTTL    = 15 * time.Minute
	startupCachePrefix = "startup:profile:"
)

type CacheRepo struct{ rdb *redis.Client }

func NewCacheRepo(rdb *redis.Client) *CacheRepo { return &CacheRepo{rdb: rdb} }

func (r *CacheRepo) CacheStartup(ctx context.Context, startup *entity.Startup) error {
	data, err := json.Marshal(startup)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, startupCachePrefix+startup.ID.String(), data, startupCacheTTL).Err()
}

func (r *CacheRepo) GetCachedStartup(ctx context.Context, id string) (*entity.Startup, error) {
	data, err := r.rdb.Get(ctx, startupCachePrefix+id).Bytes()
	if err != nil {
		return nil, err
	}
	var startup entity.Startup
	return &startup, json.Unmarshal(data, &startup)
}

func (r *CacheRepo) InvalidateStartup(ctx context.Context, id string) error {
	return r.rdb.Del(ctx, startupCachePrefix+id).Err()
}
