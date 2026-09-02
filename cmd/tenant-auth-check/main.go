package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"gatepass/internal/auth"
	"gatepass/internal/config"
	"gatepass/internal/database"
	"gatepass/internal/roles"
	"gatepass/internal/tenantdb"
	"gatepass/internal/tenants"
	"gatepass/internal/users"
)

func main() {
	tenantID := flag.Int64("tenant-id", 0, "platform tenant ID")
	email := flag.String("email", "", "tenant user email")
	password := flag.String("password", "", "password to verify")
	flag.Parse()

	if *tenantID <= 0 || strings.TrimSpace(*email) == "" || *password == "" {
		log.Fatal("tenant-id, email and password are required")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	platformDB, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("platform database: %v", err)
	}
	defer platformDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tenant, err := tenants.NewRepository(platformDB).ByID(ctx, *tenantID)
	if err != nil {
		log.Fatalf("FAIL tenant lookup: %v", err)
	}
	fmt.Printf("OK tenant: id=%d slug=%s status=%s\n", tenant.ID, tenant.Slug, tenant.Status)

	cipher, err := tenantdb.NewCipher(cfg.TenantDBEncryptionKey)
	if err != nil {
		log.Fatalf("FAIL tenant DB cipher: %v", err)
	}
	manager := tenantdb.NewManager(
		tenantdb.NewRepository(platformDB),
		cipher,
		cfg.TenantDBMaxOpenConns,
		cfg.TenantDBMaxIdleConns,
		cfg.TenantDBConnMaxLifetime,
	)
	defer manager.Close()

	tenantDB, err := manager.DB(ctx, tenant.ID)
	if err != nil {
		log.Fatalf("FAIL tenant database connection: %v", err)
	}
	fmt.Println("OK tenant database connection")

	normalizedEmail := strings.ToLower(strings.TrimSpace(*email))
	userRepo := users.NewRepository(tenantDB)
	u, err := userRepo.ByEmail(ctx, normalizedEmail)
	if err != nil {
		log.Fatalf("FAIL user lookup: %v", err)
	}
	fmt.Printf("OK user: id=%d email=%s status=%s failed_logins=%d\n", u.ID, u.Email, u.Status, u.FailedLoginCount)

	if !auth.VerifyPassword(u.PasswordHash, *password) {
		log.Fatal("FAIL password: supplied password does not match the bcrypt hash stored in this tenant database")
	}
	fmt.Println("OK password bcrypt verification")

	if u.IsLocked(time.Now().UTC()) {
		log.Fatalf("FAIL account locked until %s", u.LockedUntil.UTC().Format(time.RFC3339))
	}
	if !u.IsActive() {
		log.Fatalf("FAIL account status: %s", u.Status)
	}
	fmt.Println("OK account status")

	roleRepo := roles.NewRepository(tenantDB)
	membership, err := roleRepo.MembershipFor(ctx, u.ID)
	if err != nil {
		log.Fatalf("FAIL membership lookup: %v", err)
	}
	fmt.Printf("OK membership: id=%d user_id=%d role_id=%d status=%s\n", membership.ID, membership.UserID, membership.RoleID, membership.Status)

	if !membership.IsActive() {
		log.Fatalf("FAIL membership status: %s", membership.Status)
	}

	role, err := roleRepo.RoleByID(ctx, membership.RoleID)
	if err != nil {
		log.Fatalf("FAIL role lookup: %v", err)
	}
	fmt.Printf("AUTH CHECK PASSED: user can authenticate as role=%s (role_id=%d)\n", role.Name, role.ID)
}
