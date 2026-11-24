package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Driver struct {
	db *sql.DB
}

func New(ctx context.Context, file string) (*Driver, error) {
	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)",
		file,
	)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open %q: %w", file, err)
	}

	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: ping %q: %w", file, err)
	}

	driver := &Driver{
		db: db,
	}

	if err := driver.migrate(ctx); err != nil {
		driver.db.Close()
		return nil, err
	}

	return driver, nil
}

func (d *Driver) migrate(ctx context.Context) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS certs (
		domain             TEXT PRIMARY KEY,
		certificate        BLOB,
		private_key        BLOB,
		issuer_certificate BLOB,
		csr                BLOB,
		updated_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("sqlite: migrate: %w", err)
	}

	return nil
}
