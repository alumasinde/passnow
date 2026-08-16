package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gatepass/internal/config"
	"gatepass/internal/database"
	"gatepass/internal/gatepasses"
)

// Run is responsible only for process lifecycle: configuration, database,
// dependency composition, worker startup, HTTP serving and graceful shutdown.
// Endpoint definitions live in routes.go and dependency wiring in
// container.go.
func Run() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	container := NewContainer(cfg, db)
	server := NewServer(container)

	worker := gatepasses.NewWorker(db, cfg.GatepassWorkerInterval, cfg.ApprovedGatepassTTL, log.Default())
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go worker.Run(workerCtx)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("gatepass api listening on %s (env=%s)", cfg.HTTPAddr, cfg.Env)
		if err := server.ListenAndServe(); err != nil {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-stop:
		log.Printf("shutdown signal received: %s", sig)
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("server: %v", err)
		}
	}

	workerCancel()
	shutdownContext, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("shutdown complete")
}
