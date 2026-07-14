package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/app"
	"github.com/jorgesanchez-e/localenvironment/config"
)

func main() {
	setLog()
	app := createApp()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Infof("received signal %v, shutting down", sig)
		cancel()
	}()

	app.Run(ctx)
}

func setLog() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05",
		DisableColors:          true,
		DisableLevelTruncation: true,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func createApp() *app.DDNS {
	cnf, err := config.New()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	simpleDDNS, err := cnf.GetSimpleDDNSConfig()
	if err != nil {
		log.Fatalf("Failed to get simple DDNS config: %v", err)
	}

	app, err := app.NewDDNS(simpleDDNS)
	if err != nil {
		log.Fatalf("failed to create DDNS app: %v", err)
	}

	return app
}
