package registry_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/roshbhatia/specutil/internal/registry"
)

func TestSelectsOpenSpec(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}

	p, err := registry.SelectProvider(dir)
	if err != nil {
		t.Fatalf("SelectProvider: %v", err)
	}
	if p.Name() != "openspec" {
		t.Errorf("Name() = %q, want openspec", p.Name())
	}
}

func TestRejectsRepositoryWithoutOpenSpec(t *testing.T) {
	_, err := registry.SelectProvider(t.TempDir())
	if err == nil {
		t.Fatal("expected an error without openspec/changes")
	}
}
