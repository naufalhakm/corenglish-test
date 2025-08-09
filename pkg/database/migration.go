package database

import (
	"database/sql"
	"fmt"
	"go-corenglish/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// RunMigrations connects to the database and applies any pending database migrations.
func RunMigrations(cfg *config.Config, log *logrus.Logger) error {
	db, err := sql.Open("postgres", cfg.DatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to connect to database for migrations: %w", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://pkg/database/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	log.Info("Running database migrations...")
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Info("No new migrations to apply.")
	} else {
		log.Info("Database migrations applied successfully!")
	}

	return nil
}
