// Package config loads process-wide configuration. There is exactly ONE
// config for the whole platform binary — tenants are DATA in MySQL, not
// separate deployments or separate .env files. Never add a per-tenant
// config loader; tenant-specific behaviour belongs in the tenants/settings
// tables and is read at request time via the resolved tenant context.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// HTTP
	HTTPAddr        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	// MySQL — single local server, all tenants share the schema and are
	// isolated logically via tenant_id, not via separate databases.
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	// Auth
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	BcryptCost      int

	// Tenancy / routing
	// BaseDomain is the platform's own domain used for the subdomain and
	// path-prefix fallback, e.g. "gatepass.example.com". Tenants may
	// ADDITIONALLY map their own custom domain (tenants.custom_domain);
	// custom-domain resolution is tried first, then subdomain, then path.
	BaseDomain string

	// PlatformBootstrapToken gates the one-time tenant-provisioning
	// endpoint (see internal/platform). Empty means the endpoint is
	// closed. Set it to provision your first tenant(s), then unset/rotate
	// it — it is not meant to be a long-lived secret.
	PlatformBootstrapToken string

	Env string // "development" | "production"

	// Background workers. Workers are process-wide; tenant-specific business
	// state remains in MySQL and is always filtered by tenant_id.
	GatepassWorkerInterval time.Duration
	ApprovedGatepassTTL    time.Duration
}

func Load() (*Config, error) {
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

		BaseDomain: getEnv("BASE_DOMAIN", "localhost"),

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
	// parseTime=true so DATETIME/TIMESTAMP scan into time.Time.
	// multiStatements intentionally OFF — never allow batched statements
	// on a connection that also carries tenant-scoped queries.
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
