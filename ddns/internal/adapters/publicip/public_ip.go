package publicip

import (
	"context"
	"net"

	"github.com/jorgesanchez-e/localenvironment/ddns/internal/adapters/publicip/ipify"
)

type Client struct {
	ipifyClient ipify.Client
}

func New() *Client {
	return &Client{
		ipifyClient: ipify.New(),
	}
}

func (c Client) IPv4(ctx context.Context) (net.IP, error) {
	ip, err := c.ipifyClient.GetIP(ctx, ipify.IPv4URL)
	if err != nil {
		return nil, err
	}

	return ip, nil
}

func (c Client) IPv6(ctx context.Context) (net.IP, error) {
	ip, err := c.ipifyClient.GetIP(ctx, ipify.IPv6URL)
	if err != nil {
		return nil, err
	}

	return ip, err
}
