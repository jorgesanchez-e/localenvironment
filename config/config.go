package config

import (
	"github.com/spf13/viper"
)

const (
	configFileType string = "yaml"
	configName     string = "config"
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

	for _, path := range paths {
		cnf.vp.AddConfigPath(path)
	}

	cnf.vp.SetConfigName(configName)
	cnf.vp.SetConfigType(configFileType)
	if err := cnf.vp.ReadInConfig(); err != nil {
		return nil, err
	}

	return cnf, nil
}
