package acme

import (
	"context"
	"errors"
	"fmt"

	"github.com/jorgesanchez-e/localenvironment/apps/acme/internal/domain/certs"
	"github.com/jorgesanchez-e/localenvironment/apps/acme/internal/infra/acme/aws"
)

var (
	acmeConfigKey = "acme"

	ErrConfigNotFound         = errors.New("acme: config not found")
	ErrNoImplementationsFound = errors.New("acme: no implementations found")
	ErrInvalidImplemnentation = errors.New("acme: inevalid implementation")
	ErrReadConfig             = errors.New("acme: read config error")
)

type Service struct {
	drivers map[implementedDriver]certs.Driver
}

func New(ctx context.Context, cnf Configurator) (*Service, error) {
	srv := new(Service)

	acmeCnf := acmeConfig{}
	if err := cnf.Get(acmeConfigKey, &acmeCnf); err != nil {
		return nil, ErrConfigNotFound
	}

	implementations := acmeCnf.implementations()
	if len(implementations) == 0 {
		return nil, ErrNoImplementationsFound
	}

	srv.drivers = make(map[implementedDriver]certs.Driver)

	for _, imp := range implementations {
		switch imp.String() {
		case "AWS":
			d, err := aws.New(ctx, cnf)
			if err != nil {
				return nil, err
			}
			srv.drivers[imp] = d
		default:
			return nil, fmt.Errorf("%w, implementation name: %q", ErrInvalidImplemnentation, imp.String())
		}
	}

	return srv, nil
}

func (s *Service) Upsert(ctx context.Context, certificates []*certs.Cert) error {
	certRequests := make([]*certs.Cert, 0, len(certificates))

	for _, c := range certificates {
		needs, err := c.NeedsRenewal(nil)
		if err != nil {
			if err == certs.ErrEmptyCert {
				needs = true
			} else {
				continue
			}
		}

		if needs {
			certRequests = append(certRequests, c)
		}
	}

	var err error
	for _, implementation := range s.drivers {
		err = errors.Join(err, implementation.Upsert(ctx, certRequests))
		for _, creq := range certRequests {
			for index, c := range certificates {
				if creq.Domain() == c.Domain() {
					certificates[index] = creq
				}
			}
		}
	}

	return err
}

func (s *Service) Revoke(ctx context.Context, certs []*certs.Cert) error {
	return nil
}
