package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/dealance/services/content/internal/domain/entity"
)

const (
	reactionCountPrefix  = "content:reactions:"
	trendingHashtagsKey  = "content:trending_hashtags"
	cacheTTL             = 10 * time.Minute
)

type CacheRepo struct{ rdb *redis.Client }

func NewCacheRepo(rdb *redis.Client) *CacheRepo { return &CacheRepo{rdb: rdb} }

func (r *CacheRepo) CacheReactionCount(ctx context.Context, targetID, targetType string, count int) error {
	key := fmt.Sprintf("%s%s:%s", reactionCountPrefix, targetType, targetID)
	return r.rdb.Set(ctx, key, count, cacheTTL).Err()
}

func (r *CacheRepo) GetReactionCount(ctx context.Context, targetID, targetType string) (int, error) {
	key := fmt.Sprintf("%s%s:%s", reactionCountPrefix, targetType, targetID)
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

func (r *CacheRepo) IncrReactionCount(ctx context.Context, targetID, targetType string, delta int) error {
	key := fmt.Sprintf("%s%s:%s", reactionCountPrefix, targetType, targetID)
	return r.rdb.IncrBy(ctx, key, int64(delta)).Err()
}

func (r *CacheRepo) CacheTrendingHashtags(ctx context.Context, hashtags []entity.Hashtag) error {
	data, err := json.Marshal(hashtags)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, trendingHashtagsKey, data, 30*time.Minute).Err()
}

func (r *CacheRepo) GetTrendingHashtags(ctx context.Context) ([]entity.Hashtag, error) {
	data, err := r.rdb.Get(ctx, trendingHashtagsKey).Bytes()
	if err != nil {
		return nil, err
	}
	var hashtags []entity.Hashtag
	return hashtags, json.Unmarshal(data, &hashtags)
}
