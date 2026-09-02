package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"gatepass/internal/config"
	"gatepass/internal/database"
	"gatepass/internal/migrations"
	"gatepass/internal/tenantdb"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	action := flag.String("action", "up", "up or status")
	scope := flag.String("scope", "platform", "platform or tenant")
	dir := flag.String("dir", "", "migration directory override")
	host := flag.String("host", "", "tenant database host")
	port := flag.String("port", "3306", "tenant database port")
	name := flag.String("database", "", "tenant database name (manual credentials mode)")
	user := flag.String("user", "", "tenant database user (manual credentials mode)")
	tenantID := flag.Int64("tenant-id", 0, "platform tenant ID (loads encrypted tenant credentials)")
	password := flag.String("password", "", "tenant database password")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil { log.Fatalf("config: %v", err) }

	var db *sql.DB
	var migrationDir string
	var lockName string

	switch *scope {
	case "platform":
		db, err = database.Connect(cfg)
		migrationDir = "migrations/platform"
		lockName = migrations.LockPrefix + "_platform"
	case "tenant":
		if *tenantID > 0 {
			platformDB, connectErr := database.Connect(cfg)
			if connectErr != nil {
				log.Fatalf("platform database: %v", connectErr)
			}
			cipher, cipherErr := tenantdb.NewCipher(cfg.TenantDBEncryptionKey)
			if cipherErr != nil {
				_ = platformDB.Close()
				log.Fatalf("tenant credentials: %v", cipherErr)
			}
			manager := tenantdb.NewManager(
				tenantdb.NewRepository(platformDB),
				cipher,
				cfg.TenantDBMaxOpenConns,
				cfg.TenantDBMaxIdleConns,
				cfg.TenantDBConnMaxLifetime,
			)
			db, err = manager.DB(context.Background(), *tenantID)
			if err != nil {
				_ = manager.Close()
				_ = platformDB.Close()
				log.Fatalf("tenant database: %v", err)
			}
			defer manager.Close()
			defer platformDB.Close()
			if *name == "" {
				creds, credErr := manager.Credentials(context.Background(), *tenantID)
				if credErr != nil {
					log.Fatalf("tenant credentials: %v", credErr)
				}
				*name = creds.Database
			}
		} else {
			if *name == "" || *user == "" {
				log.Fatal("tenant scope requires -tenant-id or both -database and -user")
			}
			db, err = sql.Open("mysql", tenantDSN(tenantdb.Credentials{Host:*host,Port:*port,Database:*name,Username:*user,Password:*password}, cfg))
		}
		if *dir == "" { migrationDir = "migrations/tenant" } else { migrationDir = *dir }
		if *name == "" {
			log.Fatal("tenant migration lock requires a database name")
		}
		lockName = migrations.LockPrefix + "_tenant_" + *name
	default:
		log.Fatalf("unknown scope %q (use platform or tenant)", *scope)
	}
	if err != nil { log.Fatalf("database: %v", err) }
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if *dir != "" { migrationDir = *dir }

	switch *action {
	case "up":
		if err := migrations.RunUp(ctx, db, migrationDir, lockName); err != nil { log.Fatalf("migrate up: %v", err) }
		log.Printf("migrations complete (%s)", *scope)
	case "status":
		items, err := migrations.Status(ctx, db, migrationDir)
		if err != nil { log.Fatalf("migrate status: %v", err) }
		for _, item := range items { fmt.Println(item) }
	default:
		log.Fatalf("unknown action %q (use up or status)", *action)
	}
}

func tenantDSN(c tenantdb.Credentials, cfg *config.Config) string {
	host:=c.Host;if host==""{host=cfg.DBHost}
	port:=c.Port;if port==""{port=cfg.DBPort}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=UTC",c.Username,c.Password,host,port,c.Database)
}
