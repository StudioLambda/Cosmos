package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/studiolambda/cosmos/contract"
)

type CacheDriver struct {
	cache   contract.Cache
	options CacheDriverOptions
}

type CacheDriverOptions struct {
	prefix string
}

var ErrCacheDriverInvalidType = errors.New("invalid cache type")

func NewCacheDriver(cache contract.Cache) *CacheDriver {
	return NewCacheDriverWith(cache, CacheDriverOptions{
		prefix: "cosmos.sessions",
	})
}

func NewCacheDriverWith(cache contract.Cache, options CacheDriverOptions) *CacheDriver {
	return &CacheDriver{
		cache:   cache,
		options: options,
	}
}

func (d *CacheDriver) key(id string) string {
	return fmt.Sprintf("%s.%s", d.options.prefix, id)
}

func (d *CacheDriver) Get(ctx context.Context, id string) (contract.Session, error) {
	k := d.key(id)
	v, err := d.cache.Get(ctx, k)

	if err != nil {
		return nil, err
	}

	if s, ok := v.(contract.Session); ok {
		return s, nil
	}

	return nil, ErrCacheDriverInvalidType
}

func (d *CacheDriver) Save(ctx context.Context, session contract.Session, ttl time.Duration) error {
	k := d.key(session.SessionID())

	if err := d.cache.Put(ctx, k, session, ttl); err != nil {
		return err
	}

	return nil
}

func (d *CacheDriver) Delete(ctx context.Context, id string) error {
	k := d.key(id)

	return d.cache.Delete(ctx, k)
}
