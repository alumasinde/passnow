// Package config loads process-wide configuration. There is exactly ONE
// config for the whole platform binary — tenants are DATA in MySQL, not
// separate deployments or separate .env files. Never add a per-tenant
// config loader; tenant-specific behaviour belongs in the tenants/settings
// tables and is read at request time via the resolved tenant context.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	BcryptCost      int

	BaseDomain             string
	PlatformBootstrapToken string
	Env                    string

	GatepassWorkerInterval time.Duration
	ApprovedGatepassTTL    time.Duration
}

func Load() (*Config, error) {
	// Load .env when present. Existing process environment variables win,
	// which keeps production deployments and CI configuration authoritative.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("config: load .env: %w", err)
	}

	cfg := &Config{
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		ReadTimeout:     getDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:     getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: getDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),

		DBHost:            getEnv("DB_HOST", "127.0.0.1"),
		DBPort:            getEnv("DB_PORT", "3306"),
		DBUser:            getEnv("DB_USER", ""),
		DBPassword:        getEnv("DB_PASSWORD", ""),
		DBName:            getEnv("DB_NAME", "gatepass"),
		DBMaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 25),
		DBConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),

		JWTSecret:       getEnv("JWT_SECRET", ""),
		AccessTokenTTL:  getDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		BcryptCost:      getInt("BCRYPT_COST", 12),

		BaseDomain:             getEnv("BASE_DOMAIN", "localhost"),
		PlatformBootstrapToken: getEnv("PLATFORM_BOOTSTRAP_TOKEN", ""),
		Env:                    getEnv("APP_ENV", "development"),

		GatepassWorkerInterval: getDuration("GATEPASS_WORKER_INTERVAL", time.Minute),
		ApprovedGatepassTTL:    getDuration("APPROVED_GATEPASS_TTL", 24*time.Hour),
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" {
		return nil, fmt.Errorf("config: DB_USER and DB_PASSWORD are required")
	}
	if cfg.JWTSecret == "" || len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("config: JWT_SECRET is required and must be >= 32 chars")
	}
	return cfg, nil
}

func (c *Config) MySQLDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=UTC",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
