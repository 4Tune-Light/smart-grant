package idempotency

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const ttl = 24 * time.Hour

type CachedResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

type Store struct {
	rdb redis.UniversalClient
}

func NewStore(rdb redis.UniversalClient) *Store {
	return &Store{rdb: rdb}
}

func (s *Store) IsProcessed(ctx context.Context, key string) bool {
	if s.rdb == nil {
		return false
	}
	exists, err := s.rdb.Exists(ctx, s.key(key)).Result()
	if err != nil {
		return false
	}
	return exists > 0
}

func (s *Store) Get(ctx context.Context, key string) (*CachedResponse, error) {
	if s.rdb == nil {
		return nil, nil
	}
	data, err := s.rdb.Get(ctx, s.key(key)).Bytes()
	if err != nil {
		return nil, err
	}
	var resp CachedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *Store) Set(ctx context.Context, key string, resp *CachedResponse) error {
	if s.rdb == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.key(key), data, ttl).Err()
}

func (s *Store) key(k string) string {
	return fmt.Sprintf("idempotency:%s", k)
}
