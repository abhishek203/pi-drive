package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func Connect(databaseURL string) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &DB{conn}, nil
}

func (d *DB) Migrate(migrationsDir string) error {
	_, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations dir: %w", err)
	}

	var migrations []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".sql") {
			migrations = append(migrations, f.Name())
		}
	}
	sort.Strings(migrations)

	for _, m := range migrations {
		var exists bool
		err := d.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", m).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", m, err)
		}
		if exists {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, m))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", m, err)
		}

		tx, err := d.BeginTx(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("failed to begin tx for %s: %w", m, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", m, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", m); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", m, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", m, err)
		}

		fmt.Printf("  ✓ Applied migration: %s\n", m)
	}

	return nil
}
