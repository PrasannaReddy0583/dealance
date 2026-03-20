package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	balanceTTL    = 5 * time.Minute
	balancePrefix = "wallet:balance:"
)

type CacheRepo struct{ rdb *redis.Client }
func NewCacheRepo(rdb *redis.Client) *CacheRepo { return &CacheRepo{rdb: rdb} }

func (r *CacheRepo) CacheBalance(ctx context.Context, userID string, balance, locked int64) error {
	key := balancePrefix + userID
	pipe := r.rdb.Pipeline()
	pipe.HSet(ctx, key, "balance", balance, "locked", locked)
	pipe.Expire(ctx, key, balanceTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *CacheRepo) GetBalance(ctx context.Context, userID string) (int64, int64, error) {
	key := balancePrefix + userID
	result, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil || len(result) == 0 { return 0, 0, fmt.Errorf("cache miss") }
	balance, _ := strconv.ParseInt(result["balance"], 10, 64)
	locked, _ := strconv.ParseInt(result["locked"], 10, 64)
	return balance, locked, nil
}

func (r *CacheRepo) InvalidateBalance(ctx context.Context, userID string) error {
	return r.rdb.Del(ctx, balancePrefix+userID).Err()
}
