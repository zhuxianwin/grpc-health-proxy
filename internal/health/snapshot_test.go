package health

import (
	"testing"
	"time"
)

func TestSnapshot_RecordAndLatest(t *testing.T) {
	store := NewSnapshotStore(nil)
	r := Result{Status: StatusHealthy}
	store.Record("svc", r)

	e, ok := store.Latest("svc")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Service != "svc" {
		t.Errorf("got service %q, want svc", e.Service)
	}
	if e.Result.Status != StatusHealthy {
		t.Errorf("got status %v, want healthy", e.Result.Status)
	}
	if e.RecordedAt.IsZero() {
		t.Error("RecordedAt should not be zero")
	}
}

func TestSnapshot_MissingService(t *testing.T) {
	store := NewSnapshotStore(nil)
	_, ok := store.Latest("missing")
	if ok {
		t.Error("expected no entry for unknown service")
	}
}

func TestSnapshot_OverwritesOldEntry(t *testing.T) {
	store := NewSnapshotStore(nil)
	store.Record("svc", Result{Status: StatusHealthy})
	time.Sleep(time.Millisecond)
	store.Record("svc", Result{Status: StatusUnhealthy})

	e, _ := store.Latest("svc")
	if e.Result.Status != StatusUnhealthy {
		t.Errorf("expected unhealthy after overwrite, got %v", e.Result.Status)
	}
}

func TestSnapshot_All(t *testing.T) {
	store := NewSnapshotStore(nil)
	store.Record("a", Result{Status: StatusHealthy})
	store.Record("b", Result{Status: StatusUnhealthy})

	all := store.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestSnapshot_AllEmpty(t *testing.T) {
	store := NewSnapshotStore(nil)
	if got := store.All(); len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

func TestDefaultSnapshotConfig(t *testing.T) {
	cfg := DefaultSnapshotConfig()
	if cfg.Interval != 30*time.Second {
		t.Errorf("expected 30s interval, got %v", cfg.Interval)
	}
}
