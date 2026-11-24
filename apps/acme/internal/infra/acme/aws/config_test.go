package aws

import (
	"testing"

	"github.com/go-viper/mapstructure/v2"
	"github.com/jorgesanchez-e/localenvironment/pkg/config"
	"github.com/stretchr/testify/assert"
)

type configMock struct {
	err  error
	data map[string]any
}

func (cm *configMock) Get(key string, value any) error {
	if cm.err != nil {
		return cm.err
	}

	if key != configKey {
		return config.ErrKeyNotFound
	}

	return mapstructure.Decode(cm.data, value)
}

func Test_ReadConfig(t *testing.T) {
	testCases := []struct {
		name           string
		configurator   *configMock
		expectedConfig *acmeConfig
		expectedError  error
	}{
		{
			name: "normal-test",
			configurator: &configMock{
				err: nil,
				data: map[string]any{
					"email":       "admin@mydomain.com",
					"primary_key": "some-primary-key",
					"aws": []awsConfig{
						{
							Account:   "some-account",
							SecretKey: "some-secret-key",
							Accesskey: "some-access-key",
							Region:    "some-region",
							Domains: []route53DomainConfig{
								{
									FQDN:   "www.myawesomedomain.net",
									ZoneID: "zoneid-01",
								},
							},
						},
					},
				},
			},
			expectedConfig: &acmeConfig{
				Email:      "admin@mydomain.com",
				PrimaryKey: "some-primary-key",
				AWSConf: []awsConfig{
					{
						Account:   "some-account",
						SecretKey: "some-secret-key",
						Accesskey: "some-access-key",
						Region:    "some-region",
						Domains: []route53DomainConfig{
							{
								FQDN:   "www.myawesomedomain.net",
								ZoneID: "zoneid-01",
							},
						},
					},
				},
			},
		},
		{
			name: "validator-error",
			configurator: &configMock{
				err: nil,
				data: map[string]any{
					"email": "",
				},
			},
			expectedError:  ErrInvalidConfigurationValue,
			expectedConfig: nil,
		},
		{
			name: "no-key-found",
			configurator: &configMock{
				err: config.ErrKeyNotFound,
			},
			expectedConfig: nil,
			expectedError:  ErrReadConfig,
		},
	}

	for _, tt := range testCases {
		expectedError := tt.expectedError
		expectedConfig := tt.expectedConfig
		configuratorTest := tt.configurator

		t.Run(tt.name, func(t *testing.T) {
			result, err := readConfig(configuratorTest)

			assert.Equal(t, expectedConfig, result)
			if expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, expectedError)
			}
		})
	}
}

// func validateConfig(data *acmeConfig) error {
func Test_validateConfig(t *testing.T) {
	testCases := []struct {
		name          string
		input         *acmeConfig
		expectedError error
	}{
		{
			name: "normal-test",
			input: &acmeConfig{
				Email:      "admin@myawsomedomain.net",
				PrimaryKey: "primary-key",
				AWSConf: []awsConfig{
					{
						Account:   "some-account",
						Accesskey: "some-access-key",
						Region:    "some-region",
						SecretKey: "some-secret-key",
						Domains: []route53DomainConfig{
							{
								FQDN:   "test.myawsomedomain.net",
								ZoneID: "zone01",
							},
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "error-value-test",
			input: &acmeConfig{
				Email:      "@myawsomedomain.net",
				PrimaryKey: "primary-key",
				AWSConf: []awsConfig{
					{
						Account:   "some-account",
						Accesskey: "some-access-key",
						Region:    "some-region",
						SecretKey: "some-secret-key",
						Domains: []route53DomainConfig{
							{
								FQDN:   "test.myawsomedomain.net",
								ZoneID: "zone01",
							},
						},
					},
				},
			},
			expectedError: ErrInvalidConfigurationValue,
		},
		{
			name:          "nil-config-test",
			input:         nil,
			expectedError: ErrInvalidConfiguration,
		},
	}

	for _, tt := range testCases {
		expectedError := tt.expectedError
		input := tt.input

		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(input)

			if expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, expectedError)
			}
		})
	}
}

// func (aa awsConfig) getCredentials(domain string) (*awsCredentials, error)
func Test_getCredentials(t *testing.T) {
	testCases := []struct {
		testName            string
		input               string
		expectedCredentials *awsCredentials
		expectedError       error
		aa                  awsConfig
	}{
		{
			testName: "normal-test",
			input:    "www.myawesomedomain.com",
			expectedCredentials: &awsCredentials{
				accessKey: "access-key",
				secretKey: "secret-key",
				region:    "region",
				zoneID:    "zone-01",
			},
			aa: awsConfig{
				Account:   "some-account",
				Accesskey: "access-key",
				Region:    "region",
				SecretKey: "secret-key",
				Domains: []route53DomainConfig{
					{
						FQDN:   "www.myawesomedomain.com",
						ZoneID: "zone-01",
					},
				},
			},
		},
		{
			testName:            "credentials-not-found",
			input:               "www2.myawesomedomain.com",
			expectedError:       ErrAWSCredentialsNotFound,
			expectedCredentials: nil,
			aa: awsConfig{
				Account:   "some-account",
				Accesskey: "access-key",
				Region:    "region",
				SecretKey: "secret-key",
				Domains: []route53DomainConfig{
					{
						FQDN:   "www.myawesomedomain.com",
						ZoneID: "zone-01",
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		input := tt.input
		expectedError := tt.expectedError
		expectedCredentials := tt.expectedCredentials
		aa := tt.aa

		t.Run(tt.testName, func(t *testing.T) {
			cred, err := aa.getCredentials(input)

			assert.Equal(t, cred, expectedCredentials)
			if expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, expectedError)
			}
		})
	}
}

//5513276544 -> 0321
//28752550
