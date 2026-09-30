package main

import (
	"context"
	"flag"
	"log"
	"database/sql"
	"strings"
	"time"

	"gatepass/internal/auth"
	"gatepass/internal/database"
	"gatepass/internal/config"
	"gatepass/internal/tenantdb"
	"gatepass/internal/tenants"
	"gatepass/internal/users"
)

func main() {
	tenantID := flag.Int64("tenant-id", 0, "platform tenant ID")
	email := flag.String("email", "", "tenant user email")
	password := flag.String("password", "", "new password (minimum 12 characters)")
	flag.Parse()

	if *tenantID <= 0 {
		log.Fatal("tenant-id is required")
	}
	normalizedEmail := strings.ToLower(strings.TrimSpace(*email))
	if normalizedEmail == "" {
		log.Fatal("email is required")
	}
	if len(*password) < 12 {
		log.Fatal("password must be at least 12 characters")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.TenantDBEncryptionKey == "" {
		log.Fatal("tenant database encryption key is required")
	}

	platformDB, err := openPlatform(cfg)
	if err != nil {
		log.Fatalf("platform database: %v", err)
	}
	defer platformDB.Close()

	tenant, err := tenants.NewRepository(platformDB).ByID(context.Background(), *tenantID)
	if err != nil {
		log.Fatalf("tenant not found: %v", err)
	}

	cipher, err := tenantdb.NewCipher(cfg.TenantDBEncryptionKey)
	if err != nil {
		log.Fatalf("tenant database encryption: %v", err)
	}
	manager := tenantdb.NewManager(
		tenantdb.NewRepository(platformDB),
		cipher,
		cfg.TenantDBMaxOpenConns,
		cfg.TenantDBMaxIdleConns,
		cfg.TenantDBConnMaxLifetime,
	)
	defer manager.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tenantDB, err := manager.DB(ctx, tenant.ID)
	if err != nil {
		log.Fatalf("tenant database: %v", err)
	}

	repo := users.NewRepository(tenantDB)
	user, err := repo.ByEmail(ctx, normalizedEmail)
	if err != nil {
		log.Fatalf("tenant user not found: %v", err)
	}

	hash, err := auth.HashPassword(*password, cfg.BcryptCost)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}
	if err := repo.SetPasswordHash(ctx, user.ID, hash); err != nil {
		log.Fatalf("reset password: %v", err)
	}

	log.Printf("password reset for %s in tenant %s (id=%d)", user.Email, tenant.Slug, tenant.ID)
}

func openPlatform(cfg *config.Config) (*sql.DB, error) {
	return database.Connect(cfg)
}
