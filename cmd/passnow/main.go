package main

import (
	"fmt"
	"os"

	"gatepass/internal/app"
	"gatepass/internal/migrations"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "serve":
		app.Run()
		return
	case "migrate":
		err = runMigrations(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "passnow: %v\n", err)
		os.Exit(1)
	}
}

func runMigrations(args []string) error {
	if len(args) == 0 || args[0] == "up" {
		return migrations.Up()
	}

	switch args[0] {
	case "status", "version":
		return migrations.Status()
	default:
		return fmt.Errorf("unknown migration command %q (use: migrate, migrate up, or migrate status)", args[0])
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  go run ./cmd/passnow serve")
	fmt.Fprintln(os.Stderr, "  go run ./cmd/passnow migrate")
	fmt.Fprintln(os.Stderr, "  go run ./cmd/passnow migrate status")
}
