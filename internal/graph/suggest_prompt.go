package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/roshbhatia/specutil/internal/ir"
)

var suggestPromptTemplate = template.Must(template.New("suggest-prompt").Parse(`You are analyzing a set of software changes to suggest dependency relationships.

A dependency edge "A depends on B" means: change A cannot be started until change B is complete.
Common signals: A uses types/APIs that B introduces; A modifies something B creates; A's proposal mentions B's capability.

Output ONLY a JSON object with this exact shape (no prose, no markdown fences):
{"suggestions": [{"from": "prereq-change", "to": "dependent-change", "reason": "one sentence"}]}

If you find no dependencies, output: {"suggestions": []}

Changes to analyze:
{{range .}}
--- {{.Name}} ---
{{with .Proposal}}{{if .Why}}Why: {{.Why}}
{{end}}{{if .WhatChanges}}What changes: {{.WhatChanges}}
{{end}}{{range .Capabilities.New}}Adds capability: {{.Name}}
{{end}}{{range .Capabilities.Modified}}Modifies capability: {{.Name}}
{{end}}{{end}}{{end}}`))

func buildSuggestPrompt(changes []*ir.Change) (string, error) {
	var output strings.Builder
	if err := suggestPromptTemplate.Execute(&output, changes); err != nil {
		return "", fmt.Errorf("render suggestion prompt: %w", err)
	}
	return output.String(), nil
}

type suggestion struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

type suggestionOutput struct {
	Suggestions []suggestion `json:"suggestions"`
}

func parseSuggestionOutput(raw []byte, known map[string]bool) ([]Candidate, error) {
	trimmed := bytes.TrimSpace(raw)
	if bytes.HasPrefix(trimmed, []byte("```")) {
		lines := bytes.Split(trimmed, []byte("\n"))
		if len(lines) > 2 {
			trimmed = bytes.Join(lines[1:len(lines)-1], []byte("\n"))
		}
	}
	var response suggestionOutput
	if err := json.Unmarshal(trimmed, &response); err != nil {
		return nil, fmt.Errorf("parse suggestion provider JSON: %w\nraw output: %s", err, string(raw))
	}
	var candidates []Candidate
	for _, one := range response.Suggestions {
		if !known[one.From] || !known[one.To] || one.From == one.To {
			continue
		}
		candidates = append(candidates, Candidate{From: one.From, To: one.To, Capability: one.Reason})
	}
	return candidates, nil
}
