package ddns

import (
	"context"
	"net"
)

const (
	RType RecordType = iota
	AAAA
	A
)

type RecordType int

type Record interface {
	Address() net.IP
	FQDN() string
	RType() RecordType
}

type Storer interface {
	Save(ctx context.Context, record Record) error
	Last(ctx context.Context, fqdn string, rtype RecordType) Record
}
