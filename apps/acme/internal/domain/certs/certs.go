package certs

import "context"

type Repository interface {
	Save(ctx context.Context, cert *Cert) error
	Load(ctx context.Context, cert *Cert) error
}

type Driver interface {
	Upsert(ctx context.Context, certs []*Cert) error
	Revoke(ctx context.Context, certs []*Cert) error
}
