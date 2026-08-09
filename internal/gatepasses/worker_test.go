package gatepasses

import (
	"testing"
	"time"
)

func TestWorkerDefaults(t *testing.T) {
	w := NewWorker(nil, 0, 0, nil)
	if w.interval <= 0 {
		t.Fatal("worker interval must have a positive default")
	}
	if w.approvedTTL <= 0 {
		t.Fatal("approved TTL must have a positive default")
	}
}

func TestApprovedTTLIsIndependentFromRunInterval(t *testing.T) {
	w := NewWorker(nil, 5*time.Minute, 72*time.Hour, nil)
	if w.interval != 5*time.Minute || w.approvedTTL != 72*time.Hour {
		t.Fatalf("unexpected worker config: interval=%v ttl=%v", w.interval, w.approvedTTL)
	}
}
