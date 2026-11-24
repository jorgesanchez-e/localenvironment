package aws

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
)

const (
	configKey string = "acme"
)

var (
	ErrReadConfig                = errors.New("aws-lego: read config error")
	ErrInvalidConfiguration      = errors.New("aws-lego: invalid configuration data")
	ErrInvalidConfigurationValue = errors.New("aws-lego: invalid configuration value")
	ErrAWSCredentialsNotFound    = errors.New("aws-lego: aws credentials not found")
)

type Configurator interface {
	Get(key string, value any) error
}

type acmeConfig struct {
	Email      string      `mapstructure:"email" validate:"required,email"`
	PrimaryKey string      `mapstructure:"primary_key" validate:"required"`
	AWSConf    []awsConfig `mapstructure:"aws" validate:"required,gt=0,dive"`
}

type awsConfig struct {
	Account   string                `mapstructure:"account" validate:"required,notblank"`
	Accesskey string                `mapstructure:"access-key" validate:"required,notblank"`
	Region    string                `mapstructure:"region" validate:"required,notblank"`
	SecretKey string                `mapstructure:"secret-key" validate:"required,notblank"`
	Domains   []route53DomainConfig `mapstructure:"domains" validate:"required,gt=0,dive"`
}

type route53DomainConfig struct {
	FQDN   string `mapstructure:"name" validate:"required,fqdn"`
	ZoneID string `mapstructure:"zone-id" validate:"required,notblank"`
}

type awsCredentials struct {
	accessKey string
	secretKey string
	region    string
	zoneID    string
}

func readConfig(cnf Configurator) (*acmeConfig, error) {
	acmeCnf := acmeConfig{}

	err := cnf.Get(configKey, &acmeCnf)
	if err != nil {
		return nil, fmt.Errorf("%w, read error: %w", ErrReadConfig, err)
	}

	err = validateConfig(&acmeCnf)
	if err != nil {
		return nil, err
	}

	return &acmeCnf, nil
}

func newValidator() *validator.Validate {
	validate := validator.New()

	// RegisterValidation only fails for an empty tag or a nil func, neither can happen here.
	_ = validate.RegisterValidation("notblank", validators.NotBlank)

	return validate
}

func validateConfig(data *acmeConfig) error {
	err := newValidator().Struct(data)
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

func (aa awsConfig) getCredentials(domain string) (*awsCredentials, error) {
	credentials := &awsCredentials{}

	for _, aaDomain := range aa.Domains {
		if aaDomain.FQDN == domain {
			credentials.accessKey = aa.Accesskey
			credentials.secretKey = aa.SecretKey
			credentials.region = aa.Region
			credentials.zoneID = aaDomain.ZoneID
			return credentials, nil
		}
	}

	return nil, fmt.Errorf("%w for domain:%q", ErrAWSCredentialsNotFound, domain)
}
