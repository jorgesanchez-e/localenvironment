package route53

import (
	"context"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/go-playground/validator/v10"
)

type configDecoder interface {
	Decode(node string, item any) error
}

type updater interface {
	ChangeResourceRecordSets(
		ctx context.Context,
		params *route53.ChangeResourceRecordSetsInput,
		optFns ...func(*route53.Options),
	) (*route53.ChangeResourceRecordSetsOutput, error)
	ListResourceRecordSets(
		ctx context.Context,
		params *route53.ListResourceRecordSetsInput,
		optFns ...func(*route53.Options),
	) (*route53.ListResourceRecordSetsOutput, error)
}

type Client struct {
	updater      updater
	managedZones []zone
}

func New(ctx context.Context, configReader configDecoder) (_ *Client, err error) {
	config := r53Config{}
	if err = configReader.Decode(configKey, &config); err != nil {
		return nil, fmt.Errorf("unable to read config for key: %s, error:%w", configKey, err)
	}

	var awsCnf aws.Config
	awsCnf, err = awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithSharedConfigFiles([]string{config.CredentialsFile}),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to read aws credentials, error:%w", err)
	}

	client := Client{
		updater:      route53.NewFromConfig(awsCnf),
		managedZones: make([]zone, 0),
	}

	for _, z := range config.Zones {
		client.managedZones = append(client.managedZones, zone{
			ID:             z.ID,
			ManagedRecords: z.ManagedRecords,
		})
	}

	return &client, nil
}

func (c *Client) CreateBatches(ctx context.Context, ips []net.IP) (_ []Batch, err error) {
	mapedBatches := make(map[string]Batch, 0)

	records, err := c.records(ctx)
	if err != nil {
		return nil, err
	}

	for fqdn, currentIPs := range records {
		for _, currentIp := range currentIPs {
			for _, newIP := range ips {
				if sameVersion(currentIp, newIP) && !currentIp.Equal(newIP) {
					if zoneID := getZoneID(c.managedZones, fqdn); zoneID != "" {
						if currentBatch, exists := mapedBatches[zoneID]; exists {
							currentBatch.changeBatch.Changes = append(currentBatch.changeBatch.Changes, createChange(fqdn, newIP))
							mapedBatches[zoneID] = currentBatch
						} else {
							mapedBatches[zoneID] = Batch{
								zoneID: zoneID,
								changeBatch: &types.ChangeBatch{
									Changes: []types.Change{createChange(fqdn, newIP)},
								},
							}
						}
					}
				}
			}
		}
	}

	batches := make([]Batch, 0, len(mapedBatches))
	for _, batch := range mapedBatches {
		batches = append(batches, batch)
	}

	return batches, nil
}

func (c *Client) Update(ctx context.Context, batches []Batch) error {
	for _, batch := range batches {
		for _, zone := range c.managedZones {
			if zone.ID == batch.zoneID {
				_, err := c.updater.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
					ChangeBatch:  batch.changeBatch,
					HostedZoneId: aws.String(batch.zoneID),
				})
				if err != nil {
					return fmt.Errorf("unable to update records for zone-id %s, err:%w", batch.zoneID, err)
				}
			}
		}
	}

	return nil
}

func (c *Client) records(ctx context.Context) (map[string][]net.IP, error) {
	records := make(map[string][]net.IP, 0)

	for _, zone := range c.managedZones {
		input := &route53.ListResourceRecordSetsInput{
			HostedZoneId: aws.String(zone.ID),
		}

		output, err := c.updater.ListResourceRecordSets(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, record := range output.ResourceRecordSets {
			for _, mr := range zone.ManagedRecords {
				if mr == *record.Name {
					if (record.Type == types.RRTypeA || record.Type == types.RRTypeAaaa) && len(record.ResourceRecords) > 0 {
						ips := make([]net.IP, 0)
						for _, rr := range record.ResourceRecords {
							ips = append(ips, net.ParseIP(*rr.Value))
						}
						records[mr] = ips
					}
				}
			}
		}
	}

	if len(records) > 0 {
		return records, nil
	}

	return nil, nil
}

func (c *Client) ManagedSubdomains() []string {
	subdomains := make([]string, 0)

	for _, zone := range c.managedZones {
		subdomains = append(subdomains, zone.ManagedRecords...)
	}

	return subdomains
}

func sameVersion(ipA net.IP, ipB net.IP) bool {
	if ipA.String() == "" || ipB.String() == "" {
		return false
	}

	validate := validator.New()

	ipAErr := validate.Var(ipA.String(), "required,ipv4")
	ipBErr := validate.Var(ipB.String(), "required,ipv4")
	if ipAErr == nil && ipBErr == nil {
		return true
	}

	ipAErr = validate.Var(ipA.String(), "required,ipv6")
	ipBErr = validate.Var(ipB.String(), "required,ipv6")
	if ipAErr == nil && ipBErr == nil {
		return true
	}

	return false
}

func ipType(ip net.IP) types.RRType {
	if ip.To4() != nil {
		return types.RRTypeA
	}

	return types.RRTypeAaaa
}

func createChange(fqdn string, ip net.IP) types.Change {
	return types.Change{
		Action: types.ChangeActionUpsert,
		ResourceRecordSet: &types.ResourceRecordSet{
			Name: aws.String(fqdn),
			Type: ipType(ip),
			ResourceRecords: []types.ResourceRecord{{
				Value: aws.String(ip.String()),
			}},
			TTL: aws.Int64(300),
		},
	}
}
