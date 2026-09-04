package graph

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	providerlib "github.com/roshbhatia/go-utils/provider"
	"github.com/roshbhatia/specutil/internal/ir"
)

const suggestCapability = "graph.suggest"

type suggestProviderInput struct {
	Command string `json:"command,omitempty"`
	Prompt  string `json:"prompt"`
}

func ProviderSuggest(
	ctx context.Context,
	changes []*ir.Change,
	providerName string,
	command string,
	workingDirectory string,
) ([]Candidate, error) {
	if providerName == "" {
		return nil, errors.New("suggestion provider name is required")
	}
	manifest, err := findSuggestionProvider(providerName)
	if err != nil {
		return nil, err
	}
	if _, ok := manifest.Actions[suggestCapability]; !ok {
		return nil, fmt.Errorf("provider %q does not advertise %s", providerName, suggestCapability)
	}
	prompt, err := buildSuggestPrompt(changes)
	if err != nil {
		return nil, err
	}
	input, err := json.Marshal(suggestProviderInput{Command: command, Prompt: prompt})
	if err != nil {
		return nil, err
	}
	requestID, err := randomRequestID()
	if err != nil {
		return nil, err
	}
	invocation, err := providerlib.Invoke(ctx, manifest, providerlib.Request{
		RequestID:  requestID,
		Capability: suggestCapability,
		Input:      input,
	}, providerlib.InvokeOptions{WorkingDirectory: workingDirectory, Environment: os.Environ()})
	if err != nil {
		return nil, fmt.Errorf("suggestion provider %q: %w", providerName, err)
	}
	if invocation.Result.Status != providerlib.ResultOK {
		return nil, fmt.Errorf("suggestion provider %q: %s", providerName, invocation.Result.Message)
	}
	known := make(map[string]bool, len(changes))
	for _, change := range changes {
		known[change.Name] = true
	}
	return parseSuggestionOutput(invocation.Result.Output, known)
}

func SuggestionProviderNames() []string {
	seen := map[string]bool{}
	var names []string
	for _, directory := range suggestionProviderDirectories() {
		manifests, err := providerlib.Discover(directory)
		if err != nil {
			continue
		}
		for _, loaded := range manifests {
			if seen[loaded.Manifest.Name] {
				continue
			}
			if _, ok := loaded.Manifest.Actions[suggestCapability]; !ok {
				continue
			}
			seen[loaded.Manifest.Name] = true
			names = append(names, loaded.Manifest.Name)
		}
	}
	return names
}

func SuggestionProviders() ([]providerlib.LoadedManifest, error) {
	seen := map[string]bool{}
	var providers []providerlib.LoadedManifest
	for _, directory := range suggestionProviderDirectories() {
		manifests, err := providerlib.Discover(directory)
		if err != nil {
			return nil, err
		}
		for _, loaded := range manifests {
			if seen[loaded.Manifest.Name] {
				continue
			}
			if _, ok := loaded.Manifest.Actions[suggestCapability]; !ok {
				continue
			}
			seen[loaded.Manifest.Name] = true
			providers = append(providers, loaded)
		}
	}
	return providers, nil
}

func findSuggestionProvider(name string) (providerlib.Manifest, error) {
	for _, directory := range suggestionProviderDirectories() {
		manifests, err := providerlib.Discover(directory)
		if err != nil {
			return providerlib.Manifest{}, err
		}
		for _, loaded := range manifests {
			if loaded.Manifest.Name == name {
				return loaded.Manifest, nil
			}
		}
	}
	return providerlib.Manifest{}, fmt.Errorf("suggestion provider %q was not found", name)
}

func suggestionProviderDirectories() []string {
	var directories []string
	if override := strings.TrimSpace(os.Getenv("SPECUTIL_PROVIDERS_DIRECTORY")); override != "" {
		directories = append(directories, filepath.Clean(override))
	}
	if home := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); home != "" {
		directories = append(directories, filepath.Join(home, "specutil", "providers"))
	} else if userHome, err := os.UserHomeDir(); err == nil {
		directories = append(directories, filepath.Join(userHome, ".local", "share", "specutil", "providers"))
	}
	for _, root := range filepath.SplitList(os.Getenv("XDG_DATA_DIRS")) {
		if root != "" {
			directories = append(directories, filepath.Join(root, "specutil", "providers"))
		}
	}
	return directories
}

func randomRequestID() (string, error) {
	value := make([]byte, 12)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("create provider request id: %w", err)
	}
	return hex.EncodeToString(value), nil
}
