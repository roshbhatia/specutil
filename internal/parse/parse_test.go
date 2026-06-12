package parse

import (
	"strings"
	"testing"

	"github.com/roshbhatia/specutil/internal/ir"
)

const sampleProposal = `## Why

We need a thing.

## What Changes

- Introduce the thing.

### Non-goals

- Not the other thing.

## Capabilities

### New Capabilities
- ` + "`cli-foundation`" + `: the cobra root and build tooling.
- ` + "`spec-ingestion`" + `: the provider port and IR.

### Modified Capabilities
<!-- None. -->

## Impact

- New code everywhere.
`

func TestParseProposal(t *testing.T) {
	p, warns := ParseProposal("proposal.md", sampleProposal)
	if len(warns) != 0 {
		t.Fatalf("unexpected warnings: %v", warns)
	}
	if !strings.Contains(p.Why, "need a thing") {
		t.Errorf("Why = %q", p.Why)
	}
	if !strings.Contains(p.WhatChanges, "Introduce the thing") {
		t.Errorf("WhatChanges = %q", p.WhatChanges)
	}
	if !strings.Contains(p.NonGoals, "other thing") {
		t.Errorf("NonGoals = %q", p.NonGoals)
	}
	if len(p.Capabilities.New) != 2 {
		t.Fatalf("expected 2 new capabilities, got %d: %+v", len(p.Capabilities.New), p.Capabilities.New)
	}
	if p.Capabilities.New[0].Name != "cli-foundation" {
		t.Errorf("cap[0].Name = %q", p.Capabilities.New[0].Name)
	}
	if !strings.Contains(p.Capabilities.New[0].Description, "cobra root") {
		t.Errorf("cap[0].Description = %q", p.Capabilities.New[0].Description)
	}
	if len(p.Capabilities.Modified) != 0 {
		t.Errorf("expected 0 modified capabilities, got %+v", p.Capabilities.Modified)
	}
	// Hybrid IR: raw is retained verbatim.
	if p.Raw != sampleProposal {
		t.Errorf("Raw not retained verbatim")
	}
}

const sampleSpec = `## ADDED Requirements

### Requirement: Does a thing
The system SHALL do a thing.

#### Scenario: It does the thing
- **WHEN** invoked
- **THEN** the thing is done

#### Scenario: It rejects bad input
- **WHEN** invoked with garbage
- **THEN** it errors

## MODIFIED Requirements

### Requirement: Changed behavior
It SHALL behave differently now.

#### Scenario: New behavior
- **WHEN** triggered
- **THEN** new behavior
`

func TestParseSpec(t *testing.T) {
	spec, warns := ParseSpec("specs/x/spec.md", "x", sampleSpec)
	if len(warns) != 0 {
		t.Fatalf("unexpected warnings: %v", warns)
	}
	if spec.Capability != "x" {
		t.Errorf("Capability = %q", spec.Capability)
	}
	if len(spec.Requirements) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(spec.Requirements))
	}
	r0 := spec.Requirements[0]
	if r0.Delta != ir.DeltaAdded {
		t.Errorf("r0.Delta = %q, want ADDED", r0.Delta)
	}
	if r0.Name != "Does a thing" {
		t.Errorf("r0.Name = %q", r0.Name)
	}
	if !strings.Contains(r0.Text, "SHALL do a thing") {
		t.Errorf("r0.Text = %q", r0.Text)
	}
	if len(r0.Scenarios) != 2 {
		t.Fatalf("expected 2 scenarios, got %d", len(r0.Scenarios))
	}
	if r0.Scenarios[0].Name != "It does the thing" {
		t.Errorf("scenario name = %q", r0.Scenarios[0].Name)
	}
	if len(r0.Scenarios[0].Steps) != 2 {
		t.Errorf("expected 2 steps, got %v", r0.Scenarios[0].Steps)
	}
	if !strings.Contains(r0.Scenarios[0].Steps[0], "WHEN") {
		t.Errorf("step[0] = %q", r0.Scenarios[0].Steps[0])
	}
	if spec.Requirements[1].Delta != ir.DeltaModified {
		t.Errorf("r1.Delta = %q, want MODIFIED", spec.Requirements[1].Delta)
	}
}

func TestParseSpecWarnsOnMisnestedScenario(t *testing.T) {
	// Scenario authored at level 3 instead of 4 — must be recovered with a warning.
	misnested := `## ADDED Requirements

### Requirement: A req
Text.

### Scenario: Wrong depth
- **WHEN** x
- **THEN** y
`
	spec, warns := ParseSpec("specs/x/spec.md", "x", misnested)
	if len(spec.Requirements) != 1 {
		t.Fatalf("expected 1 requirement (stray scenario should not become one), got %d", len(spec.Requirements))
	}
	// Lenient recovery: the misnested level-3 scenario is reattached to the
	// preceding requirement, with a loud wrong-depth warning.
	if len(spec.Requirements[0].Scenarios) != 1 {
		t.Errorf("expected the stray scenario reattached, got %d scenarios", len(spec.Requirements[0].Scenarios))
	}
	foundDepthWarning := false
	for _, w := range warns {
		if strings.Contains(w.Msg, "expected 4") && strings.Contains(w.Msg, "attaching") {
			foundDepthWarning = true
		}
	}
	if !foundDepthWarning {
		t.Errorf("expected a wrong-depth recovery warning, got %v", warns)
	}
}

const sampleTasks = `## 1. Foundation

- [x] 1.1 Initialize the module
- [ ] 1.2 Verify: build succeeds
- [ ] 1.3 Confirm: tests are green

## 2. Rollout

- [ ] 2.1 Apply: push the branch
`

func TestParseTasks(t *testing.T) {
	tasks, warns := ParseTasks("tasks.md", sampleTasks)
	if len(warns) != 0 {
		t.Fatalf("unexpected warnings: %v", warns)
	}
	if len(tasks.Phases) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(tasks.Phases))
	}
	p0 := tasks.Phases[0]
	if p0.Number != "1" || p0.Name != "Foundation" {
		t.Errorf("phase0 = %q/%q", p0.Number, p0.Name)
	}
	if len(p0.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(p0.Items))
	}
	if !p0.Items[0].Done {
		t.Errorf("item 1.1 should be done")
	}
	if p0.Items[0].ID != "1.1" {
		t.Errorf("item0 ID = %q", p0.Items[0].ID)
	}
	if p0.Items[1].Kind != ir.KindVerify {
		t.Errorf("item 1.2 kind = %q, want verify", p0.Items[1].Kind)
	}
	if p0.Items[2].Kind != ir.KindConfirm {
		t.Errorf("item 1.3 kind = %q, want confirm", p0.Items[2].Kind)
	}
	if tasks.Phases[1].Items[0].Kind != ir.KindApply {
		t.Errorf("item 2.1 kind = %q, want apply", tasks.Phases[1].Items[0].Kind)
	}
}

func TestSplitSectionsIgnoresFencedHeadings(t *testing.T) {
	src := "## Real\n\ntext\n\n" + "```\n## Not a heading\n```\n\n## AlsoReal\n"
	_, roots := SplitSections(src)
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots (fenced heading ignored), got %d: %+v", len(roots), titles(roots))
	}
	if roots[0].Title != "Real" || roots[1].Title != "AlsoReal" {
		t.Errorf("titles = %v", titles(roots))
	}
}

func titles(ns []*Node) []string {
	var out []string
	for _, n := range ns {
		out = append(out, n.Title)
	}
	return out
}
