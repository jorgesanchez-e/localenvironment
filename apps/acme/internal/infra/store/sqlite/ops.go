package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jorgesanchez-e/localenvironment/apps/acme/internal/domain/certs"
	log "github.com/sirupsen/logrus"
)

type certificate struct {
	Domain      string
	Certificate []byte
	PK          []byte
	Issuer      []byte
	CSR         []byte
	UpdateAt    time.Time
}

func (d *Driver) Load(ctx context.Context, cert *certs.Cert) (err error) {
	loadCert := certificate{Domain: cert.Domain()}
	query := "SELECT certificate, private_key, issuer_certificate, csr FROM certs WHERE domain = ?"

	defer func() {
		if err != nil {
			log.Debug()
		}
	}()

	err = d.db.QueryRowContext(ctx, query, loadCert.Domain).
		Scan(
			&loadCert.Certificate,
			&loadCert.PK,
			&loadCert.Issuer,
			&loadCert.CSR,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}

		return fmt.Errorf("sqlite: unable to retrieve certification info, domain: %q, %w", loadCert.Domain, err)
	}

	return setValues(loadCert, cert)
}

func setValues(load certificate, cert *certs.Cert) error {
	if err := cert.SetCertificate(load.Certificate); err != nil {
		return fmt.Errorf("sqlite: unable to set the certificate data, domain: %q, %w", load.Domain, err)
	}

	if err := cert.SetPrivateKey(load.PK); err != nil {
		return fmt.Errorf("sqlite: unable to set the pk data, domain: %q, %w", load.Domain, err)
	}

	if err := cert.SetIssuerCertificate(load.Issuer); err != nil {
		return fmt.Errorf("sqlite: unable to set the issuer data, domain: %q, %w", load.Domain, err)
	}

	if err := cert.SetCSR(load.CSR); err != nil {
		return fmt.Errorf("sqlite: unable to set the csr data, domain: %q, %w", load.Domain, err)
	}

	return nil
}

func (d *Driver) Save(ctx context.Context, cert *certs.Cert) (err error) {
	certToSave := certificate{
		Domain:      cert.Domain(),
		Certificate: cert.Certificate(),
		PK:          cert.PrivateKey(),
		Issuer:      cert.Issuer(),
		CSR:         cert.CSR(),
		UpdateAt:    time.Now().UTC(),
	}

	const query = `
	INSERT INTO certs (domain, certificate, private_key, issuer_certificate, csr, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(domain) DO UPDATE SET
		certificate        = excluded.certificate,
		private_key        = excluded.private_key,
		issuer_certificate = excluded.issuer_certificate,
		csr                = excluded.csr,
		updated_at         = excluded.updated_at;`

	_, err = d.db.ExecContext(ctx, query,
		certToSave.Domain,
		certToSave.Certificate,
		certToSave.PK,
		certToSave.Issuer,
		certToSave.CSR,
		certToSave.UpdateAt,
	)

	if err != nil {
		return fmt.Errorf("sqlite: unable to save certificate, domain: %q: %w", certToSave.Domain, err)
	}

	return nil
}
