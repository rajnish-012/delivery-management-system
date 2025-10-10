
package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

// InitPostgres initializes the PostgreSQL connection pool
func InitPostgres(ctx context.Context) error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/delivery?sslmode=disable"
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create pool: %w", err)
	}

	// quick ping to check connection
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return fmt.Errorf("ping failed: %w", err)
	}

	Pool = pool
	return nil
}

// ClosePostgres closes the PostgreSQL connection pool
func ClosePostgres() {
	if Pool != nil {
		Pool.Close()
	}
}

// ExecMigration executes a raw SQL migration string
func ExecMigration(ctx context.Context, sql string) error {
	_, err := Pool.Exec(ctx, sql)
	return err
}

// ExecMigrationFile reads a SQL file and executes its contents
func ExecMigrationFile(ctx context.Context, filepath string) error {
	sqlBytes, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}
	return ExecMigration(ctx, string(sqlBytes))
}
