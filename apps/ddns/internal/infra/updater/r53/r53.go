package r53

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/jorgesanchez-e/localenvironment/config"
	log "github.com/sirupsen/logrus"
)

type updateGetter interface {
	ListResourceRecordSets(context.Context, *route53.ListResourceRecordSetsInput, ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error)
	ChangeResourceRecordSets(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error)
}

type awsClient struct {
	client updateGetter
	zones  []zone
}

type zone struct {
	id      string
	records []Record
}

type Record struct {
	FQDN       string
	IP         string
	RecordType string
	RecordTTL  int
}

type records []Record

type Updater struct {
	drivers []awsClient
}

func NewR53(accounts []config.AWSConfig) *Updater {
	updater := &Updater{
		drivers: []awsClient{},
	}

	for _, account := range accounts {
		awsConfig := AWSConfig(account)
		creds := awsConfig.getCredentials()
		awsRoute53Client := route53.NewFromConfig(aws.Config{Credentials: creds, Region: account.Region})
		awsDriver := awsClient{
			client: awsRoute53Client,
			zones:  awsConfig.zones(),
		}

		updater.drivers = append(updater.drivers, awsDriver)
	}

	return updater
}

func (r *Updater) GetRecords(ctx context.Context, domains []string) ([]Record, error) {
	records := make([]Record, 0)

	for _, driver := range r.drivers {
		for _, zone := range driver.zones {
			params := &route53.ListResourceRecordSetsInput{
				HostedZoneId: aws.String(zone.id),
			}

			paginator := route53.NewListResourceRecordSetsPaginator(driver.client, params)
			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)
				if err != nil {
					return nil, err
				}
				for _, cloudRecord := range page.ResourceRecordSets {
					localRecord := translateRecord(&cloudRecord)
					if localRecord == nil {
						continue
					}

					if localRecord.RecordType == "A" || localRecord.RecordType == "AAAA" {
						for _, domain := range domains {
							if domain == localRecord.FQDN {
								records = append(records, *localRecord)
							}
						}
					}
				}
			}
		}
	}

	return records, nil
}

func (r *Updater) UpdateRecords(ctx context.Context, recs []Record) error {
	ch := make(chan struct{})
	wg := sync.WaitGroup{}

	for _, driver := range r.drivers {
		for _, request := range driver.buildRequests(records(recs).check()) {
			wg.Add(1)
			go driver.do(ctx, &wg, ch, request)
		}
	}

	close(ch)
	wg.Wait()

	return nil
}

func (ac awsClient) buildRequests(inputRecords []Record) []*route53.ChangeResourceRecordSetsInput {
	requests := make([]*route53.ChangeResourceRecordSetsInput, 0, len(ac.zones))

	for _, zone := range ac.zones {
		batch := new(types.ChangeBatch)

		for _, record := range zone.records {
			changes := make([]types.Change, 0, len(inputRecords))
			for _, ir := range inputRecords {
				if ir.FQDN == record.FQDN && ir.RecordType == record.RecordType {
					changes = append(changes, types.Change{
						Action: types.ChangeActionUpsert,
						ResourceRecordSet: &types.ResourceRecordSet{
							Name: aws.String(ir.FQDN),
							Type: types.RRType(ir.RecordType),
							TTL:  aws.Int64(int64(record.RecordTTL)),
							ResourceRecords: []types.ResourceRecord{
								{
									Value: aws.String(ir.IP),
								},
							},
						},
					})
				}
			}

			if len(changes) > 0 {
				batch.Changes = append(batch.Changes, changes...)
			}
		}

		if len(batch.Changes) == 0 {
			continue
		}

		requests = append(requests, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(zone.id),
			ChangeBatch:  batch,
		})
	}

	return requests
}

func (ac awsClient) do(ctx context.Context, wg *sync.WaitGroup, ch <-chan struct{}, request *route53.ChangeResourceRecordSetsInput) {
	defer wg.Done()

	select {
	case <-ch:
		if _, err := ac.client.ChangeResourceRecordSets(ctx, request); err != nil {
			log.Errorf("failed to update records for aws hosted zone %s: %v", *request.HostedZoneId, err)
		} else {
			log.Infof("records updated successfully for aws hosted zone %s", *request.HostedZoneId)
		}
	case <-ctx.Done():
		return
	}
}

func (r records) check() []Record {
	checked := make([]Record, 0, len(r))

	for _, record := range r {
		if record.IP != "" && record.FQDN != "" && (record.RecordType == "A" || record.RecordType == "AAA") {
			checked = append(checked, record)
		}
	}
	return checked
}

func translateRecord(record *types.ResourceRecordSet) *Record {
	if record == nil || record.Name == nil || record.TTL == nil || len(record.ResourceRecords) == 0 {
		return nil
	}

	return &Record{
		FQDN:       *record.Name,
		IP:         *record.ResourceRecords[0].Value,
		RecordType: string(record.Type),
		RecordTTL:  int(*record.TTL),
	}
}
