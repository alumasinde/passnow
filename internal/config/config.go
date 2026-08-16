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

// Load loads the complete process configuration used by the HTTP application.
func Load() (*Config, error) {
	cfg := loadDatabaseConfig()

	cfg.HTTPAddr = getEnv("HTTP_ADDR", ":8080")
	cfg.ReadTimeout = getDuration("HTTP_READ_TIMEOUT", 10*time.Second)
	cfg.WriteTimeout = getDuration("HTTP_WRITE_TIMEOUT", 15*time.Second)
	cfg.IdleTimeout = getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	cfg.ShutdownTimeout = getDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second)

	cfg.DBMaxOpenConns = getInt("DB_MAX_OPEN_CONNS", 25)
	cfg.DBMaxIdleConns = getInt("DB_MAX_IDLE_CONNS", 25)
	cfg.DBConnMaxLifetime = getDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)

	cfg.JWTSecret = getEnv("JWT_SECRET", "")
	cfg.AccessTokenTTL = getDuration("ACCESS_TOKEN_TTL", 15*time.Minute)
	cfg.RefreshTokenTTL = getDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour)
	cfg.BcryptCost = getInt("BCRYPT_COST", 12)

	cfg.BaseDomain = getEnv("BASE_DOMAIN", "localhost")
	cfg.PlatformBootstrapToken = getEnv("PLATFORM_BOOTSTRAP_TOKEN", "")
	cfg.Env = getEnv("APP_ENV", "development")
	cfg.GatepassWorkerInterval = getDuration("GATEPASS_WORKER_INTERVAL", time.Minute)
	cfg.ApprovedGatepassTTL = getDuration("APPROVED_GATEPASS_TTL", 24*time.Hour)

	if err := cfg.validateApplication(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadDatabase loads only the database settings required by CLI commands such
// as migrations. It deliberately does not require JWT or HTTP configuration.
// This keeps operational commands independent of application-only secrets.
func LoadDatabase() (*Config, error) {
	cfg := loadDatabaseConfig()
	if cfg.DBUser == "" || cfg.DBPassword == "" {
		return nil, fmt.Errorf("config: DB_USER and DB_PASSWORD are required")
	}
	return cfg, nil
}

func loadDatabaseConfig() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "gatepass"),
	}
}

func (c *Config) validateApplication() error {
	if c.DBUser == "" || c.DBPassword == "" {
		return fmt.Errorf("config: DB_USER and DB_PASSWORD are required")
	}
	if c.JWTSecret == "" || len(c.JWTSecret) < 32 {
		return fmt.Errorf("config: JWT_SECRET is required and must be >= 32 chars")
	}
	return nil
}

func (c *Config) MySQLDSN() string {
	return c.mysqlDSN(false)
}

// MySQLMigrationDSN enables multiStatements only for the isolated migration
// connection. The normal application pool deliberately keeps it disabled.
func (c *Config) MySQLMigrationDSN() string {
	return c.mysqlDSN(true)
}

func (c *Config) mysqlDSN(multiStatements bool) string {
	multi := ""
	if multiStatements {
		multi = "&multiStatements=true"
	}

	// parseTime=true so DATETIME/TIMESTAMP scan into time.Time.
	// multiStatements is only enabled for the migration connection; it remains
	// disabled for normal tenant-scoped application queries.
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=UTC%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, multi,
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
