package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	validSimpleDDNSYAML = `
ddns:
  log-level: "debug"
  check-every-seconds: 300
  process-timeout-seconds: 20
  aws:
    - account-name: "example"
      region: "us-east-1"
      access-key: "1234567890"
      secret-key: "1234567890"
      zones:
        - id: "Z0123456789ABCDEF"
          name: "example.net zone."
          records:
          - fqdn: "vpn.example.net."
            record-type: A
            record-ttl: 3600
          - fqdn: "localhost.example.net."
            record-type: A
            record-ttl: 3600
          - fqdn: "test-a.example.net."
            record-type: AAAA
            record-ttl: 3600
        - id: "Z0123456789ABCDEG"
          name: "example.org zone."
          records:
          - fqdn: "vpn.example.org."
            record-type: A
            record-ttl: 3600
          - fqdn: "localhost.example.org."
            record-type: A
            record-ttl: 3600
`

	invalidConfigSimpleDDNSYAML = `
ddns:
  log-level: "debug"
  check-every-seconds: 300
  process-timeout-seconds: 20
  aws:
    - region: "us-east-1"
      zones:
        - id: "Z0123456789ABCDEF"
          records:
          - fqdn: "vpn.example.net."
            record-type: A
            record-ttl: 3600
`
)

func TestGetSimpleDDNSConfig(t *testing.T) {
	testCases := []struct {
		name           string
		yaml           string
		expectedConfig *SimpleDDNS
		expectedError  error
	}{
		{
			name: "valid simple DDNS configuration",
			yaml: validSimpleDDNSYAML,
			expectedConfig: &SimpleDDNS{
				DDNS: DDNSConfig{
					LogLevel:             "debug",
					CheckEverySeconds:    300,
					UpdateTimeoutSeconds: 20,
					AWS: []AWSConfig{
						{
							AccountName: "example",
							Region:      "us-east-1",
							AccessKey:   "1234567890",
							SecretKey:   "1234567890",
							Zones: []ZoneConfig{
								{
									ID:   "Z0123456789ABCDEF",
									Name: "example.net zone.",
									Records: []RecordConfig{
										{
											FQDN:       "vpn.example.net.",
											RecordType: "A",
											RecordTTL:  3600,
										},
										{
											FQDN:       "localhost.example.net.",
											RecordType: "A",
											RecordTTL:  3600,
										},
										{
											FQDN:       "test-a.example.net.",
											RecordType: "AAAA",
											RecordTTL:  3600,
										},
									},
								},
								{
									ID:   "Z0123456789ABCDEG",
									Name: "example.org zone.",
									Records: []RecordConfig{
										{
											FQDN:       "vpn.example.org.",
											RecordType: "A",
											RecordTTL:  3600,
										},
										{
											FQDN:       "localhost.example.org.",
											RecordType: "A",
											RecordTTL:  3600,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:           "invalid config simple DDNS configuration",
			yaml:           invalidConfigSimpleDDNSYAML,
			expectedConfig: nil,
			expectedError:  errors.New("invalid config: Key: 'SimpleDDNS.DDNS.AWS[0].AccountName' Error:Field validation for 'AccountName' failed on the 'required' tag"),
		},
	}

	for _, testCase := range testCases {
		name := testCase.name
		expectedError := testCase.expectedError
		expectedConfig := testCase.expectedConfig
		yaml := testCase.yaml

		t.Run(name, func(t *testing.T) {
			withConfigPaths(t, []string{"."})

			dir := t.TempDir()
			writeConfigFile(t, dir, yaml)

			t.Chdir(dir)

			cnf, _ := New()
			ddnsConf, err := cnf.GetSimpleDDNSConfig()

			assert.Equal(t, expectedConfig, ddnsConf)
			if expectedError != nil {
				assert.EqualError(t, err, expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
