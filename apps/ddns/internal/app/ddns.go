package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/domain"
	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/infra/ipgetter"
	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/infra/updater"
	"github.com/jorgesanchez-e/localenvironment/config"
	log "github.com/sirupsen/logrus"
)

var (
	ErrCheckInSecondsMustBeLessThanUpdateTimeoutSeconds = fmt.Errorf("checkInSeconds must be less than updateTimeoutSeconds")
	ErrCheckInSecondsMustBeGreaterThanZero              = fmt.Errorf("checkInSeconds must be greater than zero")
)

type DDNS struct {
	checkInSeconds       time.Duration
	updateTimeoutSeconds time.Duration
	domain.IPGetter
	domain.DDNS
}

func NewDDNS(awsConfig *config.SimpleDDNS) (*DDNS, error) {
	ddnsUpdater, err := updater.NewUpdater(awsConfig)
	if err != nil {
		return nil, err
	}

	checkInSeconds := time.Duration(awsConfig.DDNS.CheckEverySeconds) * time.Second
	updateTimeoutSeconds := time.Duration(awsConfig.DDNS.UpdateTimeoutSeconds) * time.Second

	if checkInSeconds <= 0 {
		return nil, ErrCheckInSecondsMustBeGreaterThanZero
	}

	if checkInSeconds < updateTimeoutSeconds {
		return nil, ErrCheckInSecondsMustBeLessThanUpdateTimeoutSeconds
	}

	return &DDNS{
		time.Duration(awsConfig.DDNS.CheckEverySeconds) * time.Second,
		time.Duration(awsConfig.DDNS.UpdateTimeoutSeconds) * time.Second,
		ipgetter.NewIPGetter(),
		ddnsUpdater,
	}, nil
}

func (ddns *DDNS) Run(ctx context.Context) {
loop:
	for {
		ctxIteration, cancel := context.WithTimeout(ctx, ddns.checkInSeconds-1*time.Second) // -1 second to avoid timeout before the checkInSeconds
		go ddns.do(ctxIteration)

		select {
		case <-ctxIteration.Done():
			cancel()
		case <-ctx.Done():
			cancel()
			break loop
		}
	}

	log.Info("DDNS loop finished")
}

func (ddns *DDNS) do(ctx context.Context) {
	var err error

	defer func() {
		if err != nil {
			log.Errorf("failed to update records: %v", err)
		}
	}()

	ip4, ip6 := ddns.getIPs(ctx)
	if ip4 == nil && ip6 == nil {
		return
	}

	records, err := ddns.GetRecords(ctx)
	if err != nil {
		return
	}

	if records = ddns.checkIPs(records, ip4, ip6); len(records) == 0 {
		log.Info("no records to update")
		return
	}

	err = ddns.UpdateRecords(ctx, records)
	if err != nil {
		return
	}
}

func (ddns *DDNS) getIPs(ctx context.Context) (*string, *string) {
	ipv4, err := ddns.GetIPV4(ctx)
	if err != nil {
		log.Errorf("failed to get IPv4: %v", err)
	}

	ipv6, err := ddns.GetIPV6(ctx)
	if err != nil {
		log.Errorf("failed to get IPv6: %v", err)
	}

	return &ipv4, &ipv6
}

func (ddns *DDNS) checkIPs(records []domain.Record, ipv4 *string, ipv6 *string) []domain.Record {
	newRecords := make([]domain.Record, 0, len(records))

	for _, record := range records {
		if record.IPType == "A" && ipv4 != nil && *ipv4 != record.IP {
			newRecords = append(newRecords, domain.Record{
				IP:     *ipv4,
				IPType: record.IPType,
				FQDN:   record.FQDN,
			})
		}

		if record.IPType == "AAAA" && ipv6 != nil && *ipv6 != record.IP && *ipv6 != "" {
			newRecords = append(newRecords, domain.Record{
				IP:     *ipv6,
				IPType: record.IPType,
				FQDN:   record.FQDN,
			})
		}
	}

	return newRecords
}
