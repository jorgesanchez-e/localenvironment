package store

import (
	"context"
	"fmt"

	"github.com/jorgesanchez-e/localenvironment/apps/acme/internal/domain/certs"
	"github.com/jorgesanchez-e/localenvironment/apps/acme/internal/infra/store/sqlite"
)

type Service struct {
	driver certs.Repository
}

func NewService(ctx context.Context, cfg Configurator) (*Service, error) {
	storeConfig, err := readConfig(cfg)
	if err != nil {
		return nil, err
	}

	return newStore(ctx, storeConfig)
}

func newStore(ctx context.Context, storeConfig *storeConfig) (*Service, error) {
	srv := new(Service)

	switch storeConfig.Type {
	case "sqlite":
		driver, err := sqlite.New(ctx, storeConfig.SQLite.File)
		if err != nil {
			return nil, err
		}

		srv.driver = driver
	default:
		return nil, fmt.Errorf("%w: %q (supported types: sqlite)", ErrUnsupportedStoreType, storeConfig.Type)
	}

	return srv, nil
}

func (s *Service) Load(ctx context.Context, cert *certs.Cert) error {
	return s.driver.Load(ctx, cert)
}

func (s *Service) Save(ctx context.Context, cert *certs.Cert) error {
	return s.driver.Save(ctx, cert)
}
