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
	name := flag.String("database", "", "tenant database name")
	user := flag.String("user", "", "tenant database user")
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
		if *name == "" || *user == "" {
			log.Fatal("tenant scope requires -database and -user")
		}
		db, err = sql.Open("mysql", tenantDSN(tenantdb.Credentials{Host:*host,Port:*port,Database:*name,Username:*user,Password:*password}, cfg))
		if *dir == "" { migrationDir = "migrations/tenant" } else { migrationDir = *dir }
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
