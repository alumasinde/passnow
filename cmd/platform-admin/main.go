package main

import (
    "context"
    "flag"
    "log"
    "strings"
    "time"

    "gatepass/internal/config"
    "gatepass/internal/database"
    "gatepass/internal/platform"
    "gatepass/internal/users"
)

func main() {
    email := flag.String("email", "", "existing user email to grant platform access")
    role := flag.String("role", "owner", "platform role: owner or admin")
    flag.Parse()

    if *email == "" {
        log.Fatal("email is required")
    }
    if *role != "owner" && *role != "admin" {
        log.Fatal("role must be owner or admin")
    }

    cfg, err := config.Load()
    if err != nil { log.Fatalf("config: %v", err) }

    db, err := database.Connect(cfg)
    if err != nil { log.Fatalf("database: %v", err) }
    defer db.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    user, err := users.NewRepository(db).ByEmail(ctx, strings.TrimSpace(strings.ToLower(*email)))
    if err != nil {
        log.Fatalf("user not found: %v", err)
    }

    if err := platform.NewAdminRepository(db).Grant(ctx, user.ID, *role); err != nil {
        log.Fatalf("grant platform access: %v", err)
    }

    log.Printf("platform %s access granted to %s", *role, user.Email)
}
