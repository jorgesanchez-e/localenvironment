package store

import (
	"errors"
	"fmt"

	validator "github.com/go-playground/validator/v10"
)

var (
	rootConfigKey = "acme.store"

	ErrUnsupportedStoreType      = errors.New("unsupported store type")
	ErrInvalidConfiguration      = errors.New("invalid configuration data")
	ErrInvalidConfigurationValue = errors.New("invalid configuration value")
)

type Configurator interface {
	Get(key string, value any) error
}

type storeConfig struct {
	Type   string        `mapstructure:"type" validate:"required,oneof=sqlite"`
	SQLite *sqliteConfig `mapstructure:"sqlite" validate:"required_if=Type sqlite"`
}

type sqliteConfig struct {
	File string `mapstructure:"file" validate:"required,filepath"`
}

func readConfig(cfg Configurator) (*storeConfig, error) {
	storeCfg := &storeConfig{
		SQLite: &sqliteConfig{},
	}

	if err := cfg.Get(rootConfigKey, storeCfg); err != nil {
		return nil, err
	}

	if err := validateConfig(storeCfg); err != nil {
		return nil, err
	}

	return storeCfg, nil
}

func validateConfig(cfg *storeConfig) error {
	validate := validator.New()

	err := validate.Struct(cfg)
	if err == nil {
		return nil
	}

	// validator only returns a non-empty ValidationErrors when a field fails.
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		fe := validationErrs[0]
		return fmt.Errorf("%w: field:%q value:%q",
			ErrInvalidConfigurationValue,
			fe.Namespace(),
			fe.Value(),
		)
	}

	// Any other error (e.g. *validator.InvalidValidationError for a nil config).
	return fmt.Errorf("%w: %q", ErrInvalidConfiguration, err.Error())
}
