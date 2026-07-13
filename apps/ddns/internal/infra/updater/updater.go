package updater

import (
	"context"
	"errors"

	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/domain"
	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/infra/updater/r53"
	"github.com/jorgesanchez-e/localenvironment/config"
)

type r53Updater interface {
	GetRecords(ctx context.Context, domains []string) ([]r53.Record, error)
	UpdateRecords(ctx context.Context, records []r53.Record) error
}

type Updater struct {
	domains    []string
	r53Updater r53Updater
}

func awsConfigDomains(awsConfig *config.SimpleDDNS) []string {
	domains := make([]string, 0)
	for _, awsConfig := range awsConfig.DDNS.AWS {
		for _, zone := range awsConfig.Zones {
			for _, record := range zone.Records {
				domains = append(domains, record.FQDN)
			}
		}
	}

	return domains
}

func NewUpdater(ddnsConfig *config.SimpleDDNS) (*Updater, error) {
	if ddnsConfig == nil {
		return nil, errors.New("ddns config is required")
	}

	if ddnsConfig.DDNS.AWS == nil {
		return nil, errors.New("aws config is required")
	}

	updater := &Updater{
		r53Updater: r53.NewR53(ddnsConfig.DDNS.AWS),
		domains:    awsConfigDomains(ddnsConfig),
	}

	return updater, nil
}

func (u *Updater) GetRecords(ctx context.Context) ([]domain.Record, error) {
	records, err := u.r53Updater.GetRecords(ctx, u.domains)
	if err != nil {
		return nil, err
	}

	domainRecords := make([]domain.Record, 0, len(records))
	for _, record := range records {
		domainRecords = append(domainRecords, domain.Record{
			IP:     record.IP,
			IPType: record.RecordType,
			FQDN:   record.FQDN,
		})
	}

	return domainRecords, nil
}

func (u *Updater) UpdateRecords(ctx context.Context, records []domain.Record) error {
	if len(records) == 0 {
		return nil
	}

	return u.r53Updater.UpdateRecords(ctx, domainRecordsToR53Records(records))
}

func domainRecordsToR53Records(records []domain.Record) []r53.Record {
	recordsList := make([]r53.Record, 0, len(records))
	for _, record := range records {
		recordsList = append(recordsList, r53.Record{
			IP:         record.IP,
			RecordType: record.IPType,
			FQDN:       record.FQDN,
		})
	}
	return recordsList
}
