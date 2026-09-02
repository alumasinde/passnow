package gatepasses

import (
	"context"
	"log"
	"time"

	"gatepass/internal/tenantdb"
	"gatepass/internal/tenants"
)

// TenantWorker runs operational maintenance against every active tenant's
// isolated database. The platform database is used only to enumerate tenants;
// tenant tables are never queried through the platform connection.
type TenantWorker struct {
	tenants     *tenants.Repository
	manager     *tenantdb.Manager
	interval    time.Duration
	approvedTTL time.Duration
	logger      *log.Logger
}

func NewTenantWorker(repo *tenants.Repository, manager *tenantdb.Manager, interval, approvedTTL time.Duration, logger *log.Logger) *TenantWorker {
	if interval <= 0 {
		interval = time.Minute
	}
	if approvedTTL <= 0 {
		approvedTTL = 24 * time.Hour
	}
	if logger == nil {
		logger = log.Default()
	}
	return &TenantWorker{tenants: repo, manager: manager, interval: interval, approvedTTL: approvedTTL, logger: logger}
}

func (w *TenantWorker) Run(ctx context.Context) {
	w.runOnce(ctx)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *TenantWorker) runOnce(ctx context.Context) {
	items, err := w.tenants.List(ctx)
	if err != nil {
		w.logger.Printf("tenant worker: list tenants: %v", err)
		return
	}

	for _, tenant := range items {
		if !tenant.IsActive() {
			continue
		}
		db, err := w.manager.DB(ctx, tenant.ID)
		if err != nil {
			// A tenant can legitimately be active while provisioning/repair is
			// still incomplete. Skip that tenant without stopping the others.
			w.logger.Printf("tenant worker: tenant %d database unavailable: %v", tenant.ID, err)
			continue
		}
		worker := NewWorker(db, w.interval, w.approvedTTL, w.logger)
		worker.runOnce(ctx)
	}
}
