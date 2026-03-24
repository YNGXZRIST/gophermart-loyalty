// Package migrations applies embedded SQL migrations.
package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// FS contains embedded SQL migration files.
//
//go:embed *sql
var FS embed.FS

// Migrate applies all up migrations to target database DSN.
func Migrate(dns string) error {
	if dns == "" {
		return fmt.Errorf("database DSN is not set")
	}
	db, err := sql.Open("pgx", dns)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres driver: %w", err)
	}

	srcDriver, err := iofs.New(FS, ".")
	if err != nil {
		return fmt.Errorf("ifs source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", srcDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
