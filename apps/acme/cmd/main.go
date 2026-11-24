package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jorgesanchez-e/localenvironment/apps/acme/internal/domain/certs"
	"github.com/jorgesanchez-e/localenvironment/apps/acme/internal/infra/acme"
	"github.com/jorgesanchez-e/localenvironment/apps/acme/internal/infra/store"
	"github.com/jorgesanchez-e/localenvironment/pkg/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	cnf, err := config.New()
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	repository, err := store.NewService(ctx, cnf)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	certManager, err := acme.New(ctx, cnf)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	c := certs.NewCert("test-a.localenvironment.dev")
	c2 := certs.NewCert("test-aaaa.localenvironment.dev")
	data := []*certs.Cert{c, c2}

	if err = certManager.Upsert(ctx, data); err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	for _, c := range data {
		if err = repository.Save(ctx, c); err != nil {
			log.Error(err)
			os.Exit(-1)
		}

		fmt.Printf("saved certificate for domain:%s", c.Domain())
	}
}
