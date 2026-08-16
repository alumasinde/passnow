package migrations

import (
	"database/sql"
	"errors"
	"fmt"

	migrationfiles "gatepass/migrations"
	"gatepass/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const migrationTable = "schema_migrations"

// Up applies all pending forward migrations in version order.
//
// The migration connection is intentionally separate from the normal
// application database pool. It enables MySQL multiStatements because the
// existing migration files contain multiple SQL statements, while the normal
// application DSN keeps multiStatements disabled.
func Up() error {
	m, cleanup, err := newMigrator()
	if err != nil {
		return err
	}
	defer cleanup()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations: apply: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("migrations: read version: %w", err)
	}

	if dirty {
		return fmt.Errorf("migrations: database is dirty at version %d", version)
	}

	if version == 0 {
		fmt.Println("migrations: no migrations applied")
		return nil
	}

	fmt.Printf("migrations: database is at version %d\n", version)
	return nil
}

// Status reports the current migration version without applying changes.
func Status() error {
	m, cleanup, err := newMigrator()
	if err != nil {
		return err
	}
	defer cleanup()

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("migrations: no migrations have been applied")
			return nil
		}
		return fmt.Errorf("migrations: read version: %w", err)
	}

	state := "clean"
	if dirty {
		state = "DIRTY"
	}
	fmt.Printf("migrations: version=%d state=%s\n", version, state)
	return nil
}

func newMigrator() (*migrate.Migrate, func(), error) {
	cfg, err := config.LoadDatabase()
	if err != nil {
		return nil, func() {}, err
	}

	db, err := sql.Open("mysql", cfg.MySQLMigrationDSN())
	if err != nil {
		return nil, func() {}, fmt.Errorf("migrations: open database: %w", err)
	}

	cleanupDB := func() { _ = db.Close() }

	if err := db.Ping(); err != nil {
		cleanupDB()
		return nil, func() {}, fmt.Errorf("migrations: ping database: %w", err)
	}

	dbDriver, err := mysql.WithInstance(db, &mysql.Config{
		MigrationsTable: migrationTable,
	})
	if err != nil {
		cleanupDB()
		return nil, func() {}, fmt.Errorf("migrations: create database driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationfiles.FS, ".")
	if err != nil {
		cleanupDB()
		return nil, func() {}, fmt.Errorf("migrations: create source driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "mysql", dbDriver)
	if err != nil {
		_ = sourceDriver.Close()
		cleanupDB()
		return nil, func() {}, fmt.Errorf("migrations: initialize: %w", err)
	}

	cleanup := func() {
		_, _ = m.Close()
		cleanupDB()
	}
	return m, cleanup, nil
}
