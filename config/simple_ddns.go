package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

const ddnsConfigName = "ddns"

type SimpleDDNS struct {
	DDNS DDNSConfig `mapstructure:"ddns"`
}

type DDNSConfig struct {
	LogLevel             string      `mapstructure:"log-level"`
	CheckEverySeconds    int         `mapstructure:"check-every-seconds"`
	UpdateTimeoutSeconds int         `mapstructure:"process-timeout-seconds"`
	AWS                  []AWSConfig `mapstructure:"aws" validate:"dive"`
}

type AWSConfig struct {
	AccountName string       `mapstructure:"account-name" validate:"required,alphanum"`
	Region      string       `mapstructure:"region"`
	AccessKey   string       `mapstructure:"access-key"`
	SecretKey   string       `mapstructure:"secret-key"`
	Zones       []ZoneConfig `mapstructure:"zones" validate:"dive"`
}

type ZoneConfig struct {
	ID      string         `mapstructure:"id" validate:"required,alphanum"`
	Name    string         `mapstructure:"name"`
	Records []RecordConfig `mapstructure:"records" validate:"dive"`
}

type RecordConfig struct {
	FQDN       string `mapstructure:"fqdn" validate:"required,fqdn"`
	RecordType string `mapstructure:"record-type" validate:"required,oneof=A AAAA"`
	RecordTTL  int    `mapstructure:"record-ttl" validate:"required,min=1"`
}

func (c *conf) GetSimpleDDNSConfig() (*SimpleDDNS, error) {
	var simpleDDNS SimpleDDNS
	if err := c.vp.UnmarshalKey(ddnsConfigName, &simpleDDNS.DDNS); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(simpleDDNS); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &simpleDDNS, nil
}
