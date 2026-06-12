package openspec

import (
	"os"
	"path/filepath"
	"testing"
)

// repoRoot walks up from the test directory to the module root (where go.mod is).
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("no go.mod found above %s", dir)
		}
		dir = parent
	}
}

func TestLoadRealChange(t *testing.T) {
	p := New(repoRoot(t))

	names, err := p.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if !contains(names, "specutil-core") {
		t.Fatalf("specutil-core not discovered; got %v", names)
	}

	c, err := p.Load("specutil-core")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if c.Proposal == nil {
		t.Fatal("proposal not loaded")
	}
	if len(c.Proposal.Capabilities.New) != 8 {
		t.Errorf("expected 8 new capabilities, got %d", len(c.Proposal.Capabilities.New))
	}
	if c.Design == nil || c.Design.Context == "" {
		t.Error("design context not loaded")
	}
	if c.Tasks == nil || len(c.Tasks.Phases) == 0 {
		t.Error("tasks not loaded")
	}
	if len(c.Specs) != 8 {
		t.Errorf("expected 8 specs, got %d", len(c.Specs))
	}

	// Every spec should have at least one requirement with at least one scenario.
	for _, s := range c.Specs {
		if len(s.Requirements) == 0 {
			t.Errorf("spec %q has no requirements", s.Capability)
		}
		for _, r := range s.Requirements {
			if len(r.Scenarios) == 0 {
				t.Errorf("spec %q requirement %q has no scenarios", s.Capability, r.Name)
			}
		}
	}

	// The real, well-formed change should parse without warnings.
	if len(c.Warnings) != 0 {
		t.Errorf("expected no warnings on the canonical change, got %d:", len(c.Warnings))
		for _, w := range c.Warnings {
			t.Logf("  %s:%d %s", w.File, w.Line, w.Msg)
		}
	}

	// Internal structure graph: change -> capability edges exist.
	edges := c.Edges()
	if len(edges) == 0 {
		t.Error("expected internal edges")
	}
}

func TestListEmptyRepo(t *testing.T) {
	p := New(t.TempDir())
	names, err := p.List()
	if err != nil {
		t.Fatalf("List on empty repo: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected no changes, got %v", names)
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
