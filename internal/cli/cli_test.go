package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/roshbhatia/specutil/internal/cli"
)

func examplesDir() string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(root, "examples", "getting-started")
}

func run(args ...string) (stdout, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer
	root := cli.NewRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs(args)
	err = root.Execute()
	return outBuf.String(), errBuf.String(), err
}

func TestRenderRFC(t *testing.T) {
	out, _, err := run("-C", examplesDir(), "render", "--as", "rfc", "--change", "add-auth-layer")
	if err != nil {
		t.Fatalf("render rfc: %v", err)
	}
	for _, want := range []string{"add-auth-layer", "JWT", "middleware"} {
		if !strings.Contains(out, want) {
			t.Errorf("render rfc output missing %q", want)
		}
	}
}

func TestRenderDesign(t *testing.T) {
	out, _, err := run("-C", examplesDir(), "render", "--as", "design", "--change", "add-auth-layer")
	if err != nil {
		t.Fatalf("render design: %v", err)
	}
	if !strings.Contains(out, "add-auth-layer") {
		t.Error("render design output missing change name")
	}
}

func TestRenderTickets(t *testing.T) {
	out, _, err := run("-C", examplesDir(), "render", "--as", "tickets", "--change", "add-auth-layer")
	if err != nil {
		t.Fatalf("render tickets: %v", err)
	}
	if !strings.Contains(out, "add-auth-layer") {
		t.Error("render tickets output missing change name")
	}
}

func TestRenderMissingAs(t *testing.T) {
	_, _, err := run("-C", examplesDir(), "render", "--change", "add-auth-layer")
	if err == nil {
		t.Error("expected error when --as is missing")
	}
}

func TestRenderUnknownTarget(t *testing.T) {
	_, _, err := run("-C", examplesDir(), "render", "--as", "nonexistent", "--change", "add-auth-layer")
	if err == nil {
		t.Error("expected error for unknown render target")
	}
}

func TestPlanLinear(t *testing.T) {
	out, _, err := run("-C", examplesDir(), "plan", "--target", "linear", "--change", "add-auth-layer")
	if err != nil {
		t.Fatalf("plan linear: %v", err)
	}
	if !strings.Contains(out, "add-auth-layer") {
		t.Error("plan output missing change name")
	}
}

func TestPlanGitHubIssues(t *testing.T) {
	out, _, err := run("-C", examplesDir(), "plan", "--target", "github-issues", "--change", "add-auth-layer")
	if err != nil {
		t.Fatalf("plan github-issues: %v", err)
	}
	if !strings.Contains(out, "add-auth-layer") {
		t.Error("plan github-issues output missing change name")
	}
}

func TestGraphMermaid(t *testing.T) {
	out, _, err := run("-C", examplesDir(), "graph", "--as", "mermaid")
	if err != nil {
		t.Fatalf("graph mermaid: %v", err)
	}
	if !strings.Contains(out, "graph") {
		t.Error("graph mermaid output missing graph keyword")
	}
}

func TestGraphDot(t *testing.T) {
	out, _, err := run("-C", examplesDir(), "graph", "--as", "dot")
	if err != nil {
		t.Fatalf("graph dot: %v", err)
	}
	if !strings.Contains(out, "digraph") {
		t.Error("graph dot output missing digraph keyword")
	}
}

func TestDiffNoLock(t *testing.T) {
	_, _, err := run("-C", examplesDir(), "diff", "--target", "linear", "--change", "add-auth-layer")
	if err != nil {
		t.Fatalf("diff with no lock: %v", err)
	}
}

func TestLockSetGet(t *testing.T) {
	dir := setupMinimalOpenspec(t, "test-change")
	_, _, err := run("-C", dir, "lock", "set", "test-id", "ext-123", "--target", "linear", "--change", "test-change")
	if err != nil {
		t.Fatalf("lock set: %v", err)
	}

	out, _, err := run("-C", dir, "lock", "get", "test-id", "--target", "linear", "--change", "test-change")
	if err != nil {
		t.Fatalf("lock get: %v", err)
	}
	if !strings.Contains(out, "ext-123") {
		t.Errorf("lock get output missing ext-123, got: %q", out)
	}
}

func TestRenderBMAD(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "..", "..")
	dir := filepath.Join(root, "examples", "bmad-project")

	out, _, err := run("-C", dir, "--from", "bmad", "render", "--as", "rfc", "--change", "story-1.1")
	if err != nil {
		t.Fatalf("render bmad rfc: %v", err)
	}
	if !strings.Contains(out, "story-1.1") {
		t.Error("render bmad output missing change name")
	}
}

func TestRenderPlanMd(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "..", "..")
	dir := filepath.Join(root, "examples", "plan-md")

	// Use -C to set the repo root so the plan provider auto-discovers plan.md.
	out, _, err := run("-C", dir, "--from", "plan", "render", "--as", "rfc")
	if err != nil {
		t.Fatalf("render plan.md rfc: %v", err)
	}
	if len(out) == 0 {
		t.Error("render plan.md produced empty output")
	}
}

// setupMinimalOpenspec creates a temp dir with a minimal openspec change and returns the root.
func setupMinimalOpenspec(t *testing.T, changeName string) string {
	t.Helper()
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "openspec", "changes", changeName)
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", changeDir, err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "proposal.md"),
		[]byte("## Why\n\nTest change.\n\n## What Changes\n\n- Something.\n"), 0o644); err != nil {
		t.Fatalf("write proposal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "tasks.md"),
		[]byte("## 1. Build\n\n- [ ] 1.1 Do the thing\n"), 0o644); err != nil {
		t.Fatalf("write tasks: %v", err)
	}
	return dir
}
