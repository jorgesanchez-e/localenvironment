package aws

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
)

type legoClient struct {
	client *lego.Client
}

func legoClientBuilder(user *acmeUser) func() (acmeClient, error) {
	return func() (acmeClient, error) {
		cfg := lego.NewConfig(user)
		cfg.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
		if staging {
			cfg.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
		}

		client, err := lego.NewClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("%w, error:%q", ErrAcmeLegoClient, err)
		}

		return &legoClient{client: client}, nil
	}
}

func (l *legoClient) SetDNS01Provider(p challenge.Provider,
	opts ...dns01.ChallengeOption) error {
	return l.client.Challenge.SetDNS01Provider(p, opts...)
}

func (l *legoClient) Register(ctx context.Context,
	opts registration.RegisterOptions) (*acme.ExtendedAccount, error) {
	return l.client.Registration.Register(ctx, opts)
}

func (l *legoClient) Obtain(ctx context.Context,
	req certificate.ObtainRequest) (*certificate.Resource, error) {
	return l.client.Certificate.Obtain(ctx, req)
}

func (l *legoClient) Renew(ctx context.Context, certRes certificate.Resource,
	options *certificate.RenewOptions) (*certificate.Resource, error) {
	return l.client.Certificate.Renew(ctx, certRes, options)
}
