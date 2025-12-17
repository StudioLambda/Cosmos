package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/studiolambda/cosmos/contract"
)

type Driver struct {
	cache   contract.Cache
	options Options
}

type Options struct {
	prefix string
}

var ErrInvalidCacheType = errors.New("invalid cache type")

func NewDriver(cache contract.Cache) *Driver {
	return NewDriverWith(cache, Options{
		prefix: "cosmos.sessions",
	})
}

func NewDriverWith(cache contract.Cache, options Options) *Driver {
	return &Driver{
		cache:   cache,
		options: options,
	}
}

func (d *Driver) key(id string) string {
	return fmt.Sprintf("%s.%s", d.options.prefix, id)
}

func (d *Driver) Get(ctx context.Context, id string) (contract.Session, error) {
	k := d.key(id)
	v, err := d.cache.Get(ctx, k)

	if err != nil {
		return nil, err
	}

	if s, ok := v.(contract.Session); ok {
		return s, nil
	}

	return nil, ErrInvalidCacheType
}

func (d *Driver) Save(ctx context.Context, session contract.Session, ttl time.Duration) error {
	k := d.key(session.SessionID())

	if err := d.cache.Put(ctx, k, session, ttl); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Delete(ctx context.Context, id string) error {
	k := d.key(id)

	return d.cache.Delete(ctx, k)
}
