package r53

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/jorgesanchez-e/localenvironment/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUpdateGetter struct {
	listOutput  *route53.ListResourceRecordSetsOutput
	listErr     error
	changeErr   error
	changeCalls int
}

func (m *mockUpdateGetter) ListResourceRecordSets(
	context.Context,
	*route53.ListResourceRecordSetsInput,
	...func(*route53.Options),
) (*route53.ListResourceRecordSetsOutput, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.listOutput != nil {
		return m.listOutput, nil
	}
	return &route53.ListResourceRecordSetsOutput{}, nil
}

func (m *mockUpdateGetter) ChangeResourceRecordSets(
	context.Context,
	*route53.ChangeResourceRecordSetsInput,
	...func(*route53.Options),
) (*route53.ChangeResourceRecordSetsOutput, error) {
	m.changeCalls++
	if m.changeErr != nil {
		return nil, m.changeErr
	}
	return &route53.ChangeResourceRecordSetsOutput{}, nil
}

func TestGetRecords(t *testing.T) {
	listErr := errors.New("list resource record sets failed")

	testCases := []struct {
		name            string
		domains         []string
		zones           []zone
		listOutput      *route53.ListResourceRecordSetsOutput
		listErr         error
		expectedRecords []Record
		expectedError   error
	}{
		{
			name: "success",
			domains: []string{
				"example.com",
			},
			zones: []zone{
				{id: "Z123"},
			},
			listOutput: &route53.ListResourceRecordSetsOutput{
				ResourceRecordSets: []types.ResourceRecordSet{
					{
						Name: aws.String("example.com"),
						Type: types.RRTypeA,
						TTL:  aws.Int64(300),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String("192.0.2.1")},
						},
					},
				},
			},
			expectedRecords: []Record{
				{
					FQDN:       "example.com",
					IP:         "192.0.2.1",
					RecordType: "A",
					RecordTTL:  300,
				},
			},
		},
		{
			name: "success AAAA record",
			domains: []string{
				"ipv6.example.com",
			},
			zones: []zone{
				{id: "Z123"},
			},
			listOutput: &route53.ListResourceRecordSetsOutput{
				ResourceRecordSets: []types.ResourceRecordSet{
					{
						Name: aws.String("ipv6.example.com"),
						Type: types.RRTypeAaaa,
						TTL:  aws.Int64(60),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String("2001:db8::1")},
						},
					},
				},
			},
			expectedRecords: []Record{
				{
					FQDN:       "ipv6.example.com",
					IP:         "2001:db8::1",
					RecordType: "AAAA",
					RecordTTL:  60,
				},
			},
		},
		{
			name: "skips non A and AAAA records",
			domains: []string{
				"example.com",
			},
			zones: []zone{
				{id: "Z123"},
			},
			listOutput: &route53.ListResourceRecordSetsOutput{
				ResourceRecordSets: []types.ResourceRecordSet{
					{
						Name: aws.String("example.com"),
						Type: types.RRTypeCname,
						TTL:  aws.Int64(300),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String("target.example.com")},
						},
					},
				},
			},
			expectedRecords: []Record{},
		},
		{
			name: "skips records whose fqdn is not requested",
			domains: []string{
				"wanted.example.com",
			},
			zones: []zone{
				{id: "Z123"},
			},
			listOutput: &route53.ListResourceRecordSetsOutput{
				ResourceRecordSets: []types.ResourceRecordSet{
					{
						Name: aws.String("other.example.com"),
						Type: types.RRTypeA,
						TTL:  aws.Int64(300),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String("192.0.2.2")},
						},
					},
				},
			},
			expectedRecords: []Record{},
		},
		{
			name: "skips invalid resource record sets",
			domains: []string{
				"example.com",
			},
			zones: []zone{
				{id: "Z123"},
			},
			listOutput: &route53.ListResourceRecordSetsOutput{
				ResourceRecordSets: []types.ResourceRecordSet{
					{
						Type: types.RRTypeA,
						TTL:  aws.Int64(300),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String("192.0.2.1")},
						},
					},
					{
						Name: aws.String("example.com"),
						Type: types.RRTypeA,
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String("192.0.2.1")},
						},
					},
					{
						Name:            aws.String("example.com"),
						Type:            types.RRTypeA,
						TTL:             aws.Int64(300),
						ResourceRecords: nil,
					},
				},
			},
			expectedRecords: []Record{},
		},
		{
			name: "list resource record sets error",
			domains: []string{
				"example.com",
			},
			zones: []zone{
				{id: "Z123"},
			},
			listErr:         listErr,
			expectedRecords: nil,
			expectedError:   listErr,
		},
		{
			name:            "no zones configured",
			domains:         []string{"example.com"},
			zones:           nil,
			expectedRecords: []Record{},
		},
	}

	for _, testCase := range testCases {
		name := testCase.name
		domains := testCase.domains
		zones := testCase.zones
		listOutput := testCase.listOutput
		listErr := testCase.listErr
		expectedRecords := testCase.expectedRecords
		expectedError := testCase.expectedError

		t.Run(name, func(t *testing.T) {
			updater := &Updater{
				drivers: []awsClient{
					{
						client: &mockUpdateGetter{
							listOutput: listOutput,
							listErr:    listErr,
						},
						zones: zones,
					},
				},
			}

			records, err := updater.GetRecords(context.Background(), domains)

			assert.Equal(t, expectedRecords, records)
			if expectedError != nil {
				assert.EqualError(t, err, expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateRecords(t *testing.T) {
	changeErr := errors.New("change resource record sets failed")

	testCases := []struct {
		name              string
		records           []Record
		zones             []zone
		changeErr         error
		expectedError     error
		expectedChangeCnt int
	}{
		{
			name: "success",
			records: []Record{
				{
					FQDN:       "example.com",
					IP:         "192.0.2.1",
					RecordType: "A",
					RecordTTL:  300,
				},
			},
			zones: []zone{
				{
					id: "Z123",
					records: []Record{
						{
							FQDN:       "example.com",
							RecordType: "A",
							RecordTTL:  300,
						},
					},
				},
			},
			expectedChangeCnt: 1,
		},
		{
			name: "change resource record sets error",
			records: []Record{
				{
					FQDN:       "example.com",
					IP:         "192.0.2.1",
					RecordType: "A",
					RecordTTL:  300,
				},
			},
			zones: []zone{
				{
					id: "Z123",
					records: []Record{
						{
							FQDN:       "example.com",
							RecordType: "A",
							RecordTTL:  300,
						},
					},
				},
			},
			changeErr:         changeErr,
			expectedChangeCnt: 1,
		},
		{
			name: "no matching zone records",
			records: []Record{
				{
					FQDN:       "other.example.com",
					IP:         "192.0.2.1",
					RecordType: "A",
					RecordTTL:  300,
				},
			},
			zones: []zone{
				{
					id: "Z123",
					records: []Record{
						{
							FQDN:       "example.com",
							RecordType: "A",
							RecordTTL:  300,
						},
					},
				},
			},
			expectedChangeCnt: 1,
		},
		{
			name:              "no zones configured",
			records:           []Record{{FQDN: "example.com", IP: "192.0.2.1", RecordType: "A"}},
			zones:             nil,
			expectedChangeCnt: 0,
		},
	}

	for _, testCase := range testCases {
		name := testCase.name
		records := testCase.records
		zones := testCase.zones
		changeErr := testCase.changeErr
		expectedError := testCase.expectedError
		expectedChangeCnt := testCase.expectedChangeCnt

		t.Run(name, func(t *testing.T) {
			mock := &mockUpdateGetter{
				changeErr: changeErr,
			}
			updater := &Updater{
				drivers: []awsClient{
					{
						client: mock,
						zones:  zones,
					},
				},
			}

			err := updater.UpdateRecords(context.Background(), records)

			if expectedError != nil {
				assert.EqualError(t, err, expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, expectedChangeCnt, mock.changeCalls)
		})
	}
}

func TestNewR53(t *testing.T) {
	accounts := []config.AWSConfig{
		{
			AccountName: "accountone",
			Region:      "us-east-1",
			AccessKey:   "AKIAEXAMPLE1",
			SecretKey:   "secret1",
			Zones: []config.ZoneConfig{
				{
					ID:   "Z111",
					Name: "example.com zone",
					Records: []config.RecordConfig{
						{
							FQDN:       "vpn.example.com",
							RecordType: "A",
							RecordTTL:  300,
						},
						{
							FQDN:       "ipv6.example.com",
							RecordType: "AAAA",
							RecordTTL:  60,
						},
					},
				},
			},
		},
		{
			AccountName: "accounttwo",
			Region:      "us-west-2",
			AccessKey:   "AKIAEXAMPLE2",
			SecretKey:   "secret2",
			Zones: []config.ZoneConfig{
				{
					ID:   "Z222",
					Name: "example.org zone",
					Records: []config.RecordConfig{
						{
							FQDN:       "app.example.org",
							RecordType: "A",
							RecordTTL:  120,
						},
					},
				},
			},
		},
	}

	updater := NewR53(accounts)

	require.NotNil(t, updater)
	require.Len(t, updater.drivers, len(accounts))

	assert.Equal(t, []zone{
		{
			id: "Z111",
			records: []Record{
				{FQDN: "vpn.example.com", RecordType: "A", RecordTTL: 300},
				{FQDN: "ipv6.example.com", RecordType: "AAAA", RecordTTL: 60},
			},
		},
	}, updater.drivers[0].zones)
	assert.NotNil(t, updater.drivers[0].client)

	assert.Equal(t, []zone{
		{
			id: "Z222",
			records: []Record{
				{FQDN: "app.example.org", RecordType: "A", RecordTTL: 120},
			},
		},
	}, updater.drivers[1].zones)
	assert.NotNil(t, updater.drivers[1].client)
}
