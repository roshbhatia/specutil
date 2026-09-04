package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProviderListDiscoversDataDirectoryManifest(t *testing.T) {
	data := t.TempDir()
	directory := filepath.Join(data, "specutil", "providers")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `version: provider/v1
name: test
description: Test provider
command: [test-provider]
actions:
  graph.suggest:
    description: Suggest graph edges
`
	if err := os.WriteFile(filepath.Join(directory, "test.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SPECUTIL_PROVIDERS_DIRECTORY", "")
	t.Setenv("XDG_DATA_HOME", data)
	t.Setenv("XDG_DATA_DIRS", t.TempDir())
	command := NewRootCmd()
	command.SetArgs([]string{"provider", "list"})
	var output strings.Builder
	command.SetOut(&output)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "test\tTest provider") {
		t.Fatalf("provider list = %q", output.String())
	}
}

func TestProviderValidateRejectsMissingRuntimeCommand(t *testing.T) {
	directory := t.TempDir()
	manifest := `version: provider/v1
name: missing
description: Missing provider
command: [definitely-missing-specutil-provider]
actions:
  graph.suggest:
    description: Suggest graph edges
`
	if err := os.WriteFile(filepath.Join(directory, "missing.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SPECUTIL_PROVIDERS_DIRECTORY", directory)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_DATA_DIRS", t.TempDir())
	command := NewRootCmd()
	command.SetArgs([]string{"provider", "validate", "missing"})
	if err := command.Execute(); err == nil {
		t.Fatal("provider validation accepted a missing runtime command")
	}
}
