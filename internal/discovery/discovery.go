package discovery

import (
	"context"
	"io"
)

type ServiceDiscovery interface {
	io.Closer
	Register(ctx context.Context) error
	Deregister(ctx context.Context) error
	FindServiceByName(ctx context.Context, name string) (addr string, err error)
}
