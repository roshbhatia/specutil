package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	providerlib "github.com/roshbhatia/go-utils/provider"
)

const capability = "graph.suggest"

type input struct {
	Command string `json:"command"`
	Prompt  string `json:"prompt"`
}

func main() {
	if err := run(os.Stdin, os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(stdin io.Reader, stdout, stderr io.Writer) error {
	var request providerlib.Request
	if err := json.NewDecoder(stdin).Decode(&request); err != nil {
		return fmt.Errorf("decode provider request: %w", err)
	}
	if request.Capability != capability {
		return writeResult(stdout, request.RequestID, providerlib.ResultDeclined, nil, "unsupported capability")
	}
	var payload input
	if err := json.Unmarshal(request.Input, &payload); err != nil {
		return writeResult(stdout, request.RequestID, providerlib.ResultError, nil, "invalid provider input")
	}
	if strings.TrimSpace(payload.Command) == "" {
		return writeResult(stdout, request.RequestID, providerlib.ResultError, nil, "command is required")
	}
	output, err := execute(payload.Command, payload.Prompt, stderr)
	if err != nil {
		return writeResult(stdout, request.RequestID, providerlib.ResultError, nil, err.Error())
	}
	output = stripFence(output)
	if !json.Valid(output) {
		return writeResult(stdout, request.RequestID, providerlib.ResultError, nil, "command returned invalid JSON")
	}
	return writeResult(stdout, request.RequestID, providerlib.ResultOK, output, "")
}

func execute(command, prompt string, stderr io.Writer) ([]byte, error) {
	providerCommand := exec.Command(command) //nolint:gosec // users select the provider command
	providerCommand.Stdin = strings.NewReader(prompt)
	providerCommand.Stderr = stderr
	output, err := providerCommand.Output()
	if err != nil {
		return nil, fmt.Errorf("run %s: %w", command, err)
	}
	return output, nil
}

func stripFence(output []byte) []byte {
	trimmed := bytes.TrimSpace(output)
	if !bytes.HasPrefix(trimmed, []byte("```")) {
		return trimmed
	}
	lines := bytes.Split(trimmed, []byte("\n"))
	if len(lines) < 3 || !bytes.Equal(bytes.TrimSpace(lines[len(lines)-1]), []byte("```")) {
		return trimmed
	}
	return bytes.TrimSpace(bytes.Join(lines[1:len(lines)-1], []byte("\n")))
}

func writeResult(
	output io.Writer,
	requestID string,
	status providerlib.ResultStatus,
	payload json.RawMessage,
	message string,
) error {
	return json.NewEncoder(output).Encode(providerlib.Result{
		Version:   providerlib.Version,
		Kind:      providerlib.FrameResult,
		RequestID: requestID,
		Status:    status,
		Output:    payload,
		Message:   message,
	})
}
