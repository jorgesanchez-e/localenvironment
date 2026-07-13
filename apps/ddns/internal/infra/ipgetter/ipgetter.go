package ipgetter

import (
	"context"

	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/domain"
)

const (
	ipTypeIPv4 ipType = iota
	ipTypeIPv6
)

type ipType int

type getter struct {
	ipify *ipify
}

func NewIPGetter() domain.IPGetter {
	return &getter{
		ipify: NewIPify(),
	}
}

func (g *getter) GetIPV4(ctx context.Context) (string, error) {
	return g.ipify.publicIP(ctx, ipTypeIPv4)
}

func (g *getter) GetIPV6(ctx context.Context) (string, error) {
	return g.ipify.publicIP(ctx, ipTypeIPv6)
}
