package aws

import (
	"context"
	"crypto/rsa"
	"encoding/pem"
	"errors"
	"fmt"
	"testing"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/registration"
	"github.com/go-viper/mapstructure/v2"
	"github.com/jorgesanchez-e/localenvironment/apps/acme/internal/domain/certs"
	"github.com/stretchr/testify/assert"
)

type testConfigurator struct {
	passedKey    string
	value2Return map[string]any
	err          error
}

func (tc *testConfigurator) Get(key string, value any) error {
	if tc.err != nil {
		return tc.err
	}

	return mapstructure.Decode(tc.value2Return, value)
}

type acmeClientMock struct {
	setDNSErr   error
	account     *acme.ExtendedAccount
	registerErr error
	resource    *certificate.Resource
	obtainErr   error
}

func (m *acmeClientMock) SetDNS01Provider(challenge.Provider, ...dns01.ChallengeOption) error {
	return m.setDNSErr
}

func (m *acmeClientMock) Register(context.Context, registration.RegisterOptions) (*acme.ExtendedAccount, error) {
	return m.account, m.registerErr
}

func (m *acmeClientMock) Obtain(context.Context, certificate.ObtainRequest) (*certificate.Resource, error) {
	return m.resource, m.obtainErr
}

func (m *acmeClientMock) Renew(ctx context.Context, certRes certificate.Resource,
	options *certificate.RenewOptions) (*certificate.Resource, error) {
	return m.resource, nil
}

// mockClientBuilder returns a builder compatible with legoClientBuilder that
// always yields the given client or error, without touching the network.
func mockClientBuilder(client acmeClient, err error) func(*acmeUser) func() (acmeClient, error) {
	return func(*acmeUser) func() (acmeClient, error) {
		return func() (acmeClient, error) {
			if err != nil {
				return nil, err
			}
			return client, nil
		}
	}
}

func testPrimaryKey(t *testing.T) string {
	t.Helper()

	pk, err := certcrypto.GeneratePrivateKey(certcrypto.EC256)
	if err != nil {
		t.Fatalf("generating private key: %v", err)
	}

	return string(certcrypto.PEMEncode(pk))
}

func Test_New(t *testing.T) {
	primaryKey := testPrimaryKey(t)
	account := &acme.ExtendedAccount{Location: "https://acme.test/acct/1"}

	testCases := []struct {
		name          string
		testCnf       *testConfigurator
		clientBuilder func(*acmeUser) func() (acmeClient, error)
		expectedError error
	}{
		{
			name: "normal-test",
			testCnf: &testConfigurator{
				value2Return: map[string]any{
					"email":       "admin@myawesomedomain.net",
					"primary_key": primaryKey,
					"aws": []awsConfig{{
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
					}},
				},
			},
			clientBuilder: mockClientBuilder(&acmeClientMock{account: account}, nil),
			expectedError: nil,
		},
		{
			name: "read-config-error",
			testCnf: &testConfigurator{
				err: fmt.Errorf("some-error"),
			},
			expectedError: ErrReadConfig,
		},
		{
			name: "primary-key-error",
			testCnf: &testConfigurator{
				value2Return: map[string]any{
					"email":       "admin@myawesomedomain.net",
					"primary_key": "no-real-primary-key",
					"aws": []awsConfig{{
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
					}},
				},
			},
			expectedError: ErrParsingPrivateKey,
		},
		{
			name: "lego-client-error",
			testCnf: &testConfigurator{
				value2Return: map[string]any{
					"email":       "admin@myawesomedomain.net",
					"primary_key": primaryKey,
					"aws": []awsConfig{{
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
					}},
				},
			},
			clientBuilder: mockClientBuilder(nil, fmt.Errorf("some-client-error")),
			expectedError: ErrAcmeLegoClient,
		},
		{
			name: "register-error",
			testCnf: &testConfigurator{
				value2Return: map[string]any{
					"email":       "admin@myawesomedomain.net",
					"primary_key": primaryKey,
					"aws": []awsConfig{{
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
					}},
				},
			},
			clientBuilder: mockClientBuilder(&acmeClientMock{registerErr: fmt.Errorf("some-register-error")}, nil),
			expectedError: ErrAcmeAccountRegistration,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.clientBuilder != nil {
				original := newClientBuilder
				newClientBuilder = tt.clientBuilder
				t.Cleanup(func() { newClientBuilder = original })
			}

			driver, err := New(context.Background(), tt.testCnf)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, driver)
				return
			}

			assert.NoError(t, err)
			if assert.NotNil(t, driver) {
				assert.Equal(t, "admin@myawesomedomain.net", driver.user.Email)
				assert.Equal(t, account, driver.user.Account)
				assert.Len(t, driver.awsAccounts, 1)
				assert.NotNil(t, driver.clientBuilder)
			}
		})
	}
}

func createCertificates(t *testing.T, domains []string) []*certs.Cert {
	t.Helper()
	c := make([]*certs.Cert, 0)
	for _, d := range domains {
		c = append(c, certs.NewCert(d))
	}

	return c
}

func testResource(t *testing.T, domain string) *certificate.Resource {
	t.Helper()

	key, err := certcrypto.GeneratePrivateKey(certcrypto.RSA2048)
	if err != nil {
		t.Fatalf("generating key: %v", err)
	}

	certPEM, err := certcrypto.GeneratePemCert(key.(*rsa.PrivateKey), domain, nil)
	if err != nil {
		t.Fatalf("generating cert: %v", err)
	}

	csrDER, err := certcrypto.CreateCSR(key, certcrypto.CSROptions{Domain: domain})
	if err != nil {
		t.Fatalf("generating csr: %v", err)
	}

	return &certificate.Resource{
		Domains:           []string{domain},
		PrivateKey:        certcrypto.PEMEncode(key),
		Certificate:       certPEM,
		IssuerCertificate: certPEM,
		CSR:               pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrDER}),
	}
}

func certEncoded(input *certificate.Resource) *certs.Cert {
	newCert := certs.NewCert(input.Domains[0])
	newCert.SetCertificate(input.Certificate)
	newCert.SetPrivateKey(input.PrivateKey)
	newCert.SetIssuerCertificate(input.IssuerCertificate)
	newCert.SetCSR(input.CSR)

	return newCert
}

func Test_Upsert(t *testing.T) {
	primaryKey := testPrimaryKey(t)
	certOK := testResource(t, "www.myawesomedomain.net")
	certOKEncoded := certEncoded(certOK)

	testCases := []struct {
		testName       string
		testCnf        *testConfigurator
		clientBuilder  func(*acmeUser) func() (acmeClient, error)
		input          []*certs.Cert
		expectedResult []*certs.Cert
		expectedError  error
	}{
		{
			testName:       "no-certificates",
			input:          []*certs.Cert{},
			expectedResult: []*certs.Cert{},
			testCnf: &testConfigurator{
				value2Return: map[string]any{
					"email":       "admin@myawesomedomain.net",
					"primary_key": primaryKey,
					"aws": []awsConfig{{
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
					}},
				},
			},
			expectedError: nil,
			clientBuilder: nil,
		},
		{
			testName: "no-credentials-found-error",
			testCnf: &testConfigurator{
				value2Return: map[string]any{
					"email":       "admin@myawesomedomain.net",
					"primary_key": primaryKey,
					"aws": []awsConfig{{
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
					}},
				},
			},
			input:          createCertificates(t, []string{"www.myawesomedomain.org"}),
			expectedResult: []*certs.Cert{},
			clientBuilder:  mockClientBuilder(&acmeClientMock{}, nil),
			expectedError:  ErrAWSCredentialsNotFound,
		},
		{
			testName: "route53-provider-error",
			testCnf: &testConfigurator{
				value2Return: map[string]any{
					"email":       "admin@myawesomedomain.net",
					"primary_key": primaryKey,
					"aws": []awsConfig{{
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
					}},
				},
			},
			input:          createCertificates(t, []string{"www.myawesomedomain.net"}),
			expectedResult: []*certs.Cert{},
			clientBuilder: mockClientBuilder(&acmeClientMock{
				setDNSErr: ErrRoute53ProviderCreation,
			}, nil),
			expectedError: ErrRoute53ProviderCreation,
		},
		{
			testName: "obtain-error",
			testCnf: &testConfigurator{
				value2Return: map[string]any{
					"email":       "admin@myawesomedomain.net",
					"primary_key": primaryKey,
					"aws": []awsConfig{{
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
					}},
				},
			},
			input:          createCertificates(t, []string{"www.myawesomedomain.net"}),
			expectedResult: []*certs.Cert{},
			clientBuilder: mockClientBuilder(&acmeClientMock{
				obtainErr: errors.New("obtain-error"),
			}, nil),
			expectedError: ErrAcmeGettingCertificate,
		},
		{
			testName: "cert-certificate-error",
			testCnf: &testConfigurator{
				value2Return: map[string]any{
					"email":       "admin@myawesomedomain.net",
					"primary_key": primaryKey,
					"aws": []awsConfig{{
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
					}},
				},
			},
			input:          createCertificates(t, []string{"www.myawesomedomain.net"}),
			expectedResult: []*certs.Cert{},
			clientBuilder:  mockClientBuilder(&acmeClientMock{}, nil),
			expectedError:  ErrInvalidCertificate,
		},
		{
			testName: "success-case",
			testCnf: &testConfigurator{
				value2Return: map[string]any{
					"email":       "admin@myawesomedomain.net",
					"primary_key": primaryKey,
					"aws": []awsConfig{{
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
					}},
				},
			},
			input:          createCertificates(t, []string{"www.myawesomedomain.net"}),
			expectedResult: []*certs.Cert{certOKEncoded},
			clientBuilder: mockClientBuilder(&acmeClientMock{
				resource: certOK,
			}, nil),
		},
	}

	for _, tt := range testCases {
		expectedResult := tt.expectedResult
		expectedError := tt.expectedError
		cnf := tt.testCnf
		certificates := tt.input

		t.Run(tt.testName, func(t *testing.T) {
			if tt.clientBuilder != nil {
				original := newClientBuilder
				newClientBuilder = tt.clientBuilder
				t.Cleanup(func() { newClientBuilder = original })
			}

			ctx := context.Background()
			driver, err := New(ctx, cnf)
			if assert.NoError(t, err) == false || assert.NotNil(t, driver) == false {
				t.Fatalf("test failed, err:%s", err)
			}

			err = driver.Upsert(ctx, certificates)

			if expectedError == nil {
				assert.NoError(t, err)
				assert.Equal(t, expectedResult, certificates)
			} else {
				assert.ErrorIs(t, err, expectedError)
			}
		})
	}
}

func Test_setCertificateData(t *testing.T) {
	defaultDomain := "www.awesomedomain.net"
	tr := testResource(t, defaultDomain)

	testCases := []struct {
		testName              string
		expectedError         error
		input                 *certificate.Resource
		expectedOutputBuilder func(input *certificate.Resource) *certs.Cert
	}{
		{
			testName:      "invalid-certificate-error",
			expectedError: ErrInvalidCertificate,
			input: func() *certificate.Resource {
				return nil
			}(),
			expectedOutputBuilder: func(input *certificate.Resource) *certs.Cert {
				return certs.NewCert(defaultDomain)
			},
		},
		{
			testName:      "setting-certificate-error",
			expectedError: ErrSettingCerticiate,
			input: func() *certificate.Resource {
				crTemp := testResource(t, defaultDomain)
				crTemp.Certificate = nil
				return crTemp
			}(),
			expectedOutputBuilder: func(input *certificate.Resource) *certs.Cert {
				return certs.NewCert(defaultDomain)
			},
		},
		{
			testName:      "setting-pk-error",
			expectedError: ErrSettingPK,
			input: func() *certificate.Resource {
				crTemp := *tr
				crTemp.PrivateKey = nil
				return &crTemp
			}(),
			expectedOutputBuilder: func(input *certificate.Resource) *certs.Cert {
				c := certs.NewCert(defaultDomain)
				c.SetCertificate(input.Certificate)
				return c
			},
		},
		{
			testName:      "setting-issuer-error",
			expectedError: ErrSettingIssuer,
			input: func() *certificate.Resource {
				crTemp := *tr
				crTemp.IssuerCertificate = nil
				return &crTemp
			}(),
			expectedOutputBuilder: func(input *certificate.Resource) *certs.Cert {
				c := certs.NewCert(defaultDomain)
				c.SetCertificate(input.Certificate)
				c.SetPrivateKey(input.PrivateKey)
				return c
			},
		},
		{
			testName:      "setting-crs-error",
			expectedError: ErrSettingCSR,
			input: func() *certificate.Resource {
				crTemp := *tr
				crTemp.CSR = tr.PrivateKey
				return &crTemp
			}(),
			expectedOutputBuilder: func(input *certificate.Resource) *certs.Cert {
				c := certs.NewCert(defaultDomain)
				c.SetCertificate(input.Certificate)
				c.SetPrivateKey(input.PrivateKey)
				c.SetIssuerCertificate(input.IssuerCertificate)
				return c
			},
		},
		{
			testName:      "success",
			expectedError: nil,
			input: func() *certificate.Resource {
				crTemp := *tr
				return &crTemp
			}(),
			expectedOutputBuilder: func(input *certificate.Resource) *certs.Cert {
				c := certs.NewCert(defaultDomain)
				c.SetCertificate(input.Certificate)
				c.SetPrivateKey(input.PrivateKey)
				c.SetIssuerCertificate(input.IssuerCertificate)
				c.SetCSR(input.CSR)
				return c
			},
		},
	}

	for _, tt := range testCases {
		expectedOutput := tt.expectedOutputBuilder(tr)
		expectedError := tt.expectedError
		input := tt.input

		t.Run(tt.testName, func(t *testing.T) {
			output := certs.NewCert(defaultDomain)
			err := setCertificateData(output, input)

			if expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, expectedError)
				assert.Equal(t, expectedOutput, output)
			}
		})
	}
}
