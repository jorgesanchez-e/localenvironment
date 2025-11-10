package route53

import (
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

const (
	configKey string = "ddns.aws"
)

/*
  aws:
    credentials-file: ""
    access:
      - name: x
        zones:
          - id: Z07083213LQ5Y6BC1F8GR
            records:
              - www.localdomain.dev
              - vpn.localdomain.dev
*/

type r53Config struct {
	CredentialsFile string `yaml:"credentials-file"`
	Zones           []zone `yaml:"zones"`
}

type zone struct {
	ID             string   `yaml:"id"`
	ManagedRecords []string `yaml:"records"`
}

type Batch struct {
	zoneID      string
	changeBatch *types.ChangeBatch
}

func getZoneID(zones []zone, fqdn string) string {
	for _, zone := range zones {
		for _, record := range zone.ManagedRecords {
			if fqdn == record {
				return zone.ID
			}
		}
	}

	return ""
}
