package main

import (
	"context"
	"strings"

	configreader "github.com/jorgesanchez-e/localenvironment/common/config-reader"
	"github.com/jorgesanchez-e/localenvironment/ddns/app/service"
	log "github.com/sirupsen/logrus"
)

type configDecoder interface {
	Decode(node string, item any) error
}

func main() {
	ctx := context.Background()
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
	})

	cnf, err := configreader.New()
	if err != nil {
		log.Fatal(err)
	}

	setLogLevel(cnf)

	srv, err := service.New(ctx, cnf)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("starting")

	cancel := srv.Run(ctx)
	<-cancel
}

func setLogLevel(cnf configDecoder) {
	logLevel := ""

	err := cnf.Decode("ddns.log-level", &logLevel)
	if err != nil {
		log.Fatal(err)
	}

	switch strings.ToLower(logLevel) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}
