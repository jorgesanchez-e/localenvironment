package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	const file string = "test.db"

	t.Run("database-created-succesfully", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		fullPath := filepath.Join(dir, file)

		t.Cleanup(func() {
			os.Remove(fullPath)
		})

		ctx := context.Background()
		d, err := New(ctx, fullPath)

		assert.NoError(t, err)
		assert.NotNil(t, d)
	})

	t.Run("ping-database-error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skipf("this test is not available on windows")
		}

		if os.Geteuid() == 0 {
			t.Skipf("this test doesn't work if it us running as root")
		}

		t.Parallel()
		ctx := context.Background()

		dir := t.TempDir()
		err := os.Chmod(dir, 0o444)
		assert.NoError(t, err)

		t.Cleanup(func() {
			_ = os.Chmod(dir, 0o755)
		})

		fullPath := filepath.Join(dir, file)
		d, err := New(ctx, fullPath)

		assert.ErrorContains(t, err, "unable to open database file")
		assert.Nil(t, d)
	})
}

func TestMigrate(t *testing.T) {
	newDriver := func(t *testing.T) *Driver {
		t.Helper()
		d, err := New(context.Background(), filepath.Join(t.TempDir(), "test.db"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = d.db.Close() })
		return d
	}

	t.Run("migrate-is-idempotent", func(t *testing.T) {
		t.Parallel()
		d := newDriver(t) // New ya ejecutó migrate una vez

		err := d.migrate(context.Background())
		assert.NoError(t, err)
	})

	t.Run("migrate-error-canceled-context", func(t *testing.T) {
		t.Parallel()
		d := newDriver(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := d.migrate(ctx)
		assert.ErrorIs(t, err, context.Canceled)
		assert.ErrorContains(t, err, "sqlite: migrate")
	})

	t.Run("migrate-error-closed-database", func(t *testing.T) {
		t.Parallel()
		d := newDriver(t)
		require.NoError(t, d.db.Close())

		err := d.migrate(context.Background())
		assert.ErrorContains(t, err, "sql: database is closed")
		assert.ErrorContains(t, err, "sqlite: migrate")
	})
}
