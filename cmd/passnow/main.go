package main

import (
	"errors"
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
		err = migrations.Up()
	case "migrate-status":
		err = migrations.Status()
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "passnow: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: go run ./cmd/passnow <serve|migrate|migrate-status>")
}

var _ = errors.New
