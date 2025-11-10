package ddns

import (
	"context"
	"net"
)

type Updater interface {
	Update(ctx context.Context, ips []net.IP) error
}
