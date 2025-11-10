package dns

import (
	"context"
	"errors"
	"net"

	"github.com/jorgesanchez-e/localenvironment/ddns/internal/adapters/dns/route53"
)

var ErrNoUpdates = errors.New("any update to do")

type configDecoder interface {
	Decode(node string, item any) error
}

type Client struct {
	ddns *route53.Client
}

func New(ctx context.Context, cnf configDecoder) (_ *Client, err error) {
	client := &Client{}
	if client.ddns, err = route53.New(ctx, cnf); err != nil {
		return nil, err
	}

	return client, nil
}

func (c Client) Update(ctx context.Context, ips []net.IP) error {
	batches, err := c.ddns.CreateBatches(ctx, ips)
	if err != nil {
		return err
	}

	if len(batches) == 0 {
		return ErrNoUpdates
	}

	if err := c.ddns.Update(ctx, batches); err != nil {
		return err
	}

	return nil
}
