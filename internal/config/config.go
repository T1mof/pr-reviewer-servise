package config

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	// Импортируем для регистрации file source драйвера миграций
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	// Импортируем для регистрации PostgreSQL драйвера
	_ "github.com/lib/pq"
)

type Config struct {
	DatabaseURL    string
	Port           string
	MigrationsPath string
	AdminToken     string
}

// Load загружает конфигурацию из переменных окружения.
func Load() *Config {
	cfg := &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/pr_service?sslmode=disable"),
		Port:           getEnv("PORT", "8080"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "file://migrations"),
		AdminToken:     getEnv("ADMIN_TOKEN", "admin-secret"),
	}

	slog.Info("Config loaded",
		"port", cfg.Port,
		"migrations_path", cfg.MigrationsPath,
	)

	return cfg
}

// ConnectDB подключается к БД с retry логикой.
func (c *Config) ConnectDB() (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	maxRetries := 10
	delay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sqlx.Connect("postgres", c.DatabaseURL)
		if err == nil {
			if err = db.Ping(); err == nil {
				db.SetMaxOpenConns(100)
				db.SetMaxIdleConns(25)
				db.SetConnMaxLifetime(5 * time.Minute)

				slog.Info("Successfully connected to database")
				return db, nil
			}
		}

		slog.Warn("Failed to connect to database",
			"attempt", i+1,
			"max_retries", maxRetries,
			"error", err,
		)

		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}

// RunMigrations применяет миграции к БД.
func (c *Config) RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		c.MigrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Migrations applied successfully")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
