package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jwt-auth-wallet/internal/config"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type DB struct {
	Pool *pgxpool.Pool
	Host string
}

func Connect(ctx context.Context, cfg *config.Config) (*DB, error) {
	// Candidate hosts to try in order:
	// 1. Host from .env (e.g. localhost, 127.0.0.1, or custom IP)
	// 2. If in Docker container and host is localhost/127.0.0.1, try host.docker.internal (host machine)
	// 3. Fallback to Docker internal service "db" (or "localhost" if running outside Docker)
	candidates := []string{cfg.DBHost}

	inDocker := isRunningInDocker()
	if cfg.DBHost == "localhost" || cfg.DBHost == "127.0.0.1" {
		if inDocker {
			candidates = append(candidates, "host.docker.internal")
		}
		if cfg.DBHost != "db" {
			candidates = append(candidates, "db")
		}
	} else if cfg.DBHost == "db" {
		candidates = append(candidates, "localhost", "127.0.0.1")
	} else {
		// If custom host fails, fallback to db (Docker) or localhost
		if inDocker {
			candidates = append(candidates, "db")
		} else {
			candidates = append(candidates, "localhost")
		}
	}

	var lastErr error
	for _, host := range candidates {
		log.Printf("Attempting database connection to host: %s (database: %s, user: %s)", host, cfg.DBName, cfg.DBUser)

		// 1. Ensure the database exists on this PostgreSQL server (auto-create if missing)
		if err := ensureDatabaseExists(ctx, host, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode); err != nil {
			log.Printf("Note: could not verify/create database %q on host %s: %v", cfg.DBName, host, err)
		}

		// 2. Attempt pool connection to the target database
		connStr := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			url.QueryEscape(cfg.DBUser),
			url.QueryEscape(cfg.DBPassword),
			host,
			cfg.DBPort,
			cfg.DBName,
			cfg.DBSSLMode,
		)

		poolConfig, err := pgxpool.ParseConfig(connStr)
		if err != nil {
			lastErr = fmt.Errorf("unable to parse database config for %s: %w", host, err)
			continue
		}

		poolConfig.MaxConns = 25
		poolConfig.MinConns = 5
		poolConfig.MaxConnLifetime = 1 * time.Hour
		poolConfig.MaxConnIdleTime = 30 * time.Minute

		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			lastErr = fmt.Errorf("unable to connect to %s: %w", host, err)
			continue
		}

		pingCtx, pingCancel := context.WithTimeout(ctx, 3*time.Second)
		err = pool.Ping(pingCtx)
		pingCancel()
		if err != nil {
			pool.Close()
			lastErr = fmt.Errorf("ping failed on %s: %w", host, err)
			log.Printf("Database on %s not ready or credentials rejected: %v", host, err)
			continue
		}

		db := &DB{Pool: pool, Host: host}
		if err := db.Migrate(ctx); err != nil {
			pool.Close()
			lastErr = fmt.Errorf("migration failed on %s: %w", host, err)
			log.Printf("Migration failed on %s: %v", host, err)
			continue
		}

		log.Printf("Successfully connected to database on %s (applied migrations and seeds)", host)
		return db, nil
	}

	return nil, fmt.Errorf("failed to connect to any database candidate (%v): %w", candidates, lastErr)
}

func ensureDatabaseExists(ctx context.Context, host, port, user, password, dbName, sslMode string) error {
	maintCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try connecting to default maintenance database "postgres"
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/postgres?sslmode=%s",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		sslMode,
	)

	conn, err := pgx.Connect(maintCtx, connStr)
	if err != nil {
		// Fallback maintenance database "template1"
		templateConnStr := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/template1?sslmode=%s",
			url.QueryEscape(user),
			url.QueryEscape(password),
			host,
			port,
			sslMode,
		)
		var errTemplate error
		conn, errTemplate = pgx.Connect(maintCtx, templateConnStr)
		if errTemplate != nil {
			return err
		}
	}
	defer conn.Close(maintCtx)

	var exists int
	err = conn.QueryRow(maintCtx, "SELECT 1 FROM pg_database WHERE datname = $1;", dbName).Scan(&exists)
	if err != nil && errors.Is(err, pgx.ErrNoRows) || exists == 0 {
		_, err = conn.Exec(maintCtx, fmt.Sprintf(`CREATE DATABASE "%s";`, dbName))
		if err != nil {
			return fmt.Errorf("CREATE DATABASE %s failed: %w", dbName, err)
		}
		log.Printf("Created database %q on %s", dbName, host)
	}

	return nil
}

func isRunningInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

func (db *DB) Migrate(ctx context.Context) error {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := migrationFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		if _, err := db.Pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
