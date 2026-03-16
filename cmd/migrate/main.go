package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var migrations = []struct {
	name string
	file string
}{
	{"create_fields", "migrations/001_create_fields.sql"},
	{"create_readings", "migrations/002_create_readings.sql"},
	{"create_alerts", "migrations/003_create_alerts.sql"},
	{"create_notifications", "migrations/004_create_notifications.sql"},
	{"create_thresholds", "migrations/005_create_thresholds.sql"},
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	_ = godotenv.Load()

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://agristream:agristream@localhost:5432/agristream"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("ping failed", "error", err)
		os.Exit(1)
	}

	for _, m := range migrations {
		sql, err := os.ReadFile(m.file)
		if err != nil {
			logger.Error("read migration failed", "file", m.file, "error", err)
			os.Exit(1)
		}

		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			logger.Error("migration failed", "name", m.name, "error", err)
			os.Exit(1)
		}

		logger.Info("migration applied", "name", m.name)
	}

	fmt.Println("All migrations applied successfully.")
}
