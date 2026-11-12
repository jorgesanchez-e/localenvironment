package certs

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"html/template"
	"strings"
	"time"
)

const (
	limitDaysDefault uint8 = 3
)

var (
	ErrInvalidPEMFormat = fmt.Errorf("invalid PEM format")
	ErrInvalidCert      = fmt.Errorf("invalid certificate")
	ErrInvalidPK        = fmt.Errorf("invalid private key")
	ErrEmptyCert        = fmt.Errorf("empty cert")
	ErrParseCert        = fmt.Errorf("failed to parse certificate")
)

type Cert struct {
	domain            string
	certURL           string
	privateKey        []byte
	certificate       []byte
	issuerCertificate []byte
	csr               []byte
}

func NewCert(domain string) *Cert {
	return &Cert{
		domain: domain,
	}
}

func (c *Cert) IsEmpty() bool {
	return len(c.privateKey) == 0 || len(c.certificate) == 0 || len(c.issuerCertificate) == 0
}

func (c *Cert) Domain() string {
	return c.domain
}

func (c *Cert) PrivateKey() []byte {
	return c.privateKey
}

func (c *Cert) Certificate() []byte {
	return c.certificate
}

func (c *Cert) Issuer() []byte {
	return c.issuerCertificate
}

func (c *Cert) CSR() []byte {
	return c.csr
}

func (c *Cert) SetCertificate(cert []byte) error {
	block, _ := pem.Decode(cert)
	if block == nil || block.Type != "CERTIFICATE" {
		return ErrInvalidCert
	}

	c.certificate = cert

	return nil
}

func (c *Cert) SetPrivateKey(pk []byte) error {
	block, _ := pem.Decode(pk)
	if block == nil || (block.Type != "RSA PRIVATE KEY" && block.Type != "PRIVATE KEY") {
		return ErrInvalidPK
	}

	c.privateKey = pk

	return nil
}

func (c *Cert) SetIssuerCertificate(issuerCert []byte) error {
	block, _ := pem.Decode(issuerCert)
	if block == nil || block.Type != "CERTIFICATE" {
		return ErrInvalidCert
	}

	c.issuerCertificate = issuerCert

	return nil
}

func (c *Cert) SetCSR(csr []byte) error {
	if csr == nil {
		return nil
	}

	block, _ := pem.Decode(csr)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return ErrInvalidPEMFormat
	}

	c.csr = csr

	return nil
}

func (c *Cert) NeedsRenewal(daysBefore *uint8) (bool, error) {
	if c.IsEmpty() {
		return false, ErrEmptyCert
	}

	block, _ := pem.Decode(c.certificate)
	if block == nil || block.Type != "CERTIFICATE" {
		return false, ErrInvalidCert
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, ErrParseCert
	}

	if daysBefore == nil || *daysBefore == 0 {
		if daysBefore == nil {
			daysBefore = new(uint8)
		}
		*daysBefore = limitDaysDefault
	}

	limitRenewalDate := cert.NotAfter.Add(-time.Duration(*daysBefore) * 24 * time.Hour)

	if time.Now().Before(limitRenewalDate) {
		return false, nil
	}

	return true, nil
}

func (c *Cert) String() string {
	var text strings.Builder

	view := struct {
		Domain     string
		CertPem    string
		PrivateKey string
	}{
		Domain:     c.domain,
		CertPem:    string(c.certificate),
		PrivateKey: string(c.privateKey),
	}

	template.Must(template.New("cert-pem").Parse(
		`Certificate Info:
		 	Domain: {{.Domain}}
		 	Cert-PEM: 
			{{.CertPem}}
		 	Private-Key: 
			{{.PrivateKey}}`)).Execute(&text, view)

	return text.String()
}
