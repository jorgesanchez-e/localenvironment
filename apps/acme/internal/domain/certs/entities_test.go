package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCert(t *testing.T) {
	testCases := []struct {
		testName     string
		domain       string
		expectedCert *Cert
	}{
		{
			testName:     "Test with example.com",
			domain:       "example.com",
			expectedCert: &Cert{domain: "example.com"},
		},
		{
			testName:     "Test with subdomain.example.com",
			domain:       "subdomain.example.com",
			expectedCert: &Cert{domain: "subdomain.example.com"},
		},
	}

	for _, tc := range testCases {
		cert := NewCert(tc.domain)
		expectedCert := tc.expectedCert

		t.Run(tc.testName, func(t *testing.T) {
			assert.Equal(t, expectedCert, cert, "Expected Cert does not match the actual Cert")
		})
	}
}

func TestIsEmpty(t *testing.T) {
	testCases := []struct {
		testName      string
		cert          Cert
		expectedEmpty bool
	}{
		{
			testName:      "Test with new cert without data",
			cert:          Cert{domain: "example.com"},
			expectedEmpty: true,
		},
		{
			testName:      "Test without private key",
			cert:          Cert{domain: "example.com", certificate: []byte("cert"), issuerCertificate: []byte("issuer")},
			expectedEmpty: true,
		},
		{
			testName:      "Test without certificate",
			cert:          Cert{domain: "example.com", privateKey: []byte("pk"), issuerCertificate: []byte("issuer")},
			expectedEmpty: true,
		},
		{
			testName:      "Test without issuer certificate",
			cert:          Cert{domain: "example.com", privateKey: []byte("pk"), certificate: []byte("cert")},
			expectedEmpty: true,
		},
		{
			testName:      "Test with all data",
			cert:          Cert{domain: "example.com", privateKey: []byte("pk"), certificate: []byte("cert"), issuerCertificate: []byte("issuer")},
			expectedEmpty: false,
		},
	}

	for _, tc := range testCases {
		isEmpty := tc.cert.IsEmpty()
		expectedEmpty := tc.expectedEmpty

		t.Run(tc.testName, func(t *testing.T) {
			assert.Equal(t, expectedEmpty, isEmpty, "Expected IsEmpty result does not match the actual result")
		})
	}
}

func TestDomain(t *testing.T) {
	testCases := []struct {
		testName       string
		cert           *Cert
		expectedDomain string
	}{
		{
			testName:       "Test with example.com",
			cert:           NewCert("example.com"),
			expectedDomain: "example.com",
		},
		{
			testName:       "Test with subdomain.example.com",
			cert:           NewCert("subdomain.example.com"),
			expectedDomain: "subdomain.example.com",
		},
		{
			testName:       "Test with empty domain",
			cert:           NewCert(""),
			expectedDomain: "",
		},
	}

	for _, tc := range testCases {
		domain := tc.cert.Domain()
		expectedDomain := tc.expectedDomain

		t.Run(tc.testName, func(t *testing.T) {
			assert.Equal(t, expectedDomain, domain, "Expected Domain does not match the actual Domain")
		})
	}
}

func TestSetCertificate(t *testing.T) {
	validCert := generateCertPEM(t, "example.com", time.Now().Add(60*24*time.Hour))

	testCases := []struct {
		testName     string
		certificate  []byte
		expectedErr  error
		expectedCert *Cert
	}{
		{
			testName:     "Test with valid certificate",
			certificate:  validCert,
			expectedErr:  nil,
			expectedCert: &Cert{domain: "example.com", certificate: validCert},
		},
		{
			testName:     "Test with invalid PEM",
			certificate:  []byte("not a pem"),
			expectedErr:  ErrInvalidCert,
			expectedCert: &Cert{domain: "example.com"},
		},
		{
			testName:     "Test with PEM of wrong type",
			certificate:  pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("data")}),
			expectedErr:  ErrInvalidCert,
			expectedCert: &Cert{domain: "example.com"},
		},
		{
			testName:     "Test with empty certificate",
			certificate:  nil,
			expectedErr:  ErrInvalidCert,
			expectedCert: &Cert{domain: "example.com"},
		},
	}

	for _, tc := range testCases {
		cert := NewCert("example.com")
		err := cert.SetCertificate(tc.certificate)
		expectedErr := tc.expectedErr
		expectedCert := tc.expectedCert

		t.Run(tc.testName, func(t *testing.T) {
			assert.ErrorIs(t, err, expectedErr, "Expected error does not match the actual error")
			assert.Equal(t, expectedCert, cert, "Expected Cert does not match the actual Cert")
		})
	}
}

func TestSetPrivateKey(t *testing.T) {
	validPK := generateRSAKeyPEM(t)

	testCases := []struct {
		testName     string
		privateKey   []byte
		expectedErr  error
		expectedCert *Cert
	}{
		{
			testName:     "Test with valid RSA private key",
			privateKey:   validPK,
			expectedErr:  nil,
			expectedCert: &Cert{domain: "example.com", privateKey: validPK},
		},
		{
			testName:     "Test with invalid PEM",
			privateKey:   []byte("not a pem"),
			expectedErr:  ErrInvalidPK,
			expectedCert: &Cert{domain: "example.com"},
		},
		{
			testName:     "Test with PEM of wrong type",
			privateKey:   pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: []byte("data")}),
			expectedErr:  ErrInvalidPK,
			expectedCert: &Cert{domain: "example.com"},
		},
		{
			testName:     "Test with empty private key",
			privateKey:   nil,
			expectedErr:  ErrInvalidPK,
			expectedCert: &Cert{domain: "example.com"},
		},
	}

	for _, tc := range testCases {
		cert := NewCert("example.com")
		err := cert.SetPrivateKey(tc.privateKey)
		expectedErr := tc.expectedErr
		expectedCert := tc.expectedCert

		t.Run(tc.testName, func(t *testing.T) {
			assert.ErrorIs(t, err, expectedErr, "Expected error does not match the actual error")
			assert.Equal(t, expectedCert, cert, "Expected Cert does not match the actual Cert")
		})
	}
}

func TestSetIssuerCertificate(t *testing.T) {
	validIssuer := generateCertPEM(t, "issuer.example.com", time.Now().Add(365*24*time.Hour))

	testCases := []struct {
		testName     string
		issuerCert   []byte
		expectedErr  error
		expectedCert *Cert
	}{
		{
			testName:     "Test with valid issuer certificate",
			issuerCert:   validIssuer,
			expectedErr:  nil,
			expectedCert: &Cert{domain: "example.com", issuerCertificate: validIssuer},
		},
		{
			testName:     "Test with invalid PEM",
			issuerCert:   []byte("not a pem"),
			expectedErr:  ErrInvalidCert,
			expectedCert: &Cert{domain: "example.com"},
		},
		{
			testName:     "Test with PEM of wrong type",
			issuerCert:   pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("data")}),
			expectedErr:  ErrInvalidCert,
			expectedCert: &Cert{domain: "example.com"},
		},
		{
			testName:     "Test with empty issuer certificate",
			issuerCert:   nil,
			expectedErr:  ErrInvalidCert,
			expectedCert: &Cert{domain: "example.com"},
		},
	}

	for _, tc := range testCases {
		cert := NewCert("example.com")
		err := cert.SetIssuerCertificate(tc.issuerCert)
		expectedErr := tc.expectedErr
		expectedCert := tc.expectedCert

		t.Run(tc.testName, func(t *testing.T) {
			assert.ErrorIs(t, err, expectedErr, "Expected error does not match the actual error")
			assert.Equal(t, expectedCert, cert, "Expected Cert does not match the actual Cert")
		})
	}
}

func TestSetCSR(t *testing.T) {
	validCSR := generateCSRPEM(t, "example.com")

	testCases := []struct {
		testName     string
		csr          []byte
		expectedErr  error
		expectedCert *Cert
	}{
		{
			testName:     "Test with valid CSR",
			csr:          validCSR,
			expectedErr:  nil,
			expectedCert: &Cert{domain: "example.com", csr: validCSR},
		},
		{
			testName:     "Test with invalid PEM",
			csr:          []byte("not a pem"),
			expectedErr:  ErrInvalidPEMFormat,
			expectedCert: &Cert{domain: "example.com"},
		},
		{
			testName:     "Test with PEM of wrong type",
			csr:          pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("data")}),
			expectedErr:  ErrInvalidPEMFormat,
			expectedCert: &Cert{domain: "example.com"},
		},
		{
			testName:     "Test with empty CSR",
			csr:          nil,
			expectedErr:  nil,
			expectedCert: &Cert{domain: "example.com"},
		},
	}

	for _, tc := range testCases {
		cert := NewCert("example.com")
		err := cert.SetCSR(tc.csr)
		expectedErr := tc.expectedErr
		expectedCert := tc.expectedCert

		t.Run(tc.testName, func(t *testing.T) {
			assert.ErrorIs(t, err, expectedErr, "Expected error does not match the actual error")
			assert.Equal(t, expectedCert, cert, "Expected Cert does not match the actual Cert")
		})
	}
}

func TestNeedsRenewal(t *testing.T) {
	uint8Ptr := func(v uint8) *uint8 { return &v }

	testCases := []struct {
		testName        string
		cert            *Cert
		daysBeforeRenew *uint8
		expectedRenew   bool
		expectedErr     error
	}{
		{
			testName:        "Test with empty cert",
			cert:            NewCert("example.com"),
			daysBeforeRenew: nil,
			expectedRenew:   false,
			expectedErr:     ErrEmptyCert,
		},
		{
			testName:        "Test with invalid PEM",
			cert:            certWithData("example.com", []byte("not a pem")),
			daysBeforeRenew: nil,
			expectedRenew:   false,
			expectedErr:     ErrInvalidCert,
		},
		{
			testName:        "Test with PEM of wrong type",
			cert:            certWithData("example.com", pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("data")})),
			daysBeforeRenew: nil,
			expectedRenew:   false,
			expectedErr:     ErrInvalidCert,
		},
		{
			testName:        "Test with unparseable certificate",
			cert:            certWithData("example.com", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("garbage")})),
			daysBeforeRenew: nil,
			expectedRenew:   false,
			expectedErr:     ErrParseCert,
		},
		{
			testName:        "Test with cert far from expiration and default toleration",
			cert:            certWithData("example.com", generateCertPEM(t, "example.com", time.Now().Add(60*24*time.Hour))),
			daysBeforeRenew: nil,
			expectedRenew:   false,
			expectedErr:     nil,
		},
		{
			testName:        "Test with cert inside default toleration",
			cert:            certWithData("example.com", generateCertPEM(t, "example.com", time.Now().Add(2*24*time.Hour))),
			daysBeforeRenew: nil,
			expectedRenew:   true,
			expectedErr:     nil,
		},
		{
			testName:        "Test with cert outside default toleration",
			cert:            certWithData("example.com", generateCertPEM(t, "example.com", time.Now().Add(4*24*time.Hour))),
			daysBeforeRenew: nil,
			expectedRenew:   false,
			expectedErr:     nil,
		},
		{
			testName:        "Test with expired cert",
			cert:            certWithData("example.com", generateCertPEM(t, "example.com", time.Now().Add(-24*time.Hour))),
			daysBeforeRenew: nil,
			expectedRenew:   true,
			expectedErr:     nil,
		},
		{
			testName:        "Test with cert inside custom toleration",
			cert:            certWithData("example.com", generateCertPEM(t, "example.com", time.Now().Add(20*24*time.Hour))),
			daysBeforeRenew: uint8Ptr(30),
			expectedRenew:   true,
			expectedErr:     nil,
		},
		{
			testName:        "Test with cert outside custom toleration",
			cert:            certWithData("example.com", generateCertPEM(t, "example.com", time.Now().Add(20*24*time.Hour))),
			daysBeforeRenew: uint8Ptr(10),
			expectedRenew:   false,
			expectedErr:     nil,
		},
		{
			// When tolerationDays is set to 0, the used toleration is set to default one, 3 days.
			testName:        "Test with zero toleration and valid cert",
			cert:            certWithData("example.com", generateCertPEM(t, "example.com", time.Now().Add(24*time.Hour))),
			daysBeforeRenew: uint8Ptr(0),
			expectedRenew:   true,
			expectedErr:     nil,
		},
	}

	for _, tc := range testCases {
		expectedRenew := tc.expectedRenew
		expectedErr := tc.expectedErr

		t.Run(tc.testName, func(t *testing.T) {
			renew, err := tc.cert.NeedsRenewal(tc.daysBeforeRenew)

			assert.ErrorIs(t, err, expectedErr, "Expected error does not match the actual error")
			assert.Equal(t, expectedRenew, renew, "Expected NeedsRenewal result does not match the actual result")
		})
	}
}

// certWithData builds a Cert with all the required data. The certificate is
// assigned directly (without SetCertificate) so NeedsRenewal can be tested
// with invalid certificates that the setter would reject.
func certWithData(domain string, certificate []byte) *Cert {
	cert := NewCert(domain)
	cert.privateKey = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: []byte("pk")})
	cert.issuerCertificate = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("issuer")})
	cert.certificate = certificate

	return cert
}

func generateCertPEM(t *testing.T, domain string, notAfter time.Time) []byte {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: domain},
		DNSNames:     []string{domain},
		NotBefore:    notAfter.Add(-90 * 24 * time.Hour),
		NotAfter:     notAfter,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func generateRSAKeyPEM(t *testing.T) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}

func generateCSRPEM(t *testing.T, domain string) []byte {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	template := &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: domain},
		DNSNames: []string{domain},
	}

	der, err := x509.CreateCertificateRequest(rand.Reader, template, key)
	if err != nil {
		t.Fatalf("failed to create CSR: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der})
}
