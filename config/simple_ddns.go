package config

import (
	"net/netip"

	"github.com/go-playground/validator/v10"
)

const ddnsConfigName = "ddns"

type SimpleDDNS struct {
	DDNS struct {
		LogLevel             string `mapstructure:"log-level"`
		CheckEverySeconds    int    `mapstructure:"check-every-seconds"`
		UpdateTimeoutSeconds int    `mapstructure:"process-timeout-seconds"`
		AWS                  []struct {
			AccountName string       `mapstructure:"account-name" validate:"required,alphanum"`
			Region      string       `mapstructure:"region"`
			AccessKey   string       `mapstructure:"access-key"`
			SecretKey   string       `mapstructure:"secret-key"`
			Zones       []zoneConfig `mapstructure:"zones" validate:"dive"`
		} `mapstructure:"aws" validate:"dive"`
	} `mapstructure:"ddns"`
}

type zoneConfig struct {
	RecordName  string `mapstructure:"record-name" validate:"required,fqdn"`
	RecordType  string `mapstructure:"record-type" validate:"required,oneof=A AAAA"`
	RecordTTL   int    `mapstructure:"record-ttl" validate:"required,min=1"`
	RecordValue string `mapstructure:"record-value" validate:"required"`
}

func (c *config) GetSimpleDDNSConfig() (*SimpleDDNS, error) {
	var simpleDDNS SimpleDDNS
	c.vp.UnmarshalKey(ddnsConfigName, &simpleDDNS.DDNS)

	validate := validator.New()
	validate.RegisterStructValidation(OwnIpValidator, zoneConfig{})
	if err := validate.Struct(simpleDDNS); err != nil {
		return nil, err
	}

	return &simpleDDNS, nil
}

func OwnIpValidator(sl validator.StructLevel) {
	cfg := sl.Current().Interface().(zoneConfig)

	if ipAddr, err := netip.ParseAddr(cfg.RecordValue); err == nil {
		if cfg.RecordType == "A" && !ipAddr.Is4() {
			sl.ReportError(cfg.RecordValue, "RecordValue", "RecordValue", "ipv4_address", "")
		}
		if cfg.RecordType == "AAAA" && !ipAddr.Is6() {
			sl.ReportError(cfg.RecordValue, "RecordValue", "RecordValue", "ipv6_address", "")
		}
	} else {
		sl.ReportError(cfg.RecordValue, "RecordValue", "RecordValue", "invalid_ip_address", "")
	}
}
