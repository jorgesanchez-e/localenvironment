package publicip

import (
	"context"
	"net"

	"github.com/jorgesanchez-e/localenvironment/ddns/internal/adapters/publicip/ipify"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	ipifyClient ipify.Client
}

func New() *Client {
	return &Client{
		ipifyClient: ipify.New(),
	}
}

func (c Client) IPv4(ctx context.Context) net.IP {
	ip, err := c.ipifyClient.GetIP(ctx, ipify.IPv4URL)
	if err != nil {
		log.Error(err)
		return nil
	}

	return ip
}

func (c Client) IPv6(ctx context.Context) net.IP {
	ip, err := c.ipifyClient.GetIP(ctx, ipify.IPv6URL)
	if err != nil {
		log.Error(err)
		return nil
	}

	return ip
}
