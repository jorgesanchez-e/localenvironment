package sqlite

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"path"
	"testing"
	"time"

	"github.com/jorgesanchez-e/localenvironment/apps/acme/internal/domain/certs"
	"github.com/stretchr/testify/assert"
)

const filePath string = "test.db"

/*

func setValues(load certificate, cert *certs.Cert) error {
	if err := cert.SetCertificate(load.Certificate); err != nil {
		return fmt.Errorf("sqlite: unable to set the certificate data, domain: %q, %w", cert.Domain(), err)
	}

	if err := cert.SetPrivateKey(load.PK); err != nil {
		return fmt.Errorf("sqlite: unable to set the pk data, domain: %q, %w", cert.Domain(), err)
	}

	if err := cert.SetIssuerCertificate(load.Issuer); err != nil {
		return fmt.Errorf("sqlite: unable to set the issuer data, domain: %q, %w", cert.Domain(), err)
	}

	if err := cert.SetCSR(load.CSR); err != nil {
		return fmt.Errorf("sqlite: unable to set the csr data, domain: %q, %w", cert.Domain(), err)
	}

	return nil
}
*/

func TestSetValues(t *testing.T) {
	tnow := time.Now().UTC()
	testCases := []struct {
		testName             string
		loaded               func(t *testing.T, domain string) certificate
		expectedErrorContain string
	}{
		{
			testName: "successfull-test",
			loaded: func(t *testing.T, domain string) certificate {
				cert := certificate{}
				cert.Domain = domain
				cert.Certificate = generateTestPEMByType(t, "CERTIFICATE", domain)
				cert.PK = generateTestPEMByType(t, "PRIVATE KEY", domain)
				cert.Issuer = generateTestPEMByType(t, "CERTIFICATE", domain)
				cert.CSR = generateTestPEMByType(t, "CERTIFICATE REQUEST", domain)
				cert.UpdateAt = tnow
				return cert
			},
		},
		{
			testName: "cert-error-test",
			loaded: func(t *testing.T, domain string) certificate {
				cert := certificate{}
				cert.Domain = domain
				cert.Certificate = generateTestPEMByType(t, "PRIVATE KEY", domain)
				cert.PK = generateTestPEMByType(t, "PRIVATE KEY", domain)
				cert.Issuer = generateTestPEMByType(t, "CERTIFICATE", domain)
				cert.CSR = generateTestPEMByType(t, "CERTIFICATE REQUEST", domain)
				cert.UpdateAt = tnow
				return cert
			},
			expectedErrorContain: "sqlite: unable to set the certificate data, domain: \"mydomain.io\", invalid certificate",
		},
		{
			testName: "private-key-error-test",
			loaded: func(t *testing.T, domain string) certificate {
				cert := certificate{}
				cert.Domain = domain
				cert.Certificate = generateTestPEMByType(t, "CERTIFICATE", domain)
				cert.PK = generateTestPEMByType(t, "CERTIFICATE", domain)
				cert.UpdateAt = tnow
				return cert
			},
			expectedErrorContain: "sqlite: unable to set the pk data, domain: \"mydomain.io\", invalid private key",
		},
		{
			testName: "issuer-error-test",
			loaded: func(t *testing.T, domain string) certificate {
				cert := certificate{}
				cert.Domain = domain
				cert.Certificate = generateTestPEMByType(t, "CERTIFICATE", domain)
				cert.PK = generateTestPEMByType(t, "PRIVATE KEY", domain)
				cert.Issuer = generateTestPEMByType(t, "PRIVATE KEY", domain)
				cert.UpdateAt = tnow
				return cert
			},
			expectedErrorContain: "sqlite: unable to set the issuer data, domain: \"mydomain.io\", invalid certificate",
		},
		{
			testName: "invalid-csr-test",
			loaded: func(t *testing.T, domain string) certificate {
				cert := certificate{}
				cert.Domain = domain
				cert.Certificate = generateTestPEMByType(t, "CERTIFICATE", domain)
				cert.PK = generateTestPEMByType(t, "PRIVATE KEY", domain)
				cert.Issuer = generateTestPEMByType(t, "CERTIFICATE", domain)
				cert.CSR = generateTestPEMByType(t, "CERTIFICATE", domain)
				cert.UpdateAt = tnow
				return cert
			},
			expectedErrorContain: "sqlite: unable to set the csr data, domain: \"mydomain.io\", invalid PEM format",
		},
	}

	for _, tc := range testCases {
		loaded := tc.loaded(t, "mydomain.io")
		expectedError := tc.expectedErrorContain

		t.Run(tc.testName, func(t *testing.T) {
			result := new(certs.Cert)
			err := setValues(loaded, result)
			if err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, expectedError)
			}
		})
	}
}

func TestOperations(t *testing.T) {
	t.Run("save-error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		fullpath := path.Join(dir, filePath)

		ctx := context.Background()
		driver, err := New(ctx, fullpath)
		defer func() {
			driver.db.Close()
		}()
		assert.NoError(t, err)
		assert.NotNil(t, driver)

		input := certs.NewCert("mydomain.io")
		assert.NoError(t, fillCert(t, input))

		ctxCanceled, cancel := context.WithCancel(ctx)
		cancel()

		err = driver.Save(ctxCanceled, input)
		assert.ErrorContains(t, err, "sqlite: unable to save certificate, domain: \"mydomain.io\": context canceled")
	})

	t.Run("load-empty", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		fullpath := path.Join(dir, filePath)

		ctx := context.Background()
		driver, err := New(ctx, fullpath)
		defer func() {
			driver.db.Close()
		}()
		assert.NoError(t, err)
		assert.NotNil(t, driver)

		output := certs.NewCert("mydomain.io")
		err = driver.Load(ctx, output)
		assert.NoError(t, err)
	})

	t.Run("load-error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		fullpath := path.Join(dir, filePath)

		ctx := context.Background()
		driver, err := New(ctx, fullpath)
		defer func() {
			driver.db.Close()
		}()
		assert.NoError(t, err)
		assert.NotNil(t, driver)

		ctxCanceled, cancel := context.WithCancel(ctx)
		cancel()

		output := certs.NewCert("mydomain.io")
		err = driver.Load(ctxCanceled, output)
		assert.ErrorContains(t, err, "sqlite: unable to retrieve certification info, domain: \"mydomain.io\", context canceled")
	})

	t.Run("save-and-load-successfull", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		fullpath := path.Join(dir, filePath)

		ctx := context.Background()
		driver, err := New(ctx, fullpath)
		defer func() {
			driver.db.Close()
		}()
		assert.NoError(t, err)
		assert.NotNil(t, driver)

		input := certs.NewCert("mydomain.io")
		assert.NoError(t, fillCert(t, input))

		err = driver.Save(ctx, input)
		assert.NoError(t, err)

		output := certs.NewCert("mydomain.io")
		err = driver.Load(ctx, output)
		assert.NoError(t, err)

		assert.Equal(t, output.Domain(), input.Domain())
		assert.Equal(t, output.Certificate(), input.Certificate())
		assert.Equal(t, output.PrivateKey(), input.PrivateKey())
		assert.Equal(t, output.Issuer(), input.Issuer())
		assert.Equal(t, output.CSR(), input.CSR())
	})
}

func generateTestPEMByType(t *testing.T, blockType string, domain string) []byte {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("unable to generate primary key, type:%q", blockType)
	}

	switch blockType {
	case "CERTIFICATE":
		tpl := &x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject:      pkix.Name{CommonName: domain},
			NotBefore:    time.Now().Add(-time.Hour),
			NotAfter:     time.Now().Add(time.Hour),
			DNSNames:     []string{domain},
		}

		der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &priv.PublicKey, priv)
		if err != nil {
			t.Fatalf("unable to create the certificate, type:%q", blockType)
		}

		return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	case "CERTIFICATE REQUEST", "NEW CERTIFICATE REQUEST":
		tpl := &x509.CertificateRequest{
			Subject:  pkix.Name{CommonName: domain},
			DNSNames: []string{domain},
		}

		der, err := x509.CreateCertificateRequest(rand.Reader, tpl, priv)
		if err != nil {
			t.Fatalf("unable to create the certificate, type:%q", blockType)
		}

		return pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})

	case "RSA PRIVATE KEY":
		return pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		})

	case "EC PRIVATE KEY":
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("unable to generate primary key, type:%q", blockType)
		}

		der, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			t.Fatalf("unable to marshal primary key, type:%q", blockType)
		}

		return pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: der,
		})

	case "PRIVATE KEY":
		der, err := x509.MarshalPKCS8PrivateKey(priv)
		if err != nil {
			t.Fatalf("unable to marshal primary key, type:%q", blockType)
		}

		return pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: der,
		})

	default:
		t.Fatalf("block type no soportado, type:%q", blockType)
	}

	return nil
}

func fillCert(t *testing.T, cert *certs.Cert) error {
	t.Helper()

	if err := cert.SetCertificate(generateTestPEMByType(t, "CERTIFICATE", cert.Domain())); err != nil {
		return err
	}

	if err := cert.SetIssuerCertificate(generateTestPEMByType(t, "CERTIFICATE", cert.Domain())); err != nil {
		return err
	}

	if err := cert.SetPrivateKey(generateTestPEMByType(t, "PRIVATE KEY", cert.Domain())); err != nil {
		return err
	}

	if err := cert.SetCSR(generateTestPEMByType(t, "CERTIFICATE REQUEST", cert.Domain())); err != nil {
		return err
	}

	return nil
}
