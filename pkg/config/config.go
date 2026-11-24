package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

const (
	configFileType string = "yaml"
	configName     string = "config"
)

var (
	ErrReadConfig  = errors.New("config: failed to read config file")
	ErrKeyNotFound = errors.New("config: key not found")
	ErrUnmarshal   = errors.New("config: failed to unmarshal key")
)

var paths = []string{
	"/usr/local/etc/localenvironment/",
	"/etc/localenvironment/",
	"$HOME/.localenvironment/",
	".",
}

type conf struct {
	vp *viper.Viper
}

func New() (*conf, error) {
	cnf := new(conf)
	cnf.vp = viper.New()
	cnf.vp.SetConfigType(configFileType)
	cnf.vp.SetConfigName(configName)

	for _, path := range paths {
		cnf.vp.AddConfigPath(path)
	}

	err := cnf.vp.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadConfig, err)
	}

	return cnf, nil
}

func (c *conf) Get(key string, value any) error {
	if !c.vp.IsSet(key) {
		return fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	if err := c.vp.UnmarshalKey(key, value); err != nil {
		return fmt.Errorf("%w %s: %w", ErrUnmarshal, key, err)
	}

	return nil
}
