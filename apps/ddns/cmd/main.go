package main

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/domain"
	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/infra/ipgetter"
	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/infra/updater"
	"github.com/jorgesanchez-e/localenvironment/config"
)

func main() {
	ctx := context.Background()

	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)

	cnf, err := config.New()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	simpleDDNS, err := cnf.GetSimpleDDNSConfig()
	if err != nil {
		log.Fatalf("Failed to get simple DDNS config: %v", err)
	}

	log.Infof("Simple DDNS config: %v", simpleDDNS)

	ipGetter := ipgetter.NewIPGetter()
	ipv4, err := ipGetter.GetIPV4(ctx)
	if err != nil {
		log.Errorf("ipgetter error %v", err)
	}

	ipv6, err := ipGetter.GetIPV6(ctx)
	if err != nil {
		log.Errorf("ipgetter error %v", err)
	}

	log.Infof("IPv4: %s", ipv4)
	log.Infof("IPv6: %s", ipv6)

	updater, err := updater.NewUpdater(simpleDDNS)
	if err != nil {
		log.Fatalf("failed to create updater: %v", err)
	}

	records, err := updater.GetRecords(ctx)
	if err != nil {
		log.Fatalf("failed to get records: %v", err)
	}

	recordsToUpdate := make([]domain.Record, 0, len(records))
	for _, record := range records {
		if record.IPType == "A" && record.IP != ipv4 && ipv4 != "" {
			recordsToUpdate = append(recordsToUpdate, domain.Record{
				FQDN:   record.FQDN,
				IPType: record.IPType,
				IP:     ipv4,
			})

		}

		if record.IPType == "AAAA" && record.IP != ipv6 && ipv6 != "" {
			recordsToUpdate = append(recordsToUpdate, domain.Record{
				FQDN:   record.FQDN,
				IPType: record.IPType,
				IP:     ipv6,
			})
		}
	}

	err = updater.UpdateRecords(ctx, recordsToUpdate)
	if err != nil {
		log.Fatalf("failed to update records: %v", err)
	} else {
		log.Infof("Records updated successfully")
	}
}
