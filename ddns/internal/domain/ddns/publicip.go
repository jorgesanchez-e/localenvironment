package ddns

import (
	"context"
	"net"
)

type Fetcher interface {
	IPv4(ctx context.Context) (net.IP, error)
	IPv6(ctx context.Context) (net.IP, error)
}
