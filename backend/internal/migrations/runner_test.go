package migrations

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadAllowsLegacyDuplicatesButRejectsNewOnes(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "0008_a.up.sql")
	write(t, dir, "0008_b.up.sql") // legacy duplicate: allowed
	if _, err := Load(dir); err != nil {
		t.Fatalf("legacy duplicate should load: %v", err)
	}
	write(t, dir, "0024_x.up.sql")
	write(t, dir, "0024_y.up.sql")
	if _, err := Load(dir); err == nil {
		t.Fatal("expected duplicate version 0024 to be rejected")
	}
}

func TestLoadIgnoresDownFiles(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "0001_a.up.sql")
	write(t, dir, "0001_a.down.sql")
	items, err := Load(dir)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%d err=%v", len(items), err)
	}
}