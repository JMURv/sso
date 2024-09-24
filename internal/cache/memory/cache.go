package memory

import (
	"context"
	errs "github.com/JMURv/sso/internal/cache"
	md "github.com/JMURv/sso/pkg/model"
	"sync"
	"time"
)

type Cache struct {
	sync.Mutex
	data map[string]*md.User
}

func New() *Cache {
	return &Cache{data: make(map[string]*md.User)}
}

func (c *Cache) Get(_ context.Context, key string) (*md.User, error) {
	c.Lock()
	defer c.Unlock()
	if v, ok := c.data[key]; !ok {
		return nil, errs.ErrNotFoundInCache
	} else {
		return v, nil
	}
}

func (c *Cache) Set(_ context.Context, t time.Duration, key string, r *md.User) error {
	c.Lock()
	defer c.Unlock()
	c.data[key] = r
	return nil
}

func (c *Cache) Delete(_ context.Context, key string) error {
	c.Lock()
	defer c.Unlock()
	delete(c.data, key)
	return nil
}
