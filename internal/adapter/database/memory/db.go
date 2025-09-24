package memory

import (
	"context"
	"time"

	"todoapp/internal/core/port"
)

type memoryRepository struct{}

func NewMemoryRepository() port.CacheRepository {
	return &memoryRepository{}
}

func (c *memoryRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

func (c *memoryRepository) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (c *memoryRepository) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *memoryRepository) DeleteByPrefix(ctx context.Context, prefix string) error {
	return nil
}

func (c *memoryRepository) Close() error {
	return nil
}
