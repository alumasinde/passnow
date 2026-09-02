package gatepasses

import (
	"testing"
	"time"
)

func TestWorkerDefaults(t *testing.T) {
	w := NewWorker(nil, 0, nil)
	if w.approvedTTL <= 0 {
		t.Fatal("approved TTL must have a positive default")
	}
	if w.batch <= 0 {
		t.Fatal("worker batch must have a positive default")
	}
}

func TestWorkerUsesConfiguredApprovedTTL(t *testing.T) {
	w := NewWorker(nil, 72*time.Hour, nil)
	if w.approvedTTL != 72*time.Hour {
		t.Fatalf("unexpected approved TTL: %v", w.approvedTTL)
	}
}
