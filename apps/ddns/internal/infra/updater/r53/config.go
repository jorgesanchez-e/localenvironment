package r53

import (
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/jorgesanchez-e/localenvironment/config"
)

type AWSConfig config.AWSConfig

func (c AWSConfig) getCredentials() credentials.StaticCredentialsProvider {
	return credentials.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, "")
}

func (c AWSConfig) zones() []zone {
	zones := make([]zone, 0, len(c.Zones))
	for _, configZone := range c.Zones {
		records := make([]Record, 0, len(configZone.Records))
		for _, record := range configZone.Records {
			records = append(records, Record{
				FQDN:       record.FQDN,
				RecordType: record.RecordType,
				RecordTTL:  record.RecordTTL,
			})
		}
		zones = append(zones, zone{
			id:      configZone.ID,
			records: records,
		})
	}

	return zones
}
