package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCacheNotFound = errors.New("cache not found")
	ErrCacheExpired  = errors.New("cache expired")
)

type Cache interface {
	Put(ctx context.Context, key string, val interface{}, ttl ...time.Duration) error
	Get(ctx context.Context, key string, val interface{}) error
	TTL(ctx context.Context, key string, ttl *time.Duration) error
	Del(ctx context.Context, keys ...string) error
}
