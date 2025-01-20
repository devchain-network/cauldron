package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/vigo/getenv"
)

func runMigration(logger *slog.Logger, dsn string) error {
	logger.Info("running migration")

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("error: [%w]", err)
	}

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("no changes detected")

			return nil
		}

		return fmt.Errorf("migration up error: [%w]", err)
	}

	logger.Info("migration completed")

	return nil
}

func main() {
	databaseURL := getenv.String("DATABASE_URL_MIGRATION", "")
	if err := getenv.Parse(); err != nil {
		log.Fatal(err)
	}

	logger, err := slogger.New(slogger.WithLogLevelName("INFO"))
	if err != nil {
		log.Fatal(err)
	}

	if err = runMigration(logger, *databaseURL); err != nil {
		log.Fatal(err)
	}
}
