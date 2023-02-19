package cache

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
	"go-micro.dev/v4/codec"
	"go-micro.dev/v4/codec/json"
)

var (
	Permenent = time.Duration(0)
)

type redisCache struct {
	cli       *redis.Client
	marshaler codec.Marshaler
}

func Redis(cli *redis.Client, marshalers ...codec.Marshaler) Cache {
	if len(marshalers) == 0 {
		return &redisCache{cli: cli, marshaler: json.Marshaler{}}
	}
	return &redisCache{cli: cli, marshaler: marshalers[0]}
}

func (r *redisCache) Put(ctx context.Context, key string, val interface{}, ttls ...time.Duration) error {
	ttl := Permenent
	if len(ttls) > 0 {
		ttl = ttls[0]
	}
	bs, err := r.marshaler.Marshal(val)
	if err != nil {
		return err
	}
	return r.cli.Set(ctx, key, bs, ttl).Err()
}

func (r *redisCache) Get(ctx context.Context, key string, val interface{}) error {
	str, err := r.cli.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return r.marshaler.Unmarshal([]byte(str), val)
}

func (r *redisCache) TTL(ctx context.Context, key string, ttl *time.Duration) error {
	var err error
	*ttl, err = r.cli.TTL(ctx, key).Result()
	return err
}

func (r *redisCache) Del(ctx context.Context, keys ...string) error {
	return r.cli.Del(ctx, keys...).Err()
}
