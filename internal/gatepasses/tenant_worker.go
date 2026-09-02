package gatepasses

import (
	"context"
	"log"
	"sync"
	"time"

	"gatepass/internal/tenantdb"
	"gatepass/internal/tenants"
)

// TenantWorker coordinates maintenance across all READY tenant databases.
// The platform DB is only used to discover tenant IDs; operational SQL runs
// against each tenant's own database pool.
type TenantWorker struct {
	tenants     *tenants.Repository
	manager     *tenantdb.Manager
	interval    time.Duration
	approvedTTL time.Duration
	batchSize   int
	logger      *log.Logger
}

func NewTenantWorker(repo *tenants.Repository, manager *tenantdb.Manager, interval, approvedTTL time.Duration, batchSize int, logger *log.Logger) *TenantWorker {
	if interval <= 0 {
		interval = time.Minute
	}
	if approvedTTL <= 0 {
		approvedTTL = 24 * time.Hour
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	if logger == nil {
		logger = log.Default()
	}
	return &TenantWorker{tenants: repo, manager: manager, interval: interval, approvedTTL: approvedTTL, batchSize: batchSize, logger: logger}
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
	ids, err := w.tenants.ReadyIDs(ctx)
	if err != nil {
		w.logger.Printf("tenant worker: list ready tenants: %v", err)
		return
	}
	var wg sync.WaitGroup
	for _, tenantID := range ids {
		tenantID := tenantID
		wg.Add(1)
		go func() {
			defer wg.Done()
			db, err := w.manager.DB(ctx, tenantID)
			if err != nil {
				w.logger.Printf("tenant worker: tenant %d database unavailable: %v", tenantID, err)
				return
			}
			worker := NewWorker(db, w.approvedTTL, w.logger)
			worker.SetBatchSize(w.batchSize)
			if err := worker.RunOnce(ctx); err != nil {
				w.logger.Printf("tenant worker: tenant %d maintenance failed: %v", tenantID, err)
			}
		}()
	}
	wg.Wait()
}
