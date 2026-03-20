package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/dealance/services/deal/internal/domain/entity"
)

const (
	dealCacheTTL    = 10 * time.Minute
	dealCachePrefix = "deal:profile:"
)

type CacheRepo struct{ rdb *redis.Client }
func NewCacheRepo(rdb *redis.Client) *CacheRepo { return &CacheRepo{rdb: rdb} }

func (r *CacheRepo) CacheDeal(ctx context.Context, deal *entity.Deal) error {
	data, err := json.Marshal(deal); if err != nil { return err }
	return r.rdb.Set(ctx, dealCachePrefix+deal.ID.String(), data, dealCacheTTL).Err()
}
func (r *CacheRepo) GetCachedDeal(ctx context.Context, id string) (*entity.Deal, error) {
	data, err := r.rdb.Get(ctx, dealCachePrefix+id).Bytes(); if err != nil { return nil, err }
	var deal entity.Deal; return &deal, json.Unmarshal(data, &deal)
}
func (r *CacheRepo) InvalidateDeal(ctx context.Context, id string) error {
	return r.rdb.Del(ctx, dealCachePrefix+id).Err()
}
