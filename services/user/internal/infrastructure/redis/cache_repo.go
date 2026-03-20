package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/dealance/services/user/internal/domain/entity"
)

const (
	profileCacheTTL = 15 * time.Minute
	profileKeyPrefix = "user:profile:"
	followCountKeyPrefix = "user:follows:"
)

// CacheRepo implements CacheRepository using Redis.
type CacheRepo struct {
	rdb *redis.Client
}

func NewCacheRepo(rdb *redis.Client) *CacheRepo {
	return &CacheRepo{rdb: rdb}
}

func (r *CacheRepo) CacheProfile(ctx context.Context, profile *entity.Profile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	key := profileKeyPrefix + profile.ID.String()
	return r.rdb.Set(ctx, key, data, profileCacheTTL).Err()
}

func (r *CacheRepo) GetCachedProfile(ctx context.Context, userID string) (*entity.Profile, error) {
	key := profileKeyPrefix + userID
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var profile entity.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *CacheRepo) InvalidateProfile(ctx context.Context, userID string) error {
	return r.rdb.Del(ctx, profileKeyPrefix+userID).Err()
}

func (r *CacheRepo) CacheFollowCounts(ctx context.Context, userID string, followers, following int) error {
	key := followCountKeyPrefix + userID
	pipe := r.rdb.Pipeline()
	pipe.HSet(ctx, key, "followers", followers, "following", following)
	pipe.Expire(ctx, key, profileCacheTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *CacheRepo) GetFollowCounts(ctx context.Context, userID string) (int, int, error) {
	key := followCountKeyPrefix + userID
	result, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil || len(result) == 0 {
		return 0, 0, fmt.Errorf("cache miss")
	}
	followers, _ := strconv.Atoi(result["followers"])
	following, _ := strconv.Atoi(result["following"])
	return followers, following, nil
}

func (r *CacheRepo) IncrFollowCount(ctx context.Context, userID string, delta int) error {
	key := followCountKeyPrefix + userID
	return r.rdb.HIncrBy(ctx, key, "followers", int64(delta)).Err()
}

func (r *CacheRepo) IncrFollowingCount(ctx context.Context, userID string, delta int) error {
	key := followCountKeyPrefix + userID
	return r.rdb.HIncrBy(ctx, key, "following", int64(delta)).Err()
}
