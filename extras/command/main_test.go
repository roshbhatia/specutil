package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	providerlib "github.com/roshbhatia/go-utils/provider"
)

func TestRunReturnsStructuredSuggestion(t *testing.T) {
	directory := t.TempDir()
	command := filepath.Join(directory, "agent")
	script := `#!/bin/sh
IFS= read -r prompt || true
[ "$prompt" = analyze ] || exit 1
printf '%s\n' '{"suggestions":[{"from":"a","to":"b","reason":"uses API"}]}'
`
	if err := os.WriteFile(command, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	input, _ := json.Marshal(input{Command: command, Prompt: "analyze"})
	request, _ := json.Marshal(providerlib.Request{
		Version:    providerlib.Version,
		Kind:       providerlib.FrameRequest,
		RequestID:  "test",
		Capability: capability,
		Input:      input,
	})
	var output bytes.Buffer
	if err := run(bytes.NewReader(request), &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var result providerlib.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != providerlib.ResultOK || !bytes.Contains(result.Output, []byte("uses API")) {
		t.Fatalf("unexpected result: %+v", result)
	}
}
