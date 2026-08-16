package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	migrationPath = "internal/infra/postgres/migrations"
)

type Config struct {
	Host     string `env:"HOST"`
	Port     string `env:"PORT"`
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`
	DBName   string `env:"NAME"`
	SSLMode  string `env:"SSLMODE"`
}

func (cfg *Config) buildConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)
}

func runMigrations(cfg Config) error {
	connStr := cfg.buildConnectionString()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("cannot open database for migrations: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("cannot create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationPath), // Path to migration dir
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("cannot initialize migrate instance: %w", err)
	}

	pkg.Logger.Info(fmt.Sprintf("Running database migrations from '%s'", migrationPath))
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	pkg.Logger.Info("Database schema is up to date")
	return nil
}

func NewDBTX(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	connStr := cfg.buildConnectionString()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	pkg.Logger.Info("Successfully connected to the database (pgx pool).")
	return pool, nil
}

func NewQueries(ctx context.Context, db *pgxpool.Pool, cfg Config) *sqlc.Queries {
	if err := runMigrations(cfg); err != nil {
		pkg.Logger.Error("database migration error", "error", err)
	}

	return sqlc.New(db)
}
