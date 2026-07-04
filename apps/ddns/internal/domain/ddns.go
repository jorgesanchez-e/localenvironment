package domain

import "context"

type IPGetter interface {
	GetIPV4(ctx context.Context) (string, error)
	GetIPV6(ctx context.Context) (string, error)
}

type Record struct {
	IP     string
	IPType string
	FQDN   string
}

type DDNS interface {
	GetRecords(ctx context.Context, fqdns []string) ([]Record, error)
	UpdateRecords(ctx context.Context, records []Record) error
}
