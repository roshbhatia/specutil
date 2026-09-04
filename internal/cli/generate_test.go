package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/roshbhatia/go-utils/completion"
)

func TestCompletionGeneratorSupportsPublishedShells(t *testing.T) {
	metadata := completionMetadata(NewRootCmd())
	for _, shell := range []string{"bash", "zsh", "fish", "nu"} {
		t.Run(shell, func(t *testing.T) {
			generated, err := completion.Generate(shell, metadata)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(generated, "specutil") {
				t.Fatalf("completion omits command name:\n%s", generated)
			}
			if !strings.Contains(generated, "__values") {
				t.Fatalf("completion omits dynamic change candidates:\n%s", generated)
			}
		})
	}
}

func TestValuesCompletesChangesInSelectedRepository(t *testing.T) {
	repository := filepath.Join(t.TempDir(), "repository with spaces")
	change := filepath.Join(repository, "openspec", "changes", "rotate-signing-keys")
	if err := os.MkdirAll(change, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(change, "proposal.md"), []byte("## Why\n\nRotate keys.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(change, "tasks.md"), []byte("## 1. Rotate\n\n- [ ] 1.1 Replace key\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	command := NewRootCmd()
	command.SetArgs([]string{"__values", "changes", `specutil --repo "` + repository + `" render `})
	var output strings.Builder
	command.SetOut(&output)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(output.String()); got != "rotate-signing-keys" {
		t.Fatalf("completion = %q", got)
	}
}

func TestGenerateCheckDetectsStaleArtifacts(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "README.md"), []byte(`# specutil

<!-- BEGIN GENERATED:cli -->
stale
<!-- END GENERATED:cli -->
`), 0o644); err != nil {
		t.Fatal(err)
	}
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	if err := generateArtifacts(NewRootCmd(), true); err == nil {
		t.Fatal("check accepted stale generated artifacts")
	}
}

func TestProjectSchemaUsesYAMLFieldNamesAndAllowsInlineRuleParameters(t *testing.T) {
	raw, err := projectConfigSchema()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"Changes"`) || !strings.Contains(string(raw), `"depends_on"`) {
		t.Fatalf("schema does not match project YAML names:\n%s", raw)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	definitions, _ := document["$defs"].(map[string]any)
	foundRule := false
	for _, definition := range definitions {
		object, _ := definition.(map[string]any)
		properties, _ := object["properties"].(map[string]any)
		if _, hasID := properties["id"]; !hasID {
			continue
		}
		if _, hasSeverity := properties["severity"]; !hasSeverity {
			continue
		}
		foundRule = true
		if object["additionalProperties"] != true {
			t.Fatal("check rules must allow inline rule parameters used by the YAML loader")
		}
	}
	if !foundRule {
		t.Fatal("schema omits check rules")
	}
}

func TestConfigSchemaWorksOutsideTheSourceRepository(t *testing.T) {
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	command := NewRootCmd()
	command.SetArgs([]string{"config", "schema"})
	var output strings.Builder
	command.SetOut(&output)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !json.Valid([]byte(output.String())) {
		t.Fatalf("config schema output is not JSON: %s", output.String())
	}
}
