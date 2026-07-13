package updater

import (
	"context"
	"errors"
	"testing"

	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/domain"
	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/infra/updater/r53"
	"github.com/jorgesanchez-e/localenvironment/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockR53Updater struct {
	getRecordsFn    func(ctx context.Context, domains []string) ([]r53.Record, error)
	updateRecordsFn func(ctx context.Context, records []r53.Record) error
}

func (m *mockR53Updater) GetRecords(ctx context.Context, domains []string) ([]r53.Record, error) {
	if m.getRecordsFn != nil {
		return m.getRecordsFn(ctx, domains)
	}
	return nil, nil
}

func (m *mockR53Updater) UpdateRecords(ctx context.Context, records []r53.Record) error {
	if m.updateRecordsFn != nil {
		return m.updateRecordsFn(ctx, records)
	}
	return nil
}

func TestNewUpdater(t *testing.T) {
	testCases := []struct {
		name            string
		ddnsConfig      *config.SimpleDDNS
		expectedError   error
		expectedDomains []string
	}{
		{
			name:          "nil config",
			ddnsConfig:    nil,
			expectedError: errors.New("ddns config is required"),
		},
		{
			name: "nil aws config",
			ddnsConfig: &config.SimpleDDNS{
				DDNS: config.DDNSConfig{
					AWS: nil,
				},
			},
			expectedError: errors.New("aws config is required"),
		},
		{
			name: "success",
			ddnsConfig: &config.SimpleDDNS{
				DDNS: config.DDNSConfig{
					AWS: []config.AWSConfig{
						{
							AccountName: "example",
							Region:      "us-east-1",
							AccessKey:   "access",
							SecretKey:   "secret",
							Zones: []config.ZoneConfig{
								{
									ID:   "Z123",
									Name: "example.com",
									Records: []config.RecordConfig{
										{FQDN: "vpn.example.com", RecordType: "A", RecordTTL: 300},
										{FQDN: "ipv6.example.com", RecordType: "AAAA", RecordTTL: 60},
									},
								},
							},
						},
					},
				},
			},
			expectedDomains: []string{"vpn.example.com", "ipv6.example.com"},
		},
	}

	for _, testCase := range testCases {
		name := testCase.name
		ddnsConfig := testCase.ddnsConfig
		expectedError := testCase.expectedError
		expectedDomains := testCase.expectedDomains

		t.Run(name, func(t *testing.T) {
			got, err := NewUpdater(ddnsConfig)

			if expectedError != nil {
				assert.Nil(t, got)
				assert.EqualError(t, err, expectedError.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, expectedDomains, got.domains)
			assert.NotNil(t, got.r53Updater)
		})
	}
}

func TestUpdater_GetRecords(t *testing.T) {
	getErr := errors.New("get records failed")

	testCases := []struct {
		name            string
		domains         []string
		mock            *mockR53Updater
		expectedRecords []domain.Record
		expectedError   error
	}{
		{
			name:    "success",
			domains: []string{"vpn.example.com"},
			mock: &mockR53Updater{
				getRecordsFn: func(_ context.Context, domains []string) ([]r53.Record, error) {
					assert.Equal(t, []string{"vpn.example.com"}, domains)
					return []r53.Record{
						{
							FQDN:       "vpn.example.com",
							IP:         "192.0.2.1",
							RecordType: "A",
							RecordTTL:  300,
						},
					}, nil
				},
			},
			expectedRecords: []domain.Record{
				{
					FQDN:   "vpn.example.com",
					IP:     "192.0.2.1",
					IPType: "A",
				},
			},
		},
		{
			name:    "r53 error",
			domains: []string{"vpn.example.com"},
			mock: &mockR53Updater{
				getRecordsFn: func(context.Context, []string) ([]r53.Record, error) {
					return nil, getErr
				},
			},
			expectedRecords: nil,
			expectedError:   getErr,
		},
	}

	for _, testCase := range testCases {
		name := testCase.name
		domains := testCase.domains
		mock := testCase.mock
		expectedRecords := testCase.expectedRecords
		expectedError := testCase.expectedError

		t.Run(name, func(t *testing.T) {
			u := &Updater{
				domains:    domains,
				r53Updater: mock,
			}

			records, err := u.GetRecords(context.Background())

			assert.Equal(t, expectedRecords, records)
			if expectedError != nil {
				assert.EqualError(t, err, expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdater_UpdateRecords(t *testing.T) {
	updateErr := errors.New("update records failed")

	testCases := []struct {
		name          string
		records       []domain.Record
		mock          *mockR53Updater
		expectedError error
	}{
		{
			name:    "empty records",
			records: nil,
			mock: &mockR53Updater{
				updateRecordsFn: func(context.Context, []r53.Record) error {
					t.Fatal("UpdateRecords should not be called for empty input")
					return nil
				},
			},
		},
		{
			name: "success",
			records: []domain.Record{
				{
					FQDN:   "vpn.example.com",
					IP:     "192.0.2.1",
					IPType: "A",
				},
			},
			mock: &mockR53Updater{
				updateRecordsFn: func(_ context.Context, records []r53.Record) error {
					assert.Equal(t, []r53.Record{
						{
							FQDN:       "vpn.example.com",
							IP:         "192.0.2.1",
							RecordType: "A",
						},
					}, records)
					return nil
				},
			},
		},
		{
			name: "r53 error",
			records: []domain.Record{
				{
					FQDN:   "vpn.example.com",
					IP:     "192.0.2.1",
					IPType: "A",
				},
			},
			mock: &mockR53Updater{
				updateRecordsFn: func(context.Context, []r53.Record) error {
					return updateErr
				},
			},
			expectedError: updateErr,
		},
	}

	for _, testCase := range testCases {
		name := testCase.name
		records := testCase.records
		mock := testCase.mock
		expectedError := testCase.expectedError

		t.Run(name, func(t *testing.T) {
			u := &Updater{
				r53Updater: mock,
			}

			err := u.UpdateRecords(context.Background(), records)

			if expectedError != nil {
				assert.EqualError(t, err, expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
